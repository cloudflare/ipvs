package ipvs

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_NewIP(t *testing.T) {
	tests := map[string]struct {
		ip       net.IP
		expected IP
	}{
		"IPv4-len16": {
			ip:       net.ParseIP("127.0.0.1").To16(),
			expected: IP{127, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		"IPv4-len4": {
			ip:       net.ParseIP("127.0.0.1").To4(),
			expected: IP{127, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		"IPv6": {
			ip:       net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
			expected: IP{0x20, 0x01, 0x0d, 0xb8, 0x85, 0xa3, 0, 0, 0, 0, 0x8a, 0x2e, 0x03, 0x70, 0x73, 0x34},
		},
		"IPv6 leading": {
			ip:       net.ParseIP("ff00::"),
			expected: IP{0xff, 0x00},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewIP(tt.ip)

			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Fatalf("diff:\n%s", diff)
			}
		})
	}
}

func TestIP_Net(t *testing.T) {
	tests := map[string]struct {
		af       AddressFamily
		ip       IP
		expected net.IP
	}{
		"IPv4": {
			af:       INET,
			ip:       IP{127, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expected: net.ParseIP("127.0.0.1"),
		},
		"IPv6": {
			af:       INET6,
			ip:       IP{0x20, 0x01, 0x0d, 0xb8, 0x85, 0xa3, 0, 0, 0, 0, 0x8a, 0x2e, 0x03, 0x70, 0x73, 0x34},
			expected: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
		},
		"IPv6 leading": {
			af:       INET6,
			ip:       IP{0xff, 0x00},
			expected: net.ParseIP("ff00::"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.ip.Net(tt.af)

			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Fatalf("diff:\n%s", diff)
			}
		})
	}
}

func Test_NewIPMask(t *testing.T) {
	tests := map[string]struct {
		mask     net.IPMask
		expected IPMask
	}{
		"IPv4": {
			mask:     net.CIDRMask(32, 32),
			expected: IPMask{255, 255, 255, 255},
		},
		"IPv4-constructor": {
			mask:     net.IPv4Mask(250, 250, 250, 250),
			expected: IPMask{250, 250, 250, 250},
		},
		"IPv6": {
			mask:     net.CIDRMask(1, 128),
			expected: IPMask{1, 0, 0, 0},
		},
		"IPv6-full": {
			mask:     net.CIDRMask(128, 128),
			expected: IPMask{128, 0, 0, 0},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewIPMask(tt.mask)

			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Fatalf("diff:\n%s", diff)
			}
		})
	}
}

func TestIPMask_Net(t *testing.T) {
	tests := map[string]struct {
		af       AddressFamily
		mask     IPMask
		expected net.IPMask
	}{
		"IPv4": {
			af:       INET,
			mask:     IPMask{255, 255, 255, 255},
			expected: net.CIDRMask(32, 32),
		},
		"IPv6": {
			af:       INET6,
			mask:     IPMask{1, 0, 0, 0},
			expected: net.CIDRMask(1, 128),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.mask.Net(tt.af)

			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Fatalf("diff:\n%s", diff)
			}
		})
	}
}
