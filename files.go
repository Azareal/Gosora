package main
import "io"
import "os"
import "errors"

type SFile struct
{
	Data []byte
	Pos int64
	Length int64
	Mimetype string
	Info os.FileInfo
	FormattedModTime string
}

func (r SFile) Read(b []byte) (n int, err error) {
	n = 0
	if r.Pos > r.Length {
		return n, io.EOF
	}
	
	size := cap(b)
	if size > 0 {
		for n < size {
			b[n] = r.Data[r.Pos]
			n++
			if r.Pos == r.Length {
				break
			}
			r.Pos++
		}
	}
	return n, nil
}

func (r SFile) Seek(offset int64, whence int) (int64, error) {
	if offset < 0 {
		return 0, errors.New("negative position")
	}
	switch whence {
		case 0:
			r.Pos = offset
		case 1:
			r.Pos += offset
		case 2:
			r.Pos = r.Length + offset
		default:
			return 0, errors.New("invalid whence")
	}
	return r.Pos, nil
}