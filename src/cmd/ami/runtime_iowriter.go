package main

type ioWriter interface{ Write([]byte) (int, error) }

