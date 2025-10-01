//go:build darwin

#import <Foundation/Foundation.h>
#import <Metal/Metal.h>
#include <stdbool.h>
#include <stdlib.h>
#include <string.h>

static int gNextId = 1;
static NSMutableDictionary<NSNumber*, NSDictionary*> *gCtxTab;
static NSMutableDictionary<NSNumber*, NSDictionary*> *gLibTab;
static NSMutableDictionary<NSNumber*, NSDictionary*> *gPipeTab;
static NSMutableDictionary<NSNumber*, NSDictionary*> *gBufTab;

static void ensureTabs(void) {
    static dispatch_once_t once;
    dispatch_once(&once, ^{
        gCtxTab = [NSMutableDictionary new];
        gLibTab = [NSMutableDictionary new];
        gPipeTab = [NSMutableDictionary new];
        gBufTab = [NSMutableDictionary new];
    });
}

bool AmiMetalAvailable(void) {
    id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
    return dev != nil;
}

int AmiMetalDeviceCount(void) {
    if (@available(macOS 10.11, *)) {
        NSArray<id<MTLDevice>> *devs = MTLCopyAllDevices();
        return (int)[devs count];
    } else {
        id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
        return dev != nil ? 1 : 0;
    }
}

char* AmiMetalDeviceNameAt(int idx) {
    if (@available(macOS 10.11, *)) {
        NSArray<id<MTLDevice>> *devs = MTLCopyAllDevices();
        if (idx < 0 || idx >= (int)[devs count]) return NULL;
        id<MTLDevice> dev = [devs objectAtIndex:(NSUInteger)idx];
        NSString *name = [dev name];
        const char *utf8 = [name UTF8String];
        if (utf8 == NULL) return NULL;
        size_t n = strlen(utf8);
        char *out = (char*)malloc(n+1);
        if (!out) return NULL;
        memcpy(out, utf8, n+1);
        return out;
    } else {
        if (idx != 0) return NULL;
        id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
        if (dev == nil) return NULL;
        NSString *name = [dev name];
        const char *utf8 = [name UTF8String];
        if (utf8 == NULL) return NULL;
        size_t n = strlen(utf8);
        char *out = (char*)malloc(n+1);
        if (!out) return NULL;
        memcpy(out, utf8, n+1);
        return out;
    }
}

void AmiMetalFreeCString(char* p) {
    if (p != NULL) free(p);
}

static char* dupErr(NSError *err) {
    if (!err) return NULL;
    const char *s = [[err localizedDescription] UTF8String];
    if (!s) s = "unknown metal error";
    size_t n = strlen(s);
    char *out = (char*)malloc(n+1);
    if (!out) return NULL;
    memcpy(out, s, n+1);
    return out;
}

int AmiMetalContextCreate(int devIndex, char** err) {
    ensureTabs();
    if (@available(macOS 10.11, *)) {
        NSArray<id<MTLDevice>> *devs = MTLCopyAllDevices();
        if (devIndex < 0 || devIndex >= (int)[devs count]) return 0;
        id<MTLDevice> dev = [devs objectAtIndex:(NSUInteger)devIndex];
        id<MTLCommandQueue> q = [dev newCommandQueue];
        if (!q) return 0;
        int id = gNextId++;
        gCtxTab[@(id)] = @{ @"dev": dev, @"q": q };
        return id;
    } else {
        if (devIndex != 0) return 0;
        id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
        if (!dev) return 0;
        id<MTLCommandQueue> q = [dev newCommandQueue];
        if (!q) return 0;
        int id = gNextId++;
        gCtxTab[@(id)] = @{ @"dev": dev, @"q": q };
        return id;
    }
}

void AmiMetalContextDestroy(int ctxId) {
    ensureTabs();
    NSNumber *k = @(ctxId);
    NSDictionary *ctx = gCtxTab[k];
    if (ctx) {
        id q = ctx[@"q"];
        (void)q; // ARC releases when dictionary entry is removed
        [gCtxTab removeObjectForKey:k];
    }
}

int AmiMetalCompileLibrary(const char* src, char** err) {
    ensureTabs();
    id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
    if (!dev) return 0;
    NSString *code = [NSString stringWithUTF8String:src ? src : ""];
    NSError *e = nil;
    id<MTLLibrary> lib = [dev newLibraryWithSource:code options:nil error:&e];
    if (!lib) { if (err) *err = dupErr(e); return 0; }
    int id = gNextId++;
    gLibTab[@(id)] = @{ @"dev": dev, @"lib": lib };
    return id;
}

