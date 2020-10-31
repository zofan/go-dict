package dict

import (
	"errors"
)

const (
	byteRS = 0x1E

	MinKeyLen = 1
	MaxKeyLen = 128
)

var (
	ErrCorrupt = errors.New(`dict: raw data is corrupt`)
)
