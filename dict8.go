package dict

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"math"
	"strings"
	"sync"
)

const (
	idSize8     = 1
	minRawSize8 = idSize8 + MinKeyLen + 1
	maxRawSize8 = idSize8 + MaxKeyLen + 1

	MaxSize8 = math.MaxUint8 * maxRawSize8
)

type Dict8 struct {
	dict map[string]uint8
	rev  map[uint8]string

	lastID uint8

	mu sync.RWMutex
}

func New8() *Dict8 {
	return &Dict8{
		dict: make(map[string]uint8),
		rev:  make(map[uint8]string),
	}
}

func (d *Dict8) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.dict)
}

func (d *Dict8) GetAll() map[uint8]string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.rev
}

func (d *Dict8) GetID(key string) uint8 {
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

func (d *Dict8) GetIDs(keys []string) []uint8 {
	ids := make([]uint8, len(keys))

	for i, key := range keys {
		ids[i] = d.GetID(key)
	}

	return ids
}

func (d *Dict8) GetKey(id uint8) (string, bool) {
	d.mu.RLock()
	key, ok := d.rev[id]
	d.mu.RUnlock()

	return key, ok
}

func (d *Dict8) GetKeys(ids []uint8) []string {
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

func (d *Dict8) GetPrefix(prefix string) map[string]uint8 {
	result := make(map[string]uint8)

	d.mu.RLock()

	for key, id := range d.dict {
		if strings.HasPrefix(key, prefix) {
			result[key] = id
		}
	}

	d.mu.RUnlock()

	return result
}

func (d *Dict8) RenameID(id uint8, newKey string) {
	d.mu.Lock()

	if key, ok := d.rev[id]; ok {
		delete(d.dict, key)

		d.rev[id] = newKey
		d.dict[newKey] = id
	}

	d.mu.Unlock()
}

func (d *Dict8) RenameKey(oldKey, newKey string) {
	d.mu.Lock()

	if id, ok := d.dict[oldKey]; ok {
		delete(d.dict, oldKey)

		d.rev[id] = newKey
		d.dict[newKey] = id
	}

	d.mu.Unlock()
}

// implement interface: BinaryUnmarshaler
func (d *Dict8) UnmarshalBinary(raw []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	sum := binary.BigEndian.Uint32(raw[0:])
	if sum != crc32.ChecksumIEEE(raw[crcSize:]) {
		return ErrCorrupt
	}

	d.lastID = uint8(raw[crcSize])
	raw = raw[idSize8+crcSize:]

	for len(raw) > minRawSize8 {
		id := uint8(raw[0])
		if id == 0 {
			return ErrCorrupt
		}

		ke := bytes.IndexByte(raw, byteRS)
		if ke < 0 {
			return ErrCorrupt
		}
		key := string(raw[idSize8:ke])

		raw = raw[ke+1:]

		d.dict[key] = id
		d.rev[id] = key
	}

	return nil
}

// implement interface: BinaryMarshaler
func (d *Dict8) MarshalBinary() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	raw := make([]byte, maxRawSize8*len(d.dict))

	raw[crcSize] = d.lastID

	offset := idSize8 + crcSize
	for key, id := range d.dict {
		raw[offset] = id
		offset += idSize8

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
