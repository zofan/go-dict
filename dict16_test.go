package dict16

import (
	"testing"
)

func TestA(t *testing.T) {
	d := New()

	id1 := d.GetID(`hello world!`)
	if id1 == 0 {
		t.Error(`expected id greatest zero`)
	}

	id2 := d.GetID(`hello world!`)
	if id1 != id2 {
		t.Error(`expected same id1 and id2`)
	}

	id3 := d.GetID(`good bye!`)

	key1, ok := d.GetKey(id1)
	if !ok {
		t.Error(`expected give key1`)
	}
	if key1 != `hello world!` {
		t.Error(`expected give key1`)
	}

	raw := d.Marshal()

	d2 := New()
	d2.Unmarshal(raw)

	id5 := d2.GetID(`good bye!`)
	if id5 != id3 {
		t.Error(`expected same id5 and id3`)
	}

	key2, ok := d2.GetKey(id1)
	if !ok {
		t.Error(`expected give key2`)
	}
	if key2 != `hello world!` {
		t.Error(`expected give key2`)
	}
}
