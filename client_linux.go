//+build linux

package ipvs

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/cloudflare/ipvs/internal/cipvs"
	"github.com/josharian/native"
	"github.com/mdlayher/genetlink"
	"github.com/mdlayher/netlink"
)

// client implements Client by connecting to IPVS
// on the local machine over netlink.
type client struct {
	c      *genetlink.Conn
	family genetlink.Family
}

// newClient creates a netlink connection,
// then passes to initClient.
func newClient() (*client, error) {
	c, err := genetlink.Dial(nil)
	if err != nil {
		return nil, err
	}

	return initClient(c)
}

// initClient configures a netlink connection for the
// IPVS family, then returns a configured client.
func initClient(c *genetlink.Conn) (*client, error) {
	f, err := c.GetFamily(cipvs.GenlName)
	if err != nil {
		c.Close()
		return nil, err
	}

	return &client{
		c:      c,
		family: f,
	}, nil
}

// Info fetches the Info object from the netlink connection.
func (c *client) Info() (Info, error) {
	msg := genetlink.Message{
		Header: genetlink.Header{
			Command: cipvs.CmdGetInfo,
			Version: cipvs.GenlVersion,
		},
	}
	flags := netlink.Request

	msgs, err := c.c.Execute(msg, c.family.ID, flags)
	if err != nil {
		return Info{}, err
	}

	if len(msgs) == 0 {
		return Info{}, os.ErrNotExist
	}

	var info Info
	for _, msg := range msgs {
		ad, err := netlink.NewAttributeDecoder(msg.Data)
		if err != nil {
			return Info{}, err
		}

		for ad.Next() {
			switch ad.Type() {
			case cipvs.InfoAttrVersion:
				version := ad.Uint32()
				info.Version[0] = int(version >> 16)
				info.Version[1] = int(version & 0xFF00 >> 8)
				info.Version[2] = int(version & 0xFF)
			case cipvs.InfoAttrConnTabSize:
				info.ConnectionTableSize = ad.Uint32()
			}
		}

		if err := ad.Err(); err != nil {
			return Info{}, err
		}
	}

	return info, nil
}

// Services returns a list of Services from the netlink connection.
func (c *client) Services() ([]ServiceExtended, error) {
	msg := genetlink.Message{
		Header: genetlink.Header{
			Command: cipvs.CmdGetService,
			Version: cipvs.GenlVersion,
		},
	}
	flags := netlink.Request | netlink.Dump

	msgs, err := c.c.Execute(msg, c.family.ID, flags)
	if err != nil {
		return nil, err
	}

	if len(msgs) == 0 {
		return nil, os.ErrNotExist
	}

	svcs := make([]ServiceExtended, 0, len(msgs))
	for _, msg := range msgs {
		var s ServiceExtended
		ad, err := netlink.NewAttributeDecoder(msg.Data)
		if err != nil {
			return nil, err
		}

		for ad.Next() {
			if ad.Type() == cipvs.CmdAttrService {
				ad.Do(unpackService(&s))
			}
		}

		if err := ad.Err(); err != nil {
			return nil, err
		}

		svcs = append(svcs, s)
	}

	return svcs, nil
}

// Services returns a list of Services from the netlink connection.
func (c *client) Service(svc Service) (ServiceExtended, error) {
	ae := netlink.NewAttributeEncoder()
	ae.Do(cipvs.CmdAttrService, packService(svc))
	b, err := ae.Encode()

	if err != nil {
		return ServiceExtended{}, err
	}

	msg := genetlink.Message{
		Header: genetlink.Header{
			Command: cipvs.CmdGetService,
			Version: cipvs.GenlVersion,
		},
		Data: b,
	}
	flags := netlink.Request

	msgs, err := c.c.Execute(msg, c.family.ID, flags)
	if err != nil {
		return ServiceExtended{}, err
	}

	if len(msgs) == 0 {
		return ServiceExtended{}, os.ErrNotExist
	}

	msg = msgs[0]
	var s ServiceExtended
	ad, err := netlink.NewAttributeDecoder(msg.Data)
	if err != nil {
		return ServiceExtended{}, err
	}

	for ad.Next() {
		if ad.Type() == cipvs.CmdAttrService {
			ad.Do(unpackService(&s))
		}
	}

	if err := ad.Err(); err != nil {
		return ServiceExtended{}, err
	}

	return s, nil
}

// CreateService creates a new virtual service.
func (c *client) CreateService(svc Service) error {
	ae := netlink.NewAttributeEncoder()
	ae.Do(cipvs.CmdAttrService, packService(svc))
	b, err := ae.Encode()

	if err != nil {
		return err
	}

	msg := genetlink.Message{
		Header: genetlink.Header{
			Command: cipvs.CmdNewService,
			Version: cipvs.GenlVersion,
		},
		Data: b,
	}
	flags := netlink.Request | netlink.Acknowledge

	r, err := c.c.Execute(msg, c.family.ID, flags)
	if err != nil {
		return err
	}

	if len(r) == 0 {
		return os.ErrInvalid
	}

	return nil
}

