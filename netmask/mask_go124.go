//go:build go1.24

package netmask

import (
	"encoding"
)

var _ encoding.BinaryAppender = (*Mask)(nil)
var _ encoding.TextAppender = (*Mask)(nil)
