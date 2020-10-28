// +build ignore

package main

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	replacer = strings.NewReplacer(
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
		`binary.BigEndian.Uint16(raw[crcSize:])`, `uint8(raw[crcSize])`,
		`binary.BigEndian.Uint16(raw[0:])`, `uint8(raw[0])`,
		`binary.BigEndian.PutUint16(raw[crcSize:], d.lastID)`, `raw[crcSize] = d.lastID`,
		`binary.BigEndian.PutUint16(raw[offset:], id)`, `raw[offset] = id`,
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

	code = replacer.Replace(code)

	err = ioutil.WriteFile(path + `dict8.go`, []byte(code), 0664)
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

	code = replacer.Replace(code)

	err = ioutil.WriteFile(path + `dict8_test.go`, []byte(code), 0664)
	if err != nil {
		panic(err)
	}
}
