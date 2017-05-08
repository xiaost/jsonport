package jsonport

import (
	"errors"
	"fmt"
)

func jsonskipObject(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, errors.New("OBJECT: expect '{' found EOF")
	}
	if b[0] != '{' {
		return 0, fmt.Errorf("OBJECT: expect '{' found '%c", b[0])
	}
	if len(b) < 2 {
		return 1, errors.New("OBJECT: expect '}' found EOF")
	}
	if b[1] == '}' {
		return 2, nil
	}

	const (
		stateMemberName  = 1
		stateColon       = 2
		stateMemberValue = 3
		stateDone        = 4
	)
	state := stateMemberName

	i := 1 // skip {
	for i < len(b) {
		if isspace(b[i]) {
			i++
			continue
		}
		if state == stateMemberName {
			_, ii, err := parseString(b[i:])
			if err != nil {
				return i, fmt.Errorf("OBJECT member.name: %s", err)
			}
			i += ii
			state = stateColon
			continue
		}
		if state == stateColon {
			if b[i] != ':' {
				return i, fmt.Errorf("OBJECT: expect ':' found '%c'", b[i])
			}
			i++
			state = stateMemberValue
			continue
		}

		if state == stateMemberValue {
			ii, err := jsonskip(b[i:])
			if err != nil {
				return i, fmt.Errorf("OBJECT member.value: %s", err)
			}
			i += ii
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
				return i, nil
			}
			return i, fmt.Errorf("OBJECT: expect ',' or '}' found '%c'", b[i])
		}
	}
	return i, errObjectEOF
}

func jsonskipArray(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, errors.New("ARRAY: expect '[' found EOF")
	}
	if b[0] != '[' {
		return 0, fmt.Errorf("ARRAY: expect '[' found '%c'", b[0])
	}
	if len(b) < 2 {
		return 1, errors.New("ARRAY: expect ']' found EOF")
	}
	if b[1] == ']' {
		return 2, nil
	}

	const (
		stateValue = 1
		stateDone  = 2
	)
	state := stateValue

	pos := 0
	i := 1 // skip [
	for i < len(b) {
		if isspace(b[i]) {
			i++
			continue
		}
		if state == stateValue {
			ii, err := jsonskip(b[i:])
			if err != nil {
				return i, fmt.Errorf("ARRAY: index: %d err: %s", pos, err)
			}
			pos += 1
			i += ii
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
				return i, nil
			}
			return i, fmt.Errorf("ARRAY: expect ',' or ']' found '%c'", b[i])
		}
	}
	return i, errors.New("ARRAY")
}

func jsonskip(b []byte) (int, error) {
	i := skipspace(b)
	b = b[i:]
	if len(b) == 0 {
		return 0, errJSONEOF
	}
	switch b[0] {
	case '{': // skip to unquoted '}'
		return jsonskipObject(b)
	case '[':
		return jsonskipArray(b)

	case '"':
		_, ii, err := parseString(b)
		return i + ii, err
	case 't', 'f':
		_, ii, err := parseBool(b)
		return i + ii, err
	case 'n':
		ii, err := parseNull(b)
		return i + ii, err
	default:
		_, ii, err := parseNumber(b)
		return i + ii, err
	}
}
