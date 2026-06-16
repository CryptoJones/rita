package util

import (
	"net"
	"testing"
)

// BenchmarkNewFixedStringHash measures the md5 row-key hashing that runs once per
// imported record, making it one of the hottest paths in the importer.
func BenchmarkNewFixedStringHash(b *testing.B) {
	args := []string{"192.168.1.100", "8.8.8.8", "udp", "53", "ffffffff-ffff-ffff-ffff-ffffffffffff"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := NewFixedStringHash(args...); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewFixedStringFromHex(b *testing.B) {
	const h = "0123456789abcdef0123456789abcdef"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := NewFixedStringFromHex(h); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseNetworkID covers the per-connection network classification used
// throughout import.
func BenchmarkParseNetworkID(b *testing.B) {
	ip := net.ParseIP("10.45.12.7")
	const agentID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseNetworkID(ip, agentID)
	}
}

func BenchmarkIPIsPubliclyRoutable(b *testing.B) {
	ips := []net.IP{
		net.ParseIP("8.8.8.8"),
		net.ParseIP("192.168.1.1"),
		net.ParseIP("172.16.0.5"),
		net.ParseIP("2606:4700:4700::1111"),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IPIsPubliclyRoutable(ips[i%len(ips)])
	}
}

func BenchmarkValidFQDN(b *testing.B) {
	hosts := []string{"www.example.com", "a.b.c.example.co.uk", "not_a_host", "192.168.1.1"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidFQDN(hosts[i%len(hosts)])
	}
}

func BenchmarkContainsDomain(b *testing.B) {
	domains := []string{"*.example.com", "google.com", "*.internal.lan", "cdn.cloudflare.net"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ContainsDomain(domains, "host.example.com")
	}
}
