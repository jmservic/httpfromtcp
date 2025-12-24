package headers

import (
	"bytes"
	"errors"
	//"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

const crlf = "\r\n"

var ascii_Letters_Digits *unicode.RangeTable = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: uint16('A'), Hi: uint16('Z'), Stride: 1},
		{Lo: uint16('0'), Hi: uint16('9'), Stride: 1},
		{Lo: uint16('a'), Hi: uint16('z'), Stride: 1},
	},
}

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}

	header := string(data[:idx])
	key, value, found := strings.Cut(header, ":")
	if !found {
		return 0, false, errors.New("Couldn't determine key-value pair in header")
	}

	if unicode.IsSpace(rune(key[len(key)-1])) {
		return 0, false, errors.New("Found space between the colon and the key.")
	}

	key = strings.TrimSpace(key)
	if strings.IndexFunc(key, invalidToken) != -1 {
		return 0, false, errors.New("Key contains invalid tokens")
	}

	value = strings.TrimSpace(value)
	h.Set(key, value)
	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if prevVal, ok := h[key]; ok {
		value = prevVal + ", " + value
	}
	h[key] = value
}

func (h Headers) Replace(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Get(key string) (string, bool) {
	val, ok := h[strings.ToLower(key)]
	return val, ok
}

func (h Headers) Delete(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}

func invalidToken(r rune) bool {
	//fmt.Printf("%s: %t\n", string(r), unicode.In(r, unicode.ASCII_Hex_Digit))
	if unicode.In(r, ascii_Letters_Digits) {
		return false
	}
	if strings.ContainsAny(string(r), "!#$%&'*+-.^`|~") {
		return false
	}
	return true
}
