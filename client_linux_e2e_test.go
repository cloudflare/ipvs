// Tests in this file differ from those in the other files, as these
// will actually connect over netlink and configure IPVS in the kernel, then
// check the results against the "ipvsadm" tool.
//
// Is it intended to be ran only within a virtualized environment. Running
// on your host system may break your system's network configuration.
//
// You have been warned.

package ipvs_test

import (
	"log"
	"net/netip"
	"testing"

	"github.com/cloudflare/ipvs"
	"github.com/cloudflare/ipvs/internal/cmd/umlrunner/detectvirt"
	"github.com/cloudflare/ipvs/netmask"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
)

func TestClient(t *testing.T) {
	if !detectvirt.UserModeLinux() {
		t.Skip("did not detect test running in User Mode Linux environment")
	}

	c, err := ipvs.New()
	if err != nil {
		log.Fatalf("error updating service: %v", err)
	}

	t.Cleanup(func() {
		res := icmd.RunCommand("ipvsadm", "-C")
		res.Assert(t, icmd.Success)
	})

	svc := ipvs.Service{
		Address:   netip.MustParseAddr("127.0.1.1"),
		Netmask:   netmask.MaskFrom4([...]byte{255, 255, 255, 255}),
		Family:    ipvs.INET,
		Protocol:  ipvs.TCP,
		Port:      443,
		Scheduler: "mh",
	}

	assert.NilError(t, c.CreateService(svc))
	assert.NilError(t, c.CreateDestination(svc, ipvs.Destination{
		Address:   netip.MustParseAddr("127.0.2.1"),
		FwdMethod: ipvs.DirectRoute,
		Weight:    0,
		Port:      80,
		Family:    ipvs.INET,
	}))

	res := icmd.RunCommand("ipvsadm", "-Sn")
	res.Assert(t, icmd.Expected{
		ExitCode: 0,
		Timeout:  false,
		Out:      string(golden.Get(t, "ipvsadm-simple-service.golden")),
	})
}
