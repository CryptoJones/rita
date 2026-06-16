package util

import (
	"encoding/hex"
	"testing"
)

// FuzzNewFixedStringFromHex ensures hex parsing never panics on arbitrary input
// and that valid 16-byte hex round-trips back to the same (upper-cased) value.
func FuzzNewFixedStringFromHex(f *testing.F) {
	f.Add("0123456789abcdef0123456789abcdef")
	f.Add("")
	f.Add("zz")
	f.Add("deadbeef")
	f.Fuzz(func(t *testing.T, h string) {
		fs, err := NewFixedStringFromHex(h)
		if err != nil {
			return // rejecting bad input is fine; we only care that it doesn't panic
		}
		// A successful parse of a full 16-byte (32 hex char) value must round-trip.
		decoded, decErr := hex.DecodeString(h)
		if decErr == nil && len(decoded) == 16 {
			if got := fs.Hex(); got != toUpperHex(h) {
				t.Fatalf("round-trip mismatch: input %q -> %q", h, got)
			}
		}
	})
}

func toUpperHex(s string) string {
	b, err := hex.DecodeString(s)
	if err != nil {
		return ""
	}
	return (&FixedString{Data: to16(b)}).Hex()
}

func to16(b []byte) [16]byte {
	var out [16]byte
	copy(out[:], b)
	return out
}

// FuzzNewFixedStringHash ensures the hash helper never panics for any combination
// of string arguments and always returns a 16-byte digest on success.
func FuzzNewFixedStringHash(f *testing.F) {
	f.Add("a", "b")
	f.Add("", "")
	f.Add("192.168.1.1", "8.8.8.8")
	f.Fuzz(func(t *testing.T, a, b string) {
		fs, err := NewFixedStringHash(a, b)
		if err != nil {
			return
		}
		if len(fs.Data) != 16 {
			t.Fatalf("expected 16-byte hash, got %d", len(fs.Data))
		}
	})
}

// FuzzValidFQDN ensures FQDN validation is panic-free on arbitrary input.
func FuzzValidFQDN(f *testing.F) {
	f.Add("example.com")
	f.Add("")
	f.Add("...")
	f.Add("a.-b.com")
	f.Fuzz(func(t *testing.T, s string) {
		_ = ValidFQDN(s)
	})
}

// FuzzContainsDomain ensures wildcard domain matching is panic-free for arbitrary
// patterns and hosts.
func FuzzContainsDomain(f *testing.F) {
	f.Add("*.example.com", "host.example.com")
	f.Add("", "")
	f.Add("*", "anything")
	f.Fuzz(func(t *testing.T, pattern, host string) {
		_ = ContainsDomain([]string{pattern}, host)
	})
}
