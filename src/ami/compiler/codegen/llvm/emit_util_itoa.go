package llvm

// itoa is a small helper to avoid importing strconv repeatedly in minimal emitter surface
func itoa(n int) string {
    if n == 0 { return "0" }
    var buf [20]byte
    i := len(buf)
    for n > 0 {
        i--
        buf[i] = byte('0' + (n % 10))
        n /= 10
    }
    return string(buf[i:])
}
