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
	idSize32     = 32 / 8
	minRawSize32 = idSize32 + MinKeyLen + 1
	maxRawSize32 = idSize32 + MaxKeyLen + 1

	MaxSize32 = math.MaxUint32 * maxRawSize32
)

type Dict32 struct {
	dict map[string]uint32
	rev  map[uint32]string

	lastID uint32

	mu sync.RWMutex
}

func New32() *Dict32 {
	return &Dict32{
		dict: make(map[string]uint32),
		rev:  make(map[uint32]string),
	}
}

func (d *Dict32) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.dict)
}

func (d *Dict32) All() map[uint32]string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.rev
}

func (d *Dict32) GetID(key string) uint32 {
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

func (d *Dict32) GetIDs(keys []string) []uint32 {
	ids := make([]uint32, len(keys))

	for i, key := range keys {
		ids[i] = d.GetID(key)
	}

	return ids
}

func (d *Dict32) GetKey(id uint32) (string, bool) {
	d.mu.RLock()
	key, ok := d.rev[id]
	d.mu.RUnlock()

	return key, ok
}

func (d *Dict32) GetKeys(ids []uint32) []string {
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

func (d *Dict32) GetPrefix(prefix string) map[string]uint32 {
	result := make(map[string]uint32)

	d.mu.RLock()

	for key, id := range d.dict {
		if strings.HasPrefix(key, prefix) {
			result[key] = id
		}
	}

	d.mu.RUnlock()

	return result
}

func (d *Dict32) RenameID(id uint32, newKey string) {
	d.mu.Lock()

	if key, ok := d.rev[id]; ok {
		delete(d.dict, key)

		d.rev[id] = newKey
		d.dict[newKey] = id
	}

	d.mu.Unlock()
}

func (d *Dict32) RenameKey(oldKey, newKey string) {
	d.mu.Lock()

	if id, ok := d.dict[oldKey]; ok {
		delete(d.dict, oldKey)

		d.rev[id] = newKey
		d.dict[newKey] = id
	}

	d.mu.Unlock()
}

// implement interface: BinaryUnmarshaler
func (d *Dict32) UnmarshalBinary(raw []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.lastID = binary.BigEndian.Uint32(raw)
	raw = raw[idSize32:]

	for len(raw) > minRawSize32 {
		id := binary.BigEndian.Uint32(raw)
		if id == 0 {
			return ErrCorrupt
		}

		raw = raw[idSize32:]
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
func (d *Dict32) MarshalBinary() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	raw := make([]byte, maxRawSize32*len(d.dict))

	binary.BigEndian.PutUint32(raw, d.lastID)

	offset := idSize32
	for key, id := range d.dict {
		binary.BigEndian.PutUint32(raw[offset:], id)
		offset += idSize32

		for i, c := range key {
			raw[offset+i] = byte(c)
		}

		offset += len(key)
		raw[offset] = byteRS

		offset += 1
	}

	return raw[:offset], nil
}

func (d *Dict32) LoadFile(file string) error {
	fh, err := os.OpenFile(file, os.O_CREATE|os.O_RDONLY, 0664)
	if err != nil {
		return err
	}
	defer fh.Close()

	d.mu.Lock()
	defer d.mu.Unlock()

	raw := make([]byte, idSize32)
	_, err = fh.ReadAt(raw, 0)
	if err != nil {
		return err
	}
	d.lastID = binary.BigEndian.Uint32(raw)

	var offset int64 = idSize32
	for {
		raw := make([]byte, maxRawSize32)
		_, err = fh.ReadAt(raw, offset)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		id := binary.BigEndian.Uint32(raw)
		if id == 0 {
			return ErrCorrupt
		}

		raw = raw[idSize32:]
		ke := bytes.IndexByte(raw, byteRS)
		if ke < 0 {
			return ErrCorrupt
		}
		key := string(raw[:ke])

		raw = raw[ke+1:]

		d.dict[key] = id
		d.rev[id] = key

		offset += idSize32 + 1 + int64(len(key))
	}

	return nil
}

func (d *Dict32) SaveFile(file string) error {
	fh, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer fh.Close()

	d.mu.RLock()
	defer d.mu.RUnlock()

	raw := make([]byte, idSize32)
	binary.BigEndian.PutUint32(raw, d.lastID)
	_, err = fh.Write(raw)
	if err != nil {
		return err
	}

	for key, id := range d.dict {
		offset := 0
		raw := make([]byte, idSize32+len(key)+1)

		binary.BigEndian.PutUint32(raw[offset:], id)
		offset += idSize32

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