// RemoveService deletes a virtual service, and any configured Destinations,
// from IPVS.
func (c *client) RemoveService(svc Service) error {
	ae := netlink.NewAttributeEncoder()
	ae.Do(cipvs.CmdAttrService, packService(svc))
	b, err := ae.Encode()

	if err != nil {
		return err
	}

	msg := genetlink.Message{
		Header: genetlink.Header{
			Command: cipvs.CmdDelService,
			Version: cipvs.GenlVersion,
		},
		Data: b,
	}
	flags := netlink.Request | netlink.Acknowledge

	_, err = c.c.Execute(msg, c.family.ID, flags)
	return err
}

// UpdateService replaces the configuration of a Service.
func (c *client) UpdateService(svc Service) error {
	ae := netlink.NewAttributeEncoder()
	ae.Do(cipvs.CmdAttrService, packService(svc))
	b, err := ae.Encode()

	if err != nil {
		return err
	}

	msg := genetlink.Message{
		Header: genetlink.Header{
			Command: cipvs.CmdSetService,
			Version: cipvs.GenlVersion,
		},
		Data: b,
	}
	flags := netlink.Request | netlink.Acknowledge

	r, err := c.c.Execute(msg, c.family.ID, flags)
	if err != nil {
		return err
	}

	if len(r) == 0 {
		return os.ErrInvalid
	}

	return nil
}

// Destinations returns the configured Destinations for a service.
func (c *client) Destinations(svc Service) ([]DestinationExtended, error) {
	ae := netlink.NewAttributeEncoder()
	ae.Do(cipvs.CmdAttrService, packService(svc))
	b, err := ae.Encode()

	if err != nil {
		return nil, err
	}

	msg := genetlink.Message{
		Header: genetlink.Header{
			Command: cipvs.CmdGetDest,
			Version: cipvs.GenlVersion,
		},
		Data: b,
	}
	flags := netlink.Request | netlink.Dump

	msgs, err := c.c.Execute(msg, c.family.ID, flags)
	if err != nil {
		return nil, err
	}

	if len(msgs) == 0 {
		return nil, os.ErrNotExist
	}

	dests := make([]DestinationExtended, 0, len(msgs))
	for _, msg := range msgs {
		var dest DestinationExtended
		ad, err := netlink.NewAttributeDecoder(msg.Data)
		if err != nil {
			return nil, err
		}

		for ad.Next() {
			if ad.Type() == cipvs.CmdAttrDest {
				ad.Do(unpackDestination(&dest))
			}
		}

		if err := ad.Err(); err != nil {
			return nil, err
		}

		dests = append(dests, dest)
	}

	return dests, nil
}

// CreateDestination creates a Destination for the Service.
func (c *client) CreateDestination(svc Service, dest Destination) error {
	ae := netlink.NewAttributeEncoder()
	ae.Do(cipvs.CmdAttrService, packService(svc))
	ae.Do(cipvs.CmdAttrDest, packDest(dest))
	b, err := ae.Encode()

	if err != nil {
		return err
	}

	msg := genetlink.Message{
		Header: genetlink.Header{
			Command: cipvs.CmdNewDest,
			Version: cipvs.GenlVersion,
		},
		Data: b,
	}
	flags := netlink.Request | netlink.Acknowledge

	r, err := c.c.Execute(msg, c.family.ID, flags)
	if err != nil {
		return err
	}

	if len(r) == 0 {
		return os.ErrInvalid
	}

	return nil
}

// UpdateDestination replaces the configuration of a Destination.
func (c *client) UpdateDestination(svc Service, dest Destination) error {
	ae := netlink.NewAttributeEncoder()
	ae.Do(cipvs.CmdAttrService, packService(svc))
	ae.Do(cipvs.CmdAttrDest, packDest(dest))
	b, err := ae.Encode()

	if err != nil {
		return err
	}

	msg := genetlink.Message{
		Header: genetlink.Header{
			Command: cipvs.CmdSetDest,
			Version: cipvs.GenlVersion,
		},
		Data: b,
	}
	flags := netlink.Request | netlink.Acknowledge

	r, err := c.c.Execute(msg, c.family.ID, flags)
	if err != nil {
		return err
	}

	if len(r) == 0 {
		return os.ErrInvalid
	}

	return nil
}

