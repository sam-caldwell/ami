package main

func anyKvEmit(cases []runtimeCase) bool {
    for _, c := range cases { if c.Spec.KvEmit { return true } }
    return false
}

