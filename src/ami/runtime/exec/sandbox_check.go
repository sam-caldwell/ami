package exec

func sandboxCheck(p SandboxPolicy, cap string) error {
    if !p.allow(cap) { return ErrSandboxDenied{Cap: cap} }
    return nil
}