// RemoveDestination removes the Destinaation from a Service.
func (c *client) RemoveDestination(svc Service, dest Destination) error {
	ae := netlink.NewAttributeEncoder()
	ae.Do(cipvs.CmdAttrService, packService(svc))
	ae.Do(cipvs.CmdAttrDest, packDest(dest))
	b, err := ae.Encode()

	if err != nil {
		return err
	}

	msg := genetlink.Message{
		Header: genetlink.Header{
			Command: cipvs.CmdDelDest,
			Version: cipvs.GenlVersion,
		},
		Data: b,
	}
	flags := netlink.Request | netlink.Acknowledge

	_, err = c.c.Execute(msg, c.family.ID, flags)
	return err
}

// Close implements io.Closer
func (c *client) Close() error {
	return c.c.Close()
}

// unpackService unpacks a Service from a netlink-encoded message
func unpackService(svc *ServiceExtended) func(b []byte) error {
	return func(b []byte) error {
		ad, err := netlink.NewAttributeDecoder(b)
		if err != nil {
			return err
		}

		var addr []byte
		var flags []byte
		for ad.Next() {
			switch ad.Type() {
			case cipvs.SvcAttrAf:
				svc.Family = AddressFamily(ad.Uint16())
			case cipvs.SvcAttrProtocol:
				svc.Protocol = Protocol(ad.Uint16())
			case cipvs.SvcAttrAddr:
				addr = ad.Bytes()
			case cipvs.SvcAttrPort:
				ad.Do(unpackPort(&svc.Port))
			case cipvs.SvcAttrFlags:
				flags = ad.Bytes()
			case cipvs.SvcAttrFwmark:
				svc.FWMark = ad.Uint32()
			case cipvs.SvcAttrSchedName:
				svc.Scheduler = ad.String()
			case cipvs.SvcAttrTimeout:
				svc.Timeout = ad.Uint32()
			case cipvs.SvcAttrNetmask:
				copy(svc.Netmask[:], ad.Bytes())
			case cipvs.SvcAttrStats:
				ad.Do(unpackStats(&svc.Stats))
			case cipvs.SvcAttrStats64:
				ad.Do(unpackStats64(&svc.Stats64))
			}
		}
		if err = ad.Err(); err != nil {
			return err
		}

		if svc.FWMark == 0 {
			copy(svc.Address[:], addr)
		}

		if len(flags) != 8 {
			return fmt.Errorf("ipvs: flags attribute is not a uint32; length: %d", len(flags))
		}
		f := native.Endian.Uint32(flags)
		svc.Flags = Flags(f)

		return nil
	}
}

// packService encodes the service attributes
func packService(svc Service) func() ([]byte, error) {
	return func() ([]byte, error) {
		flags := make([]byte, 4)
		native.Endian.PutUint32(flags, uint32(svc.Flags))
		flags = append(flags, []byte{0xFF, 0xFF, 0xFF, 0xFF}...)

		ae := netlink.NewAttributeEncoder()
		ae.Uint16(cipvs.SvcAttrAf, uint16(svc.Family))
		ae.String(cipvs.SvcAttrSchedName, svc.Scheduler)
		ae.Bytes(cipvs.SvcAttrFlags, flags)
		ae.Uint32(cipvs.SvcAttrTimeout, svc.Timeout)
		ae.Bytes(cipvs.SvcAttrNetmask, svc.Netmask[:])

		if svc.FWMark != 0 {
			ae.Uint32(cipvs.SvcAttrFwmark, svc.FWMark)
		} else {
			ae.Uint16(cipvs.SvcAttrProtocol, uint16(svc.Protocol))
			ae.Bytes(cipvs.SvcAttrAddr, svc.Address[:])
			ae.Do(cipvs.SvcAttrPort, packPort(svc.Port))
		}

		return ae.Encode()
	}
}

// unpackDestination unpacks a Destination from a netlink-encoded message
func unpackDestination(dest *DestinationExtended) func(b []byte) error {
	return func(b []byte) error {
		ad, err := netlink.NewAttributeDecoder(b)
		if err != nil {
			return err
		}

		for ad.Next() {
			switch ad.Type() {
			case cipvs.DestAttrAddr:
				copy(dest.Address[:], ad.Bytes())
			case cipvs.DestAttrPort:
				ad.Do(unpackPort(&dest.Port))
			case cipvs.DestAttrFwdMethod:
				dest.FwdMethod = ForwardType(ad.Uint32())
			case cipvs.DestAttrWeight:
				dest.Weight = ad.Uint32()
			case cipvs.DestAttrUThresh:
				dest.UpperThreshold = ad.Uint32()
			case cipvs.DestAttrLThresh:
				dest.LowerThreshold = ad.Uint32()
			case cipvs.DestAttrActiveConns:
				dest.ActiveConnections = ad.Uint32()
			case cipvs.DestAttrInactConns:
				dest.InactiveConnections = ad.Uint32()
			case cipvs.DestAttrPersistConns:
				dest.PersistentConnections = ad.Uint32()
			case cipvs.DestAttrAddrFamily:
				dest.Family = AddressFamily(ad.Uint16())
			case cipvs.DestAttrStats:
				ad.Do(unpackStats(&dest.Stats))
			case cipvs.DestAttrStats64:
				ad.Do(unpackStats64(&dest.Stats64))
			}
		}
		if err = ad.Err(); err != nil {
			return err
		}

		return nil
	}
}

