package dict

import (
	"errors"
)

const (
	byteRS = 0x1E

	crcSize = 4

	MinKeyLen = 1
	MaxKeyLen = 128
)

var (
	ErrCorrupt = errors.New(`dict: raw data is corrupt`)
)
