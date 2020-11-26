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

func (d *Dict16) LoadFile(file string) error {
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

		id, err := strconv.ParseInt(idRaw, 10, idSize16*8)
		if err != nil {
			continue
		}

		if uint16(id) > d.lastID {
			d.lastID = uint16(id)
		}

		d.dict[key] = uint16(id)
		d.rev[uint16(id)] = key
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

	for k, id := range d.dict {
		if _, err = fh.WriteString(strconv.Itoa(int(id)) + `;` + k + "\n"); err != nil {
			return err
		}
	}

	return nil
}