// packDest encodes the destination attributes
func packDest(dest Destination) func() ([]byte, error) {
	return func() ([]byte, error) {
		ae := netlink.NewAttributeEncoder()
		ae.Uint16(cipvs.DestAttrAddrFamily, uint16(dest.Family))
		ae.Bytes(cipvs.DestAttrAddr, dest.Address[:])
		ae.Do(cipvs.DestAttrPort, packPort(dest.Port))
		ae.Uint32(cipvs.DestAttrFwdMethod, uint32(dest.FwdMethod))
		ae.Uint32(cipvs.DestAttrWeight, dest.Weight)
		ae.Uint32(cipvs.DestAttrUThresh, dest.UpperThreshold)
		ae.Uint32(cipvs.DestAttrLThresh, dest.LowerThreshold)

		return ae.Encode()
	}
}

// unpackStats unpacks Stats from the 32-bit netlink message.
func unpackStats(stats *Stats) func(b []byte) error {
	return func(b []byte) error {
		ad, err := netlink.NewAttributeDecoder(b)
		if err != nil {
			return err
		}

		for ad.Next() {
			switch ad.Type() {
			case cipvs.StatsAttrConns:
				stats.Connections = uint64(ad.Uint32())
			case cipvs.StatsAttrInpkts:
				stats.IncomingPackets = uint64(ad.Uint32())
			case cipvs.StatsAttrOutpkts:
				stats.OutgoingPackets = uint64(ad.Uint32())
			case cipvs.StatsAttrInbytes:
				stats.IncomingBytes = ad.Uint64()
			case cipvs.StatsAttrOutbytes:
				stats.OutgoingBytes = ad.Uint64()

			case cipvs.StatsAttrCps:
				stats.ConnectionRate = uint64(ad.Uint32())
			case cipvs.StatsAttrInpps:
				stats.IncomingPacketRate = uint64(ad.Uint32())
			case cipvs.StatsAttrOutpps:
				stats.OutgoingPacketRate = uint64(ad.Uint32())
			case cipvs.StatsAttrInbps:
				stats.IncomingByteRate = uint64(ad.Uint32())
			case cipvs.StatsAttrOutbps:
				stats.OutgoingByteRate = uint64(ad.Uint32())
			}
		}

		return ad.Err()
	}
}

// unpackStats64 unpacks Stats from the 64-but netlink messages
func unpackStats64(stats *Stats) func(b []byte) error {
	return func(b []byte) error {
		ad, err := netlink.NewAttributeDecoder(b)
		if err != nil {
			return err
		}

		for ad.Next() {
			switch ad.Type() {
			case cipvs.StatsAttrConns:
				stats.Connections = ad.Uint64()
			case cipvs.StatsAttrInpkts:
				stats.IncomingPackets = ad.Uint64()
			case cipvs.StatsAttrOutpkts:
				stats.OutgoingPackets = ad.Uint64()
			case cipvs.StatsAttrInbytes:
				stats.IncomingBytes = ad.Uint64()
			case cipvs.StatsAttrOutbytes:
				stats.OutgoingBytes = ad.Uint64()

			case cipvs.StatsAttrCps:
				stats.ConnectionRate = ad.Uint64()
			case cipvs.StatsAttrInpps:
				stats.IncomingPacketRate = ad.Uint64()
			case cipvs.StatsAttrOutpps:
				stats.OutgoingPacketRate = ad.Uint64()
			case cipvs.StatsAttrInbps:
				stats.IncomingByteRate = ad.Uint64()
			case cipvs.StatsAttrOutbps:
				stats.OutgoingByteRate = ad.Uint64()
			}
		}

		return ad.Err()
	}
}

// unpackPort unpacks a port from a netlink message.
func unpackPort(port *uint16) func(b []byte) error {
	return func(b []byte) error {
		if len(b) != 2 {
			return fmt.Errorf("ipvs: port attribute is not a uint16; length: %d", len(b))
		}

		x := binary.BigEndian.Uint16(b)
		*port = x

		return nil
	}
}

// packPort packs a port into a byte slice for a netlink message.
func packPort(port uint16) func() ([]byte, error) {
	return func() ([]byte, error) {
		out := make([]byte, 2)
		binary.BigEndian.PutUint16(out, port)

		return out, nil
	}
}
