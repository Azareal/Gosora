package uutils

import (
	"reflect"
	"runtime"
	"unsafe"
)

// TODO: Add a safe build mode for things like Google Appengine

func StringToBytes(s string) (bytes []byte) {
	str := (*reflect.StringHeader)(unsafe.Pointer(&s))
	slice := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	slice.Data = str.Data
	slice.Len = str.Len
	slice.Cap = str.Len
	runtime.KeepAlive(&s)
	return bytes
}

func BytesToString(bytes []byte) (s string) {
	slice := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	str := (*reflect.StringHeader)(unsafe.Pointer(&s))
	str.Data = slice.Data
	str.Len = slice.Len
	runtime.KeepAlive(&bytes)
	return s
}

//go:noescape
//go:linkname nanotime runtime.nanotime
func nanotime() int64

func Nanotime() int64 {
	return nanotime()
}