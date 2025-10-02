package main

func key(name, ver string) string { if ver == "" { return name }; return name + "@" + ver }

