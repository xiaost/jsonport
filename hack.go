package jsonport

import "unsafe"

func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
