package detectvirt

import (
	"io"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
)

func TestUserModeLinux(t *testing.T) {
	type testCase struct {
		name     string
		reader   io.Reader
		expected bool
	}

	run := func(t *testing.T, tc testCase) {
		assert.Equal(t, detectUserModeLinux(tc.reader), tc.expected)
	}

	testCases := []testCase{
		{
			name:   "Intel Celeron",
			reader: golden.Open(t, "intel-celeron.golden"),
		},
		{
			name:     "User Mode Linux",
			reader:   golden.Open(t, "uml.golden"),
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
