package dict

import (
	"bufio"
	"math"
	"os"
	"strconv"
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

func (d *Dict8) All() map[uint8]string {
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

func (d *Dict8) LoadFile(file string) error {
	fh, err := os.OpenFile(file, os.O_CREATE|os.O_RDONLY, 0664)
	if err != nil {
		return err
	}
	defer fh.Close()

	d.mu.Lock()
	defer d.mu.Unlock()

	s := bufio.NewScanner(fh)

	for s.Scan() {
		line := strings.TrimSpace(s.Text())

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		i := strings.IndexByte(line, ';')
		if i <= 0 {
			continue
		}
		idRaw, key := line[:i], line[i+1:]

		id, err := strconv.ParseInt(idRaw, 10, idSize8*8)
		if err != nil {
			continue
		}

		if uint8(id) > d.lastID {
			d.lastID = uint8(id)
		}

		d.dict[key] = uint8(id)
		d.rev[uint8(id)] = key
	}

	return nil
}

func (d *Dict8) SaveFile(file string) error {
	fh, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer fh.Close()

	d.mu.RLock()
	defer d.mu.RUnlock()

	for k, id := range d.dict {
		if _, err = fh.WriteString(strconv.Itoa(int(id)) + `;` + k + "\n"); err != nil {
			return err
		}
	}

	return nil
}
