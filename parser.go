package jsonport

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var (
	jsontrue  = []byte("true")
	jsonfalse = []byte("false")
	jsonnull  = []byte("null")
)

func isSpace(b byte) bool {
	switch b {
	case ' ':
	case '\t':
	case '\n':
	case '\r':
	case '\v':
	case '\f':
	default:
		return false
	}
	return true
}

func skipSpace(b []byte) int {
	for i := range b {
		if isSpace(b[i]) {
			continue
		}
		return i
	}
	return len(b)
}

func parse(b []byte) (Json, int, error) {
	i := skipSpace(b)

	var j Json
	p := b[i:]
	switch p[0] {
	case '{':
		o, ii, err := parseObject(p)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.v = o
		j.tp = OBJECT
	case '[':
		a, ii, err := parseArray(p)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.v = a
		j.tp = ARRAY
	case '"':
		s, ii, err := parseString(p)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.v = s
		j.tp = STRING
	case 't', 'f':
		tf, ii, err := parseBool(p)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.v = tf
		j.tp = BOOL
	case 'n':
		ii, err := parseNull(p)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.v = nil
		j.tp = NULL
	default:
		n, ii, err := parseNumber(p)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.v = n
		j.tp = NUMBER
	}
	return j, i, nil
}

func parseString(b []byte) (string, int, error) {
	if b[0] != '"' {
		return "", 0, fmt.Errorf("STRING: expect '\"' found '%c'", b[0])
	}
	var i int
	var quoted bool
	for i = 1; i < len(b); i++ {
		if !quoted && b[i] == '\\' {
			quoted = true
			continue
		}
		if !quoted && b[i] == '"' {
			s, ok := unquote(b[:i+1])
			if !ok {
				return "", i, errors.New("STRING: unquote err")
			}
			return s, i + 1, nil
		}
		quoted = false
	}
	return "", i, errors.New("STRING: end of string err")
}

func parseObject(b []byte) (map[string]Json, int, error) {
	if len(b) == 0 {
		return nil, 0, errors.New("OBJECT: expect '{' found EOF")
	}
	if b[0] != '{' {
		return nil, 0, fmt.Errorf("OBJECT: expect '{' found '%c", b[0])
	}
	if len(b) < 2 {
		return nil, 1, errors.New("OBJECT: expect '}' found EOF")
	}
	if b[1] == '}' {
		return map[string]Json{}, 2, nil
	}

	const (
		stateMemberName  = 1
		stateColon       = 2
		stateMemberValue = 3
		stateDone        = 4
	)
	state := stateMemberName

	var m = make(map[string]Json)
	var k string

	i := 1 // skip {
	for i < len(b) {
		if isSpace(b[i]) {
			i++
			continue
		}
		if state == stateMemberName {
			s, ii, err := parseString(b[i:])
			if err != nil {
				return nil, i, fmt.Errorf("OBJECT member.name: %s", err)
			}
			i += ii
			k = s
			state = stateColon
			continue
		}
		if state == stateColon {
			if b[i] != ':' {
				return nil, i, fmt.Errorf("OBJECT: expect ':' found '%c'", b[i])
			}
			i++
			state = stateMemberValue
			continue
		}

		if state == stateMemberValue {
			j, ii, err := parse(b[i:])
			if err != nil {
				return nil, i, fmt.Errorf("OBJECT member.value: %s", err)
			}
			i += ii
			m[k] = j
			state = stateDone
			continue
		}

		if state == stateDone {
			if b[i] == ',' {
				i++
				state = stateMemberName
				continue
			}
			if b[i] == '}' {
				i++
				return m, i, nil
			}
			return nil, i, fmt.Errorf("OBJECT: expect ',' or '}' found '%c'", b[i])
		}
	}
	return nil, i, errors.New("OBJECT: internal err")
}

func parseArray(b []byte) ([]Json, int, error) {
	if len(b) == 0 {
		return nil, 0, errors.New("ARRAY: expect '[' found EOF")
	}
	if b[0] != '[' {
		return nil, 0, fmt.Errorf("ARRAY: expect '[' found '%c'", b[0])
	}
	if len(b) < 2 {
		return nil, 1, errors.New("ARRAY: expect ']' found EOF")
	}
	if b[1] == ']' {
		return []Json{}, 2, nil
	}

	const (
		stateValue = 1
		stateDone  = 2
	)
	state := stateValue

	var a []Json

	i := 1 // skip [
	for i < len(b) {
		if isSpace(b[i]) {
			i++
			continue
		}
		if state == stateValue {
			j, ii, err := parse(b[i:])
			if err != nil {
				return a, i, fmt.Errorf("ARRAY value: %s", err)
			}
			i += ii
			a = append(a, j)
			state = stateDone
			continue
		}

		if state == stateDone {
			if b[i] == ',' {
				i++
				state = stateValue
				continue
			}
			if b[i] == ']' {
				i++
				return a, i, nil
			}
			return nil, i, fmt.Errorf("ARRAY: expect ',' or ']' found '%c'", b[i])
		}
	}
	return nil, i, errors.New("ARRAY: internal err")
}

func parseNumber(b []byte) (Number, int, error) {
	if len(b) == 0 {
		return zero, 0, errors.New("parse EOF err")
	}
	c := b[0]
	if c != '-' && (c < '0' || c > '9') {
		return zero, 0, errors.New("Unknown type")
	}
	var i int
	for ; i < len(b); i++ {
		c := b[i]
		switch {
		case c >= '0' && c <= '9':
		case c == '.':
		case c == 'e':
		case c == 'E':
		case c == '+':
		case c == '-':
		default:
			return Number(string(b[:i])), i, nil
		}
	}
	return Number(string(b[:i])), i, nil
}

func parseBool(b []byte) (bool, int, error) {
	if bytes.HasPrefix(b, jsontrue) {
		return true, len(jsontrue), nil
	}
	if bytes.HasPrefix(b, jsonfalse) {
		return false, len(jsonfalse), nil
	}
	return false, 0, errors.New("BOOL: not true nor false")
}

func parseNull(b []byte) (int, error) {
	if bytes.HasPrefix(b, jsonnull) {
		return len(jsonnull), nil
	}
	return 0, errors.New("NULL: parse err")
}

// unquote converts a quoted JSON string literal s into an actual string t.
// The rules are different than for Go, so cannot use strconv.Unquote.
func unquote(s []byte) (t string, ok bool) {
	s, ok = unquoteBytes(s)
	t = string(s)
	return
}

func unquoteBytes(s []byte) (t []byte, ok bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return
	}
	s = s[1 : len(s)-1]

	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room?  Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	r, err := strconv.ParseUint(string(s[2:6]), 16, 64)
	if err != nil {
		return -1
	}
	return rune(r)
}
