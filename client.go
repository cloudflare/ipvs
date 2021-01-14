// Package ipvs provides access to Linux's IPVS kernel service
// via netlink.
package ipvs

import (
	"fmt"
	"strings"
)

// Client represents an opaque IPVS client.
// This would most commonly be connected to IPVS running on the same machine,
// but may represent a connection to a broker on another machine.
type Client interface {
	Info() (Info, error)

	Services() ([]ServiceExtended, error)
	Service(Service) (ServiceExtended, error)
	CreateService(Service) error
	UpdateService(Service) error
	RemoveService(Service) error

	Destinations(Service) ([]DestinationExtended, error)
	CreateDestination(Service, Destination) error
	UpdateDestination(Service, Destination) error
	RemoveDestination(Service, Destination) error
}

// Service represents a virtual server.
//
// When referencing an existing Service, only the identifying fields
// (Address, Port, Family, and Protocol) are required to be set.
type Service struct {
	Address   IP
	Netmask   IPMask
	Scheduler string
	Timeout   uint32
	Flags     Flags
	Port      uint16
	FWMark    uint32
	Family    AddressFamily
	Protocol  Protocol
}

// ServiceExtended contains fields that are not necessary for
// comparison of the identity of a Service.
type ServiceExtended struct {
	Service
	Stats   Stats
	Stats64 Stats
}

// Destination represents a connection to the real server.
type Destination struct {
	Address        IP
	FwdMethod      ForwardType
	Weight         uint32
	UpperThreshold uint32
	LowerThreshold uint32
	Port           uint16
	Family         AddressFamily
}

// DestinationExtended contains fields that are not neccesarry
// for comparison of the identity of a Destination.
type DestinationExtended struct {
	Destination
	ActiveConnections     uint32
	InactiveConnections   uint32
	PersistentConnections uint32
	Stats                 Stats
	Stats64               Stats
}

// Stats represents the statistics of a Service as a whole,
// or the individual Destination connections.
type Stats struct {
	Connections     uint64
	IncomingPackets uint64
	OutgoingPackets uint64
	IncomingBytes   uint64
	OutgoingBytes   uint64

	ConnectionRate     uint64
	IncomingPacketRate uint64 // pktbs
	OutgoingPacketRate uint64 // pktbs
	IncomingByteRate   uint64 // bps
	OutgoingByteRate   uint64 // bps
}

// Info returns basic high-level information about the IPVS instance.
type Info struct {
	Version             [3]int
	ConnectionTableSize uint32
}

// New returns an instance of Client.
func New() (Client, error) {
	// BUG(terin): We might want to make the client type configurable in calls to New.
	return newClient()
}

//go:generate stringer -type=ForwardType,AddressFamily,Protocol --output zz_generated.stringer.go

// ForwardType configures how IPVS forwards traffic to the real server.
type ForwardType uint32

// Well-known forwarding types.
const (
	Masquarade ForwardType = iota
	Local
	Tunnel
	DirectRoute
	Bypass
)

// AddressFamily determines if the Service or Destination is configured to use
// IPv4 or IPv6 family.
type AddressFamily uint16

// Address families known to IPVS.
const (
	INET  AddressFamily = 0x2
	INET6 AddressFamily = 0xA
)

// Protocol configures how IPVS listens for connections to the virtual service.
type Protocol uint16

// The protocols IPVS is aware of.
const (
	TCP  Protocol = 0x06
	UDP  Protocol = 0x11
	SCTP Protocol = 0x84
)

// Flags tweak the behavior of a virtual service, and the chosen scheduler.
type Flags uint32

// Well-known flags.
const (
	ServicePersistent    Flags = 0x0001
	ServiceHashed        Flags = 0x0002
	ServiceOnePacket     Flags = 0x0004
	ServiceSchedulerOpt1 Flags = 0x0008
	ServiceSchedulerOpt2 Flags = 0x0010
	ServiceSchedulerOpt3 Flags = 0x0020
)

// String returns a human readable representation of flags.
func (i Flags) String() string {
	flags := []string{}

	if i&ServicePersistent != 0 {
		flags = append(flags, "ServicePersistent")
	}
	if i&ServiceHashed != 0 {
		flags = append(flags, "ServiceHashed")
	}
	if i&ServiceOnePacket != 0 {
		flags = append(flags, "ServiceOnePacket")
	}
	if i&ServiceSchedulerOpt1 != 0 {
		flags = append(flags, "ServiceSchedulerOpt1")
	}
	if i&ServiceSchedulerOpt2 != 0 {
		flags = append(flags, "ServiceSchedulerOpt2")
	}
	if i&ServiceSchedulerOpt3 != 0 {
		flags = append(flags, "ServiceSchedulerOpt3")
	}
	if j := i &^ (ServicePersistent | ServiceHashed | ServiceOnePacket | ServiceSchedulerOpt1 | ServiceSchedulerOpt2 | ServiceSchedulerOpt3); j != 0 {
		flags = append(flags, fmt.Sprintf("%#x", uint32(j)))
	}

	return strings.Join(flags, " | ")
}
