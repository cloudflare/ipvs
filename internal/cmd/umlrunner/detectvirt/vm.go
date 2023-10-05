package detectvirt

import (
	"bufio"
	"io"
	"os"
	"strings"
)

func UserModeLinux() bool {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return false
	}

	return detectUserModeLinux(f)
}

func detectUserModeLinux(r io.Reader) bool {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		l := scanner.Text()

		l, ok := strings.CutPrefix(l, "vendor_id\t: ")
		if !ok {
			continue
		}

		return strings.HasPrefix(l, "User Mode Linux")
	}

	return false
}
