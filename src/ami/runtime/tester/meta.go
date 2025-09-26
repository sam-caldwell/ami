package tester

type metaInfo struct {
    sleepMs    int
    errorCode  string
    kvPipeline string
    kvNode     string
    kvPutKey   string
    kvPutVal   any
    kvGetKey   string
    kvEmit     bool
}

