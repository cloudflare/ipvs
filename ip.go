package ipvs

import (
	"bytes"
	"net"

	"github.com/mdlayher/netlink/nlenc"
)

// IP holds IPVS's representation of an IPv4 or IPv6 address.
type IP [net.IPv6len]byte

// IPMask hold IPVS's representation of an IPv4 or IPv6 network mask.
type IPMask [4]byte

var v4InV6Prefix = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}

// NewIP converts from net.IP to an IPVS representation of the IP address.
func NewIP(ip net.IP) IP {
	p := IP{}
	if len(ip) == net.IPv4len {
		copy(p[:], ip)
		return p
	}
	if bytes.Equal(v4InV6Prefix, ip[:12]) {
		copy(p[:], ip[12:])
		return p
	}
	copy(p[:], ip)
	return p
}

// Net is a helper function to convert from an IPVS IP to the net.IP
// representation. As the address is otherwise ambiguous, the address
// family must be provided.
func (ip IP) Net(af AddressFamily) net.IP {
	switch af {
	case INET:
		return ip[0:4]
	case INET6:
		return ip[:]
	default:
		return nil
	}
}

// NewIPMask converts from net.IPMask to an IPVS representation.
func NewIPMask(mask net.IPMask) IPMask {
	p := IPMask{}
	if len(mask) == net.IPv4len {
		copy(p[:], mask)
		return p
	}

	ones, _ := mask.Size()
	nlenc.PutUint32(p[:], uint32(ones))
	return p
}

// Net is a helper function to convert from an IPVS IP mask to the
// net.IPMask representation. As the mask is otherwise ambiguous, the
// address family must be provided.
func (mask IPMask) Net(af AddressFamily) net.IPMask {
	switch af {
	case INET:
		return mask[:]
	case INET6:
		ones := nlenc.Uint32(mask[:])
		return net.CIDRMask(int(ones), 8*net.IPv6len)
	default:
		return nil
	}
}
