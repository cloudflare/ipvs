# ipvs [![Go Reference](https://pkg.go.dev/badge/github.com/cloudflare/ipvs#section-readme.svg)](https://pkg.go.dev/github.com/cloudflare/ipvs#section-readme)

Package `ipvs` provides programmatic access to Linux's IPVS to manage services
and destinations using the [netlink](github.com/mdlayher/netlink) and
[genetlink](https://pkg.go.dev/github.com/mdlayher/genetlink) packages. This
package can be used in environment without the `ipvsadm` tool, and in programs
compiled without CGO.

Usage examples can be found in the [Go Reference](https://pkg.go.dev/github.com/cloudflare/ipvs#pkg-examples).

## Supported Versions

### Go

This project follows the [Go Release Policy][go-release]: the last two major Go releases are supported and tested.
Changes which break unsupported Go releases are not considered breaking changes.

[go-release]: https://go.dev/doc/devel/release#policy

### Linux

This project supports the Linux kernels from [kernel.org][] that are designated "stable" or "longterm".
Changes which break unsupported Linux releases are not considered breaking changes.
