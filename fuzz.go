//+build gofuzz

package ipvs

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func FuzzService(data []byte) int {
	var svc Service
	if err := unpackService(&svc)(data); err != nil {
		return 0
	}

	return 1
}

func FuzzServiceRoundTrip(data []byte) int {
	var svc1 Service
	var svc2 Service

	if err := unpackService(&svc1)(data); err != nil {
		return 0
	}

	p, err := packService(svc1)()
	if err != nil {
		panic(err)
	}

	if err := unpackService(&svc2)(p); err != nil {
		panic(err)
	}

	if svc1.FWMark != 0 {
		if svc1.FWMark != svc2.FWMark {
			panic("FWMark differ")
		}
	} else {
		switch false {
		case cmp.Equal(svc1.Address, svc2.Address):
			panic("addresses differ")
		case svc1.Port == svc2.Port:
			panic("port differ")
		case svc1.Protocol == svc2.Protocol:
			panic("protocol differ")
		}
	}

	opts := []cmp.Option{
		cmpopts.IgnoreFields(
			Service{},
			"Stats",
			"Stats64",
			"FWMark",
			"Address",
			"Port",
			"Protocol",
		),
		cmpopts.EquateEmpty(),
	}
	if diff := cmp.Diff(svc1, svc2, opts...); diff != "" {
		panic(diff)
	}

	return 1
}

func FuzzDestination(data []byte) int {
	var dest Destination
	if err := unpackDestination(&dest)(data); err != nil {
		return 0
	}

	return 1
}

func FuzzDestintationRoundTrip(data []byte) int {
	var dest Destination
	var dest2 Destination

	if err := unpackDestination(&dest)(data); err != nil {
		return 0
	}

	p, err := packDest(dest)()
	if err != nil {
		panic(err)
	}

	if err := unpackDestination(&dest2)(p); err != nil {
		panic(err)
	}

	opt := cmpopts.IgnoreFields(
		Destination{},
		"ActiveConnections",
		"InactiveConnections",
		"PersistentConnections",
		"Stats",
		"Stats64",
	)

	if diff := cmp.Diff(dest, dest2, opt); diff != "" {
		panic(diff)
	}

	return 1
}
