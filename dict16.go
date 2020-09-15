package dict16

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
	"sync"
)

const (
	byteRS = 0x1E

	idSize     = 16 / 8
	minRawSize = idSize + MinKeyLen + 1
	maxRawSize = idSize + MaxKeyLen + 1

	MinKeyLen = 1
	MaxKeyLen = 128
	MaxSize   = math.MaxUint16 * maxRawSize
)

type Dict16 struct {
	dict map[string]uint16
	rev  map[uint16]string

	lastID uint16

	mu sync.Mutex
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
	d.mu.Lock()
	key, ok := d.rev[id]
	d.mu.Unlock()

	return key, ok
}

func (d *Dict16) GetKeys(ids []uint16) []string {
	keys := make([]string, len(ids))

	for i, id := range ids {
		if key, ok := d.GetKey(id); ok {
			keys[i] = key
		}
	}

	return keys
}

func (d *Dict16) GetPrefix(prefix string) map[string]uint16 {
	result := make(map[string]uint16)

	for key, id := range d.dict {
		if strings.HasPrefix(key, prefix) {
			result[key] = id
		}
	}

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

func (d *Dict16) Unmarshal(raw []byte) {
	d.mu.Lock()

	d.lastID = binary.BigEndian.Uint16(raw[0:])
	raw = raw[idSize:]

	for len(raw) > minRawSize {
		id := binary.BigEndian.Uint16(raw[0:])
		if id == 0 {
			break
		}

		ke := bytes.IndexByte(raw, byteRS)
		if ke < 1 || len(raw[idSize:ke]) == 0 {
			break
		}
		key := raw[idSize:ke]

		raw = raw[ke+1:]

		d.dict[string(key)] = id
		d.rev[id] = string(key)
	}

	d.mu.Unlock()
}

func (d *Dict16) Marshal() []byte {
	d.mu.Lock()

	raw := make([]byte, maxRawSize*len(d.dict))

	binary.BigEndian.PutUint16(raw[0:], d.lastID)

	offset := idSize
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

	d.mu.Unlock()

	return raw[:offset]
}
