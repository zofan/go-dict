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
		t.Error(`expected key1 equal "hello world!"`)
	}

	raw, _ := d.MarshalBinary()

	d2 := New()
	_ = d2.UnmarshalBinary(raw)

	id5 := d2.GetID(`good bye!`)
	if id5 != id3 {
		t.Error(`expected same id5 and id3`)
	}

	key2, ok := d2.GetKey(id1)
	if !ok {
		t.Error(`expected give key2`)
	}
	if key2 != `hello world!` {
		t.Error(`expected key2 equal "hello world!"`)
	}

	d2.RenameID(id1, `new_key`)
	key2, ok = d2.GetKey(id1)
	if key2 != `new_key` {
		t.Error(`expected key2 equal "new_key"`)
	}

	id22 := d2.GetID(`new_key`)
	if id2 != id22 {
		t.Error(`expected id2 equal id22`)
	}

	d2.RenameKey(`new_key`, `new_key2`)
	key2, ok = d2.GetKey(id1)
	if key2 != `new_key2` {
		t.Error(`expected key2 equal "new_key2"`)
	}

	id22 = d2.GetID(`new_key2`)
	if id2 != id22 {
		t.Error(`expected id2 equal id22`)
	}
}
