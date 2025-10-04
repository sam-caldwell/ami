package io

// Policy controls I/O capability enforcement for stdlib io. Defaults allow all.
type Policy struct { AllowFS, AllowNet, AllowDevice bool }

