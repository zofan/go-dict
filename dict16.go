package dict

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os"
	"strings"
	"sync"
)

const (
	idSize16     = 16 / 8
	minRawSize16 = idSize16 + MinKeyLen + 1
	maxRawSize16 = idSize16 + MaxKeyLen + 1

	MaxSize16 = math.MaxUint16 * maxRawSize16
)

type Dict16 struct {
	dict map[string]uint16
	rev  map[uint16]string

	lastID uint16

	mu sync.RWMutex
}

func New16() *Dict16 {
	return &Dict16{
		dict: make(map[string]uint16),
		rev:  make(map[uint16]string),
	}
}

func (d *Dict16) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.dict)
}

func (d *Dict16) All() map[uint16]string {
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

	d.lastID = binary.BigEndian.Uint16(raw)
	raw = raw[idSize16:]

	for len(raw) > minRawSize16 {
		id := binary.BigEndian.Uint16(raw)
		if id == 0 {
			return ErrCorrupt
		}

		raw = raw[idSize16:]
		ke := bytes.IndexByte(raw, byteRS)
		if ke < 0 {
			return ErrCorrupt
		}
		key := string(raw[:ke])

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

	raw := make([]byte, maxRawSize16*len(d.dict))

	binary.BigEndian.PutUint16(raw, d.lastID)

	offset := idSize16
	for key, id := range d.dict {
		binary.BigEndian.PutUint16(raw[offset:], id)
		offset += idSize16

		for i, c := range key {
			raw[offset+i] = byte(c)
		}

		offset += len(key)
		raw[offset] = byteRS

		offset += 1
	}

	return raw[:offset], nil
}

func (d *Dict16) LoadFile(file string) error {
	fh, err := os.OpenFile(file, os.O_CREATE|os.O_RDONLY, 0664)
	if err != nil {
		return err
	}
	defer fh.Close()

	d.mu.Lock()
	defer d.mu.Unlock()

	raw := make([]byte, idSize16)
	_, err = fh.ReadAt(raw, 0)
	if err != nil {
		return err
	}
	d.lastID = binary.BigEndian.Uint16(raw)

	var offset int64 = idSize16
	for {
		raw := make([]byte, maxRawSize16)
		_, err = fh.ReadAt(raw, offset)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		id := binary.BigEndian.Uint16(raw)
		if id == 0 {
			return ErrCorrupt
		}

		raw = raw[idSize16:]
		ke := bytes.IndexByte(raw, byteRS)
		if ke < 0 {
			return ErrCorrupt
		}
		key := string(raw[:ke])

		raw = raw[ke+1:]

		d.dict[key] = id
		d.rev[id] = key

		offset += idSize16 + 1 + int64(len(key))
	}

	return nil
}

func (d *Dict16) SaveFile(file string) error {
	fh, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer fh.Close()

	d.mu.RLock()
	defer d.mu.RUnlock()

	raw := make([]byte, idSize16)
	binary.BigEndian.PutUint16(raw, d.lastID)
	_, err = fh.Write(raw)
	if err != nil {
		return err
	}

	for key, id := range d.dict {
		offset := 0
		raw := make([]byte, idSize16+len(key)+1)

		binary.BigEndian.PutUint16(raw[offset:], id)
		offset += idSize16

		for i, c := range key {
			raw[offset+i] = byte(c)
		}
		offset += len(key)

		raw[offset] = byteRS
		offset += 1

		_, err = fh.Write(raw)
		if err != nil {
			return err
		}
	}

	return nil
}