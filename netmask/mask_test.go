package netmask

import (
	"testing"

	"gotest.tools/v3/assert"
	"pgregory.net/rapid"
)

func TestNetmask_MaskFrom4(t *testing.T) {
	type testCase struct {
		name     string
		mask     [4]byte
		expected Mask
	}

	run := func(t *testing.T, tc testCase) {
		mask := MaskFrom4(tc.mask)
		assert.Equal(t, mask, tc.expected)
	}

	testCases := []testCase{
		{
			name: "0",
			mask: [...]byte{0, 0, 0, 0},
			expected: Mask{
				mask: 0,
				z:    z4,
			},
		},
		{
			name: "8",
			mask: [...]byte{255, 0, 0, 0},
			expected: Mask{
				mask: 0xFF00_0000,
				z:    z4,
			},
		},
		{
			name: "24",
			mask: [...]byte{255, 255, 255, 0},
			expected: Mask{
				mask: 0xFFFF_FF00,
				z:    z4,
			},
		},
		{
			name: "32",
			mask: [...]byte{255, 255, 255, 255},
			expected: Mask{
				mask: 0xFFFF_FFFF,
				z:    z4,
			},
		},
		{
			name: "non-canonical mask",
			mask: [...]byte{0, 0, 255, 0},
			expected: Mask{
				mask: 0x0000_FF00,
				z:    z4,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_MaskFrom16(t *testing.T) {
	type testCase struct {
		name     string
		mask     [16]byte
		expected Mask
	}

	run := func(t *testing.T, tc testCase) {
		mask := MaskFrom16(tc.mask)
		assert.Equal(t, mask, tc.expected)
	}

	testCases := []testCase{
		{
			name: "0",
			mask: [...]byte{
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
			},
			expected: Mask{
				mask: 0,
				z:    z6,
			},
		},
		{
			name: "128",
			mask: [...]byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			},
			expected: Mask{
				mask: 128,
				z:    z6,
			},
		},
		{
			name: "96",
			mask: [...]byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0,
			},
			expected: Mask{
				mask: 96,
				z:    z6,
			},
		},
		{
			name: "non-prefix",
			mask: [...]byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0,
				0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0,
			},
			expected: Mask{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNemask_MaskFromSlice(t *testing.T) {
	type testCase struct {
		name     string
		mask     []byte
		expected Mask
	}

	run := func(t *testing.T, tc testCase) {
		mask, ok := MaskFromSlice(tc.mask)
		assert.Assert(t, ok)
		assert.Equal(t, mask, tc.expected)
	}

	testCases := []testCase{
		{
			name: "IPv4 /0",
			mask: []byte{0, 0, 0, 0},
			expected: Mask{
				mask: 0,
				z:    z4,
			},
		},
		{
			name: "IPv4 /24",
			mask: []byte{255, 255, 255, 0},
			expected: Mask{
				mask: 0xFFFF_FF00,
				z:    z4,
			},
		},
		{
			name: "IPv6 /128",
			mask: []byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			},
			expected: Mask{
				mask: 128,
				z:    z6,
			},
		},
		{
			name: "IPv6 /96",
			mask: []byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0,
			},
			expected: Mask{
				mask: 96,
				z:    z6,
			},
		},
		{
			name: "non-prefix",
			mask: []byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0,
				0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0,
			},
			expected: Mask{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNemask_AsSlice(t *testing.T) {
	type testCase struct {
		name     string
		mask     Mask
		expected []byte
	}

	run := func(t *testing.T, tc testCase) {
		slice := tc.mask.AsSlice()
		assert.DeepEqual(t, slice, tc.expected)
	}

	testCases := []testCase{
		{
			name:     "zero Mask",
			mask:     Mask{},
			expected: nil,
		},
		{
			name: "IPv4 /0",
			mask: Mask{
				mask: 0,
				z:    z4,
			},
			expected: []byte{0, 0, 0, 0},
		},
		{
			name: "IPv4 /24",
			mask: Mask{
				mask: 0xFFFF_FF00,
				z:    z4,
			},
			expected: []byte{255, 255, 255, 0},
		},
		{
			name: "IPv6 /128",
			mask: Mask{
				mask: 128,
				z:    z6,
			},
			expected: []byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			},
		},
		{
			name: "IPv6 /96",
			mask: Mask{
				mask: 96,
				z:    z6,
			},
			expected: []byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_SliceRoundtrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		mask := rapid.Custom[Mask](func(t *rapid.T) Mask {
			z := rapid.SampledFrom([]int8{z4, z6}).Draw(t, "z")

			switch z {
			case z4:
				return Mask{mask: rapid.Uint32().Draw(t, "mask"), z: z4}
			case z6:
				return MaskFrom(rapid.IntRange(0, 128).Draw(t, "ones"), 128)
			default:
				return Mask{}
			}
		}).Draw(t, "mask")

		slice := mask.AsSlice()
		out, ok := MaskFromSlice(slice)
		assert.Assert(t, ok)

		assert.DeepEqual(t, out, mask)
	})
}

func TestNetmask_MaskFrom(t *testing.T) {
	type testCase struct {
		name     string
		ones     int
		bits     int
		expected Mask
	}

	run := func(t *testing.T, tc testCase) {
		mask := MaskFrom(tc.ones, tc.bits)
		assert.Equal(t, mask, tc.expected)
	}

	testCases := []testCase{
		{
			name: "IPv4 /0",
			ones: 0,
			bits: 32,
			expected: Mask{
				mask: 0,
				z:    z4,
			},
		},
		{
			name: "IPv4 /24",
			ones: 24,
			bits: 32,
			expected: Mask{
				mask: 0xFFFF_FF00,
				z:    z4,
			},
		},
		{
			name: "IPv6 /128",
			ones: 128,
			bits: 128,
			expected: Mask{
				mask: 128,
				z:    z6,
			},
		},
		{
			name: "IPv6 /96",
			ones: 96,
			bits: 128,
			expected: Mask{
				mask: 96,
				z:    z6,
			},
		},
		{
			name:     "invalid ones",
			ones:     255,
			bits:     32,
			expected: Mask{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_IsValid(t *testing.T) {
	type testCase struct {
		name     string
		mask     Mask
		expected bool
	}

	run := func(t *testing.T, tc testCase) {
		assert.Equal(t, tc.mask.IsValid(), tc.expected)
	}

	testCases := []testCase{
		{
			name: "IPv4",
			mask: Mask{
				mask: 0,
				z:    z4,
			},
			expected: true,
		},
		{
			name: "IPv6",
			mask: Mask{
				mask: 128,
				z:    z6,
			},
			expected: true,
		},
		{
			name:     "invalid",
			mask:     Mask{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_Is4(t *testing.T) {
	type testCase struct {
		name     string
		mask     Mask
		expected bool
	}

	run := func(t *testing.T, tc testCase) {
		assert.Equal(t, tc.mask.Is4(), tc.expected)
	}

	testCases := []testCase{
		{
			name: "IPv4",
			mask: Mask{
				mask: 0,
				z:    z4,
			},
			expected: true,
		},
		{
			name: "IPv6",
			mask: Mask{
				mask: 128,
				z:    z6,
			},
			expected: false,
		},
		{
			name:     "invalid",
			mask:     Mask{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_Is6(t *testing.T) {
	type testCase struct {
		name     string
		mask     Mask
		expected bool
	}

	run := func(t *testing.T, tc testCase) {
		assert.Equal(t, tc.mask.Is6(), tc.expected)
	}

	testCases := []testCase{
		{
			name: "IPv4",
			mask: Mask{
				mask: 0,
				z:    z4,
			},
			expected: false,
		},
		{
			name: "IPv6",
			mask: Mask{
				mask: 128,
				z:    z6,
			},
			expected: true,
		},
		{
			name:     "invalid",
			mask:     Mask{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_Bits(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		var ones int
		var mask Mask
		z := rapid.SampledFrom([]int8{z0, z4, z6}).Draw(t, "z")

		switch z {
		case z4:
			ones = rapid.IntRange(0, 32).Draw(t, "ones")
			mask = MaskFrom(ones, 32)
		case z6:
			ones = rapid.IntRange(0, 128).Draw(t, "ones")
			mask = MaskFrom(ones, 128)
		default:
			ones = -1
			mask = Mask{}
		}

		assert.DeepEqual(t, mask.Bits(), ones)
	})
}

func TestNetmask_AppendBinary(t *testing.T) {
	type testCase struct {
		name     string
		mask     Mask
		expected []byte
	}

	run := func(t *testing.T, tc testCase) {
		b := make([]byte, 4, 32)
		out, err := tc.mask.AppendBinary(b)
		assert.NilError(t, err)
		assert.DeepEqual(t, out[4:], tc.expected)
	}

	testCases := []testCase{
		{name: "zero mask", mask: Mask{}, expected: []byte{}},
		{name: "ipv4 mask", mask: MaskFrom(31, 32), expected: []byte{0xFF, 0xFF, 0xFF, 0xFE}},
		{name: "weird ipv4 mask", mask: MaskFrom4([...]byte{0xFF, 0x00, 0xFF, 0x00}), expected: []byte{0xFF, 0x00, 0xFF, 0x00}},
		{name: "ipv6 mask", mask: MaskFrom(128, 128), expected: []byte{128}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_MarshalBinary(t *testing.T) {
	type testCase struct {
		name     string
		mask     Mask
		expected []byte
	}

	run := func(t *testing.T, tc testCase) {
		out, err := tc.mask.MarshalBinary()
		assert.NilError(t, err)
		assert.DeepEqual(t, out, tc.expected)
	}

	testCases := []testCase{
		{name: "zero mask", mask: Mask{}, expected: []byte{}},
		{name: "ipv4 mask", mask: MaskFrom(31, 32), expected: []byte{0xFF, 0xFF, 0xFF, 0xFE}},
		{name: "weird ipv4 mask", mask: MaskFrom4([...]byte{0xFF, 0x00, 0xFF, 0x00}), expected: []byte{0xFF, 0x00, 0xFF, 0x00}},
		{name: "ipv6 mask", mask: MaskFrom(128, 128), expected: []byte{128}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_UnmarshalBinary(t *testing.T) {
	type testCase struct {
		name     string
		mask     []byte
		expected Mask
	}

	run := func(t *testing.T, tc testCase) {
		out := Mask{}
		err := out.UnmarshalBinary(tc.mask)
		assert.NilError(t, err)
		assert.DeepEqual(t, out, tc.expected)
	}

	testCases := []testCase{
		{name: "zero mask", mask: []byte{}, expected: Mask{}},
		{name: "ipv4 mask", mask: []byte{0xFF, 0xFF, 0xFF, 0xFE}, expected: MaskFrom(31, 32)},
		{name: "weird ipv4 mask", mask: []byte{0xFF, 0x00, 0xFF, 0x00}, expected: MaskFrom4([...]byte{0xFF, 0x00, 0xFF, 0x00})},
		{name: "ipv6 mask", mask: []byte{128}, expected: MaskFrom(128, 128)},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_BinaryMarshaller(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		mask := rapid.Custom[Mask](func(t *rapid.T) Mask {
			z := rapid.SampledFrom([]int8{z0, z4, z6}).Draw(t, "z")

			switch z {
			case z4:
				return MaskFrom(rapid.IntRange(0, 32).Draw(t, "ones"), 32)
			case z6:
				return MaskFrom(rapid.IntRange(0, 128).Draw(t, "ones"), 128)
			default:
				return Mask{}
			}
		}).Draw(t, "mask")

		p, err := mask.MarshalBinary()
		assert.NilError(t, err)

		out := Mask{}
		assert.NilError(t, out.UnmarshalBinary(p))

		assert.DeepEqual(t, out, mask)
	})
}

func TestNetmask_BinaryEncodingEquivalence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		mask := rapid.Custom(func(t *rapid.T) Mask {
			z := rapid.SampledFrom([]int8{z0, z4, z6}).Draw(t, "z")

			switch z {
			case z4:
				return MaskFrom(rapid.IntRange(0, 32).Draw(t, "ones"), 32)
			case z6:
				return MaskFrom(rapid.IntRange(0, 128).Draw(t, "ones"), 128)
			default:
				return Mask{}
			}
		}).Draw(t, "mask")

		p, err := mask.MarshalBinary()
		assert.NilError(t, err)

		b := make([]byte, 4, 32)
		out, err := mask.AppendBinary(b)
		assert.NilError(t, err)

		assert.DeepEqual(t, p, out[4:])
	})
}

func TestNetmask_MarshalText(t *testing.T) {
	type testCase struct {
		name     string
		mask     Mask
		expected string
	}

	run := func(t *testing.T, tc testCase) {
		out, err := tc.mask.MarshalText()
		assert.NilError(t, err)
		assert.DeepEqual(t, string(out), tc.expected)
	}

	testCases := []testCase{
		{name: "zero mask", mask: Mask{}, expected: ""},
		{name: "ipv4 mask", mask: MaskFrom(31, 32), expected: "255.255.255.254"},
		{name: "weird ipv4 mask", mask: MaskFrom4([...]byte{0xFF, 0x00, 0xFF, 0x00}), expected: "255.0.255.0"},
		{name: "ipv6 mask", mask: MaskFrom(128, 128), expected: "128"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_UnmarshalText(t *testing.T) {
	type testCase struct {
		name     string
		mask     []byte
		expected Mask
	}

	run := func(t *testing.T, tc testCase) {
		out := Mask{}
		err := out.UnmarshalText(tc.mask)
		assert.NilError(t, err)
		assert.DeepEqual(t, out, tc.expected)
	}

	testCases := []testCase{
		{name: "zero mask", mask: []byte{}, expected: Mask{}},
		{name: "ipv4 mask", mask: []byte("255.255.255.254"), expected: Mask{mask: 4294967294, z: z4}},
		{name: "weird ipv4 mask", mask: []byte("255.0.255.0"), expected: Mask{mask: 4278255360, z: z4}},
		{name: "ipv6 mask", mask: []byte("56"), expected: Mask{mask: 56, z: z6}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestNetmask_TextMarshaller(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		mask := rapid.Custom[Mask](func(t *rapid.T) Mask {
			z := rapid.SampledFrom([]int8{z0, z4, z6}).Draw(t, "z")

			switch z {
			case z4:
				return MaskFrom(rapid.IntRange(0, 32).Draw(t, "ones"), 32)
			case z6:
				return MaskFrom(rapid.IntRange(0, 128).Draw(t, "ones"), 128)
			default:
				return Mask{}
			}
		}).Draw(t, "mask")

		p, err := mask.MarshalText()
		assert.NilError(t, err)

		out := Mask{}
		assert.NilError(t, out.UnmarshalText(p))

		assert.DeepEqual(t, out, mask)
	})
}

func TestNetmask_TextEncodingEquivalence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		mask := rapid.Custom(func(t *rapid.T) Mask {
			z := rapid.SampledFrom([]int8{z0, z4, z6}).Draw(t, "z")

			switch z {
			case z4:
				return MaskFrom(rapid.IntRange(0, 32).Draw(t, "ones"), 32)
			case z6:
				return MaskFrom(rapid.IntRange(0, 128).Draw(t, "ones"), 128)
			default:
				return Mask{}
			}
		}).Draw(t, "mask")

		p, err := mask.MarshalText()
		assert.NilError(t, err)

		b := make([]byte, 4, 32)
		out, err := mask.AppendText(b)
		assert.NilError(t, err)

		assert.DeepEqual(t, p, out[4:])
	})
}

func TestNetmask_String(t *testing.T) {
	type testCase struct {
		name     string
		mask     Mask
		expected string
	}

	run := func(t *testing.T, tc testCase) {
		out := tc.mask.String()
		assert.DeepEqual(t, out, tc.expected)
	}

	testCases := []testCase{
		{name: "zero mask", mask: Mask{}, expected: "invalid Mask"},
		{name: "ipv4 mask", mask: MaskFrom(31, 32), expected: "255.255.255.254"},
		{name: "weird ipv4 mask", mask: MaskFrom4([...]byte{0xFF, 0x00, 0xFF, 0x00}), expected: "255.0.255.0"},
		{name: "ipv6 mask", mask: MaskFrom(128, 128), expected: "128"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
