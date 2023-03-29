# ipvs [![Go Reference](https://pkg.go.dev/badge/github.com/cloudflare/ipvs#section-readme.svg)](https://pkg.go.dev/github.com/cloudflare/ipvs#section-readme)

Package `ipvs` provides programmatic access to Linux's IPVS to manage services
and destinations using the [netlink](github.com/mdlayher/netlink) and
[genetlink](https://pkg.go.dev/github.com/mdlayher/genetlink) packages. This
package can be used in environment without the `ipvsadm` tool, and in programs
compiled without CGO.

Usage examples can be found in the [Go Reference](https://pkg.go.dev/github.com/cloudflare/ipvs#pkg-examples).
