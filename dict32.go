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

func (d *Dict32) LoadFile(file string) error {
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

		id, err := strconv.ParseInt(idRaw, 10, idSize32*8)
		if err != nil {
			continue
		}

		if uint32(id) > d.lastID {
			d.lastID = uint32(id)
		}

		d.dict[key] = uint32(id)
		d.rev[uint32(id)] = key
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

	for k, id := range d.dict {
		if _, err = fh.WriteString(strconv.Itoa(int(id)) + `;` + k + "\n"); err != nil {
			return err
		}
	}

	return nil
}
