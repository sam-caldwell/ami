package amitime

// Add returns a new Time advanced by duration d from t.
func Add(t Time, d Duration) Time { return Time{t: t.t.Add(d)} }

