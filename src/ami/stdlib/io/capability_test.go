package io

import "testing"

func TestCapability_FS_Denied(t *testing.T) {
    defer ResetPolicy()
    SetPolicy(Policy{AllowFS:false, AllowNet:true, AllowDevice:true})
    if _, err := Create("build/test_cap.txt"); err == nil || err != ErrCapabilityDenied {
        t.Fatalf("expected ErrCapabilityDenied on Create, got %v", err)
    }
    if _, err := Open("/etc/hosts"); err == nil || err != ErrCapabilityDenied {
        t.Fatalf("expected ErrCapabilityDenied on Open, got %v", err)
    }
    if _, err := Stat("/etc/hosts"); err == nil || err != ErrCapabilityDenied {
        t.Fatalf("expected ErrCapabilityDenied on Stat, got %v", err)
    }
}

func TestCapability_Net_Denied(t *testing.T) {
    defer ResetPolicy()
    SetPolicy(Policy{AllowFS:true, AllowNet:false, AllowDevice:true})
    if _, err := OpenSocket(UDP, "127.0.0.1", 0); err == nil || err != ErrCapabilityDenied {
        t.Fatalf("expected ErrCapabilityDenied on OpenSocket, got %v", err)
    }
}

func TestCapability_Default_Allows(t *testing.T) {
    ResetPolicy()
    // Should be allowed by default
    f, err := CreateTemp()
    if err != nil { t.Fatalf("CreateTemp: %v", err) }
    _ = f.Close()
}

