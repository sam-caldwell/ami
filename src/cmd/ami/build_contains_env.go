package main

func containsEnv(list []string, env string) bool {
    for _, e := range list {
        if e == env {
            return true
        }
    }
    return false
}

