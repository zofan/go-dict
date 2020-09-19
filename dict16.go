package dict16

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"math"
	"strings"
	"sync"
)

const (
	byteRS = 0x1E

	idSize     = 2
	crcSize    = 4
	minRawSize = idSize + MinKeyLen + 1
	maxRawSize = idSize + MaxKeyLen + 1

	MinKeyLen = 1
	MaxKeyLen = 128
	MaxSize   = math.MaxUint16 * maxRawSize
)

var (
	ErrCorrupt = errors.New(`dict16: raw data is corrupt`)
)

type Dict16 struct {
	dict map[string]uint16
	rev  map[uint16]string

	lastID uint16

	mu sync.RWMutex
}

func New() *Dict16 {
	return &Dict16{
		dict: make(map[string]uint16),
		rev:  make(map[uint16]string),
	}
}

func (d *Dict16) Count() int {
	return len(d.dict)
}

func (d *Dict16) GetAll() map[uint16]string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.rev
}

func (d *Dict16) GetID(key string) uint16 {
	d.mu.Lock()

	id, ok := d.dict[key]
	if !ok {
		d.lastID++
		id = d.lastID
		d.dict[key] = id
		d.rev[id] = key
	}

	d.mu.Unlock()

	return id
}

func (d *Dict16) GetIDs(keys []string) []uint16 {
	ids := make([]uint16, len(keys))

	for i, key := range keys {
		ids[i] = d.GetID(key)
	}

	return ids
}

func (d *Dict16) GetKey(id uint16) (string, bool) {
	d.mu.RLock()
	key, ok := d.rev[id]
	d.mu.RUnlock()

	return key, ok
}

func (d *Dict16) GetKeys(ids []uint16) []string {
	keys := make([]string, len(ids))

	d.mu.RLock()

	for i, id := range ids {
		if key, ok := d.rev[id]; ok {
			keys[i] = key
		}
	}

	d.mu.RUnlock()

	return keys
}

func (d *Dict16) GetPrefix(prefix string) map[string]uint16 {
	result := make(map[string]uint16)

	d.mu.RLock()

	for key, id := range d.dict {
		if strings.HasPrefix(key, prefix) {
			result[key] = id
		}
	}

	d.mu.RUnlock()

	return result
}

func (d *Dict16) RenameID(id uint16, newKey string) {
	d.mu.Lock()

	if key, ok := d.rev[id]; ok {
		delete(d.dict, key)

		d.rev[id] = newKey
		d.dict[newKey] = id
	}

	d.mu.Unlock()
}

func (d *Dict16) RenameKey(oldKey, newKey string) {
	d.mu.Lock()

	if id, ok := d.dict[oldKey]; ok {
		delete(d.dict, oldKey)

		d.rev[id] = newKey
		d.dict[newKey] = id
	}

	d.mu.Unlock()
}

// implement interface: BinaryUnmarshaler
func (d *Dict16) UnmarshalBinary(raw []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	sum := binary.BigEndian.Uint32(raw[0:])
	if sum != crc32.ChecksumIEEE(raw[crcSize:]) {
		return ErrCorrupt
	}

	d.lastID = binary.BigEndian.Uint16(raw[crcSize:])
	raw = raw[idSize+crcSize:]

	for len(raw) > minRawSize {
		id := binary.BigEndian.Uint16(raw[0:])
		if id == 0 {
			return ErrCorrupt
		}

		ke := bytes.IndexByte(raw, byteRS)
		if ke < 0 {
			return ErrCorrupt
		}
		key := string(raw[idSize:ke])

		raw = raw[ke+1:]

		d.dict[key] = id
		d.rev[id] = key
	}

	return nil
}

// implement interface: BinaryMarshaler
func (d *Dict16) MarshalBinary() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	raw := make([]byte, maxRawSize*len(d.dict))

	binary.BigEndian.PutUint16(raw[crcSize:], d.lastID)

	offset := idSize + crcSize
	for key, id := range d.dict {
		binary.BigEndian.PutUint16(raw[offset:], id)
		offset += idSize

		for i, c := range key {
			raw[offset+i] = byte(c)
		}

		offset += len(key)
		raw[offset] = byteRS

		offset += 1
	}

	binary.BigEndian.PutUint32(raw[0:], crc32.ChecksumIEEE(raw[crcSize:offset]))

	return raw[:offset], nil
}
