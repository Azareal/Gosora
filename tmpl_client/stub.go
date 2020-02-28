package tmpl

import (
	//"reflect"
	//"runtime"
	//"unsafe"
	"github.com/Azareal/Gosora/uutils"
)

var GetFrag = func(name string) [][]byte {
	return nil
}

type WriteString interface {
	WriteString(s string) (n int, err error)
}

var StringToBytes = uutils.StringToBytes

/*
func StringToBytes(s string) (bytes []byte) {
	str := (*reflect.StringHeader)(unsafe.Pointer(&s))
	slice := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	slice.Data = str.Data
	slice.Len = str.Len
	slice.Cap = str.Len
	runtime.KeepAlive(&s)
	return bytes
}
*/