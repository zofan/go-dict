// +build ignore

package main

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	replacer8 = strings.NewReplacer(
		`"encoding/binary"`, `_ "encoding/binary"`,
		`MaxUint16`, `MaxUint8`,
		`Test16`, `Test8`,
		`Dict16`, `Dict8`,
		`New16`, `New8`,
		`uint16`, `uint8`,
		`MaxSize16`, `MaxSize8`,
		`idSize16`, `idSize8`,
		`minRawSize16`, `minRawSize8`,
		`maxRawSize16`, `maxRawSize8`,
		`16 / 8`, `1`,
		`binary.BigEndian.Uint16(raw)`, `uint8(raw[0])`,
		`binary.BigEndian.PutUint16(raw, d.lastID)`, `raw[0] = d.lastID`,
		`binary.BigEndian.PutUint16(raw[offset:], id)`, `raw[offset] = id`,
	)

	replacer32 = strings.NewReplacer(
		`MaxUint16`, `MaxUint32`,
		`Test16`, `Test32`,
		`Dict16`, `Dict32`,
		`New16`, `New32`,
		`uint16`, `uint32`,
		`MaxSize16`, `MaxSize32`,
		`idSize16`, `idSize32`,
		`minRawSize16`, `minRawSize32`,
		`maxRawSize16`, `maxRawSize32`,
		`16 / 8`, `32 / 8`,
		`binary.BigEndian.Uint16(raw)`, `binary.BigEndian.Uint32(raw)`,
		`binary.BigEndian.PutUint16(raw, d.lastID)`, `binary.BigEndian.PutUint32(raw, d.lastID)`,
		`binary.BigEndian.PutUint16(raw[offset:], id)`, `binary.BigEndian.PutUint32(raw[offset:], id)`,
	)
)

func main() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic(`can't get path`)
	}
	path := filepath.Dir(file) + `/`

	genCode(path)
	genTest(path)
}

func genCode(path string) {
	raw, err := ioutil.ReadFile(path + `dict16.go`)
	if err != nil {
		panic(err)
	}

	code := string(raw)

	err = ioutil.WriteFile(path + `dict8.go`, []byte(replacer8.Replace(code)), 0664)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(path + `dict32.go`, []byte(replacer32.Replace(code)), 0664)
	if err != nil {
		panic(err)
	}
}

func genTest(path string) {
	raw, err := ioutil.ReadFile(path + `dict16_test.go`)
	if err != nil {
		panic(err)
	}

	code := string(raw)

	err = ioutil.WriteFile(path + `dict8_test.go`, []byte(replacer8.Replace(code)), 0664)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(path + `dict32_test.go`, []byte(replacer32.Replace(code)), 0664)
	if err != nil {
		panic(err)
	}
}
