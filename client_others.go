//+build !linux

package ipvs

import (
	"fmt"
	"runtime"
)

var (
	errUnimplemented = fmt.Errorf("ipvs is not implemented on %s/%s",
		runtime.GOOS, runtime.GOARCH)
)

type client struct{}

func newClient() (*client, error) {
	return nil, errUnimplemented
}

func (c *client) Info() (Info, error) {
	return Info{}, errUnimplemented
}

func (c *client) Services() ([]ServiceExtended, error) {
	return nil, errUnimplemented
}

func (c *client) Service(Service) (ServiceExtended, error) {
	return ServiceExtended{}, errUnimplemented
}

func (c *client) CreateService(Service) error {
	return errUnimplemented
}

func (c *client) UpdateService(Service) error {
	return errUnimplemented
}

func (c *client) RemoveService(Service) error {
	return errUnimplemented
}

func (c *client) Destinations(Service) ([]DestinationExtended, error) {
	return nil, errUnimplemented
}

func (c *client) CreateDestination(Service, Destination) error {
	return errUnimplemented
}

func (c *client) UpdateDestination(Service, Destination) error {
	return errUnimplemented
}

func (c *client) RemoveDestination(Service, Destination) error {
	return errUnimplemented
}
