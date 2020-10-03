// +build ignore

package main

import (
	"io/ioutil"
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
	genCode()
	genTest()
}

func genCode() {
	raw, err := ioutil.ReadFile(`dict16.go`)
	if err != nil {
		panic(err)
	}

	code := string(raw)

	code = replacer.Replace(code)

	err = ioutil.WriteFile(`dict8.go`, []byte(code), 0666)
	if err != nil {
		panic(err)
	}
}

func genTest() {
	raw, err := ioutil.ReadFile(`dict16_test.go`)
	if err != nil {
		panic(err)
	}

	code := string(raw)

	code = replacer.Replace(code)

	err = ioutil.WriteFile(`dict8_test.go`, []byte(code), 0666)
	if err != nil {
		panic(err)
	}
}
