package kvstore

func applyOptions(opts []PutOption) putOptions {
    var o putOptions
    for _, fn := range opts { fn(&o) }
    return o
}

