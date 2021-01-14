//+build linux

package ipvs

import (
	"io"
	"net"
	"os"
	"testing"

	"github.com/cloudflare/ipvs/internal/cipvs"
	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/genetlink"
	"github.com/mdlayher/genetlink/genltest"
	"github.com/mdlayher/netlink"
	"github.com/mdlayher/netlink/nltest"
)

const familyID = 0x24

func TestServices_IsNotExist(t *testing.T) {
	fn := func(gerq genetlink.Message, _ netlink.Message) ([]genetlink.Message, error) {
		return nil, io.EOF
	}
	client := testClient(t, genltest.CheckRequest(familyID, cipvs.CmdGetService, netlink.Request|netlink.Dump, fn))
	defer client.Close()

	if _, err := client.Services(); !os.IsNotExist(err) {
		t.Fatalf("expected to not exists, but got: %v", err)
	}
}

func TestServices(t *testing.T) {
	tests := map[string]struct {
		msgs     []genetlink.Message
		services []Service
	}{
		"single": {
			msgs: []genetlink.Message{
				{
					Data: nltest.MustMarshalAttributes([]netlink.Attribute{
						{
							Type: cipvs.CmdAttrService,
							Data: nltest.MustMarshalAttributes([]netlink.Attribute{
								{
									Type: cipvs.SvcAttrAf,
									Data: []byte{0x02, 0x00},
								},
								{
									Type: cipvs.SvcAttrProtocol,
									Data: []byte{0x06, 0x00},
								},
								{
									Type: cipvs.SvcAttrAddr,
									Data: []byte{0x7F, 0, 0x01, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
								},
								{
									Type: cipvs.SvcAttrPort,
									Data: []byte{0x00, 0x50},
								},
								{
									Type: cipvs.SvcAttrSchedName,
									Data: []byte("wlc"),
								},
								{
									Type: cipvs.SvcAttrTimeout,
									Data: []byte{0x68, 0x01, 0x00, 0x00},
								},
								{
									Type: cipvs.SvcAttrNetmask,
									Data: []byte{0xFF, 0xFF, 0xFF, 0xFE},
								},
								{
									Type: cipvs.SvcAttrFlags,
									Data: []byte{0, 0, 0, 0, 0, 0, 0, 0},
								},
							}),
						},
					}),
				},
			},
			services: []Service{
				{
					Family:    INET,
					Protocol:  TCP,
					Address:   NewIP(net.IPv4(127, 0, 1, 1)),
					Port:      80,
					Scheduler: "wlc",
					Timeout:   360,
					Netmask:   NewIPMask(net.CIDRMask(31, 32)),
				},
			},
		},
		"single-ipv6": {
			msgs: []genetlink.Message{
				{
					Data: nltest.MustMarshalAttributes([]netlink.Attribute{
						{
							Type: cipvs.CmdAttrService,
							Data: nltest.MustMarshalAttributes([]netlink.Attribute{
								{
									Type: cipvs.SvcAttrAf,
									Data: []byte{0x0A, 0x00},
								},

								{
									Type: cipvs.SvcAttrProtocol,
									Data: []byte{0x06, 0x00},
								},
								{
									Type: cipvs.SvcAttrAddr,
									Data: []byte{0xFF, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
								},
								{
									Type: cipvs.SvcAttrPort,
									Data: []byte{0x00, 0x50},
								},
								{
									Type: cipvs.SvcAttrSchedName,
									Data: []byte("wlc"),
								},
								{
									Type: cipvs.SvcAttrTimeout,
									Data: []byte{0x68, 0x01, 0x00, 0x00},
								},
								{
									Type: cipvs.SvcAttrNetmask,
									Data: []byte{0x80, 0x00, 0x00, 0x00},
								},
								{
									Type: cipvs.SvcAttrFlags,
									Data: []byte{0, 0, 0, 0, 0, 0, 0, 0},
								},
							}),
						},
					}),
				},
			},
			services: []Service{
				{
					Family:    INET6,
					Protocol:  TCP,
					Address:   NewIP(net.ParseIP("ff00::")),
					Port:      80,
					Scheduler: "wlc",
					Timeout:   360,
					Netmask:   NewIPMask(net.CIDRMask(128, 128)),
				},
			},
		},
		"multiple": {
			msgs: []genetlink.Message{
				{
					Data: nltest.MustMarshalAttributes([]netlink.Attribute{
						{
							Type: cipvs.CmdAttrService,
							Data: nltest.MustMarshalAttributes([]netlink.Attribute{
								{
									Type: cipvs.SvcAttrAf,
									Data: []byte{0x02, 0x00},
								},
								{
									Type: cipvs.SvcAttrProtocol,
									Data: []byte{0x06, 0x00},
								},
								{
									Type: cipvs.SvcAttrAddr,
									Data: []byte{0x7F, 0, 0x01, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
								},
								{
									Type: cipvs.SvcAttrPort,
									Data: []byte{0x00, 0x50},
								},
								{
									Type: cipvs.SvcAttrSchedName,
									Data: []byte("wlc"),
								},
								{
									Type: cipvs.SvcAttrTimeout,
									Data: []byte{0x68, 0x01, 0x00, 0x00},
								},
								{
									Type: cipvs.SvcAttrNetmask,
									Data: []byte{0xFF, 0xFF, 0xFF, 0xFE},
								},
								{
									Type: cipvs.SvcAttrFlags,
									Data: []byte{0, 0, 0, 0, 0, 0, 0, 0},
								},
							}),
						},
					}),
				},
				{
					Data: nltest.MustMarshalAttributes([]netlink.Attribute{
						{
							Type: cipvs.CmdAttrService,
							Data: nltest.MustMarshalAttributes([]netlink.Attribute{
								{
									Type: cipvs.SvcAttrAf,
									Data: []byte{0x0A, 0x00},
								},
								{
									Type: cipvs.SvcAttrProtocol,
									Data: []byte{0x06, 0x00},
								},
								{
									Type: cipvs.SvcAttrAddr,
									Data: []byte{0xFF, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
								},
								{
									Type: cipvs.SvcAttrPort,
									Data: []byte{0x00, 0x50},
								},
								{
									Type: cipvs.SvcAttrSchedName,
									Data: []byte("wlc"),
								},
								{
									Type: cipvs.SvcAttrTimeout,
									Data: []byte{0x68, 0x01, 0x00, 0x00},
								},
								{
									Type: cipvs.SvcAttrNetmask,
									Data: []byte{0x80, 0x00, 0x00, 0x00},
								},
								{
									Type: cipvs.SvcAttrFlags,
									Data: []byte{0, 0, 0, 0, 0, 0, 0, 0},
								},
							}),
						},
					}),
				},
			},
			services: []Service{
				{
					Family:    INET,
					Protocol:  TCP,
					Address:   NewIP(net.IPv4(127, 0, 1, 1)),
					Port:      80,
					Scheduler: "wlc",
					Timeout:   360,
					Netmask:   NewIPMask(net.CIDRMask(31, 32)),
				},
				{
					Family:    INET6,
					Protocol:  TCP,
					Address:   NewIP(net.ParseIP("ff00::")),
					Port:      80,
					Scheduler: "wlc",
					Timeout:   360,
					Netmask:   NewIPMask(net.CIDRMask(128, 128)),
				},
			},
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			fn := func(gerq genetlink.Message, _ netlink.Message) ([]genetlink.Message, error) {
				return tt.msgs, nil
			}
			client := testClient(t, genltest.CheckRequest(familyID, cipvs.CmdGetService, netlink.Request|netlink.Dump, fn))

			defer client.Close()

			se, err := client.Services()
			if err != nil {
				t.Fatalf("failed to get services: %v", err)
			}

			services := []Service{}
			for _, svc := range se {
				services = append(services, svc.Service)
			}

			if diff := cmp.Diff(tt.services, services); diff != "" {
				t.Fatalf("unexpected services (-want +got):\n%s", diff)
			}
		})
	}
}

func testClient(t *testing.T, fn genltest.Func) *client {
	t.Helper()

	family := genetlink.Family{
		ID:      familyID,
		Version: cipvs.GenlVersion,
		Name:    cipvs.GenlName,
	}

	conn := genltest.Dial(genltest.ServeFamily(family, fn))
	client, err := initClient(conn)
	if err != nil {
		t.Fatalf("failed to open client: %v", err)
	}

	return client
}

func TestUnpackServiceCrashers(t *testing.T) {
	var crashers = []string{
		"\x06\x00\x01\x00\n\x0000\x14\x00\x03\x0000000000" +
			"00000000\f\x00\t\x0000000000",
		"\x05\x00\x04\x000000",
		"\x06\x00\x01\x00\x02\x0000\x14\x00\x03\x0000000000" +
			"00000000\x06\x00\a\x000000\b\x00\t\x00" +
			"0000",
	}

	for _, crash := range crashers {
		var svc ServiceExtended
		_ = unpackService(&svc)([]byte(crash))
	}
}