int AmiMetalCreatePipeline(int libId, const char* name, char** err) {
    ensureTabs();
    NSDictionary *libent = gLibTab[@(libId)];
    if (!libent) return 0;
    id<MTLDevice> dev = libent[@"dev"];
    id<MTLLibrary> lib = libent[@"lib"];
    NSString *fname = [NSString stringWithUTF8String:name ? name : ""];
    id<MTLFunction> fn = [lib newFunctionWithName:fname];
    if (!fn) { if (err) *err = strdup("function not found"); return 0; }
    NSError *e = nil;
    id<MTLComputePipelineState> ps = [dev newComputePipelineStateWithFunction:fn error:&e];
    if (!ps) { if (err) *err = dupErr(e); return 0; }
    int id = gNextId++;
    gPipeTab[@(id)] = @{ @"dev": dev, @"ps": ps };
    return id;
}

int AmiMetalAlloc(unsigned long n, char** err) {
    ensureTabs();
    id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
    if (!dev) return 0;
    id<MTLBuffer> b = [dev newBufferWithLength:n options:MTLResourceStorageModeShared];
    if (!b) return 0;
    int id = gNextId++;
    gBufTab[@(id)] = @{ @"dev": dev, @"buf": b, @"len": @(n) };
    return id;
}

void AmiMetalFreeBuffer(int bufId) {
    ensureTabs();
    [gBufTab removeObjectForKey:@(bufId)];
}

int AmiMetalCopyToDevice(int bufId, const void* src, unsigned long n, char** err) {
    ensureTabs();
    NSDictionary *ent = gBufTab[@(bufId)];
    if (!ent) return -1;
    id<MTLBuffer> b = ent[@"buf"];
    unsigned long len = [ent[@"len"] unsignedLongValue];
    if (n > len) return -2;
    if (n == 0) return 0;
    void *dst = [b contents];
    memcpy(dst, src, n);
    return 0;
}

int AmiMetalCopyFromDevice(int bufId, void* dst, unsigned long n, char** err) {
    ensureTabs();
    NSDictionary *ent = gBufTab[@(bufId)];
    if (!ent) return -1;
    id<MTLBuffer> b = ent[@"buf"];
    unsigned long len = [ent[@"len"] unsignedLongValue];
    if (n > len) return -2;
    if (n == 0) return 0;
    void *src = [b contents];
    memcpy(dst, src, n);
    return 0;
}

int AmiMetalDispatch(int ctxId, int pipeId,
                     unsigned int gx, unsigned int gy, unsigned int gz,
                     unsigned int tx, unsigned int ty, unsigned int tz,
                     const int* bufIds, int bufCount, char** err) {
    ensureTabs();
    NSDictionary *ctx = gCtxTab[@(ctxId)];
    NSDictionary *p = gPipeTab[@(pipeId)];
    if (!ctx || !p) return -1;
    id<MTLDevice> devCtx = ctx[@"dev"];
    id<MTLCommandQueue> q = ctx[@"q"];
    id<MTLDevice> devP = p[@"dev"];
    if (devCtx != devP) { if (err) *err = strdup("device mismatch"); return -2; }
    id<MTLComputePipelineState> ps = p[@"ps"];
    id<MTLCommandBuffer> cb = [q commandBuffer];
    id<MTLComputeCommandEncoder> enc = [cb computeCommandEncoder];
    [enc setComputePipelineState:ps];
    for (int i = 0; i < bufCount; i++) {
        NSDictionary *be = gBufTab[@(bufIds[i])];
        if (!be) { [enc endEncoding]; return -3; }
        id<MTLBuffer> b = be[@"buf"];
        [enc setBuffer:b offset:0 atIndex:(NSUInteger)i];
    }
    MTLSize grid = MTLSizeMake(gx, gy, gz);
    MTLSize tpg = MTLSizeMake(tx ? tx : 1, ty ? ty : 1, tz ? tz : 1);
    [enc dispatchThreads:grid threadsPerThreadgroup:tpg];
    [enc endEncoding];
    [cb commit];
    [cb waitUntilCompleted];
    return 0;
}
