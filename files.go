package main
import "log"
import "strings"
import "mime"
import "errors"
import "os"
import "io"
import "io/ioutil"
import "path/filepath"
import "net/http"

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

func add_static_file(path string, prefix string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	f, err := fi.Stat()
	if err != nil {
		return err
	}
	
	log.Print("Adding the '" + path + "' static file")
	path = strings.TrimPrefix(path, prefix)
	log.Print("Added the '" + path + "' static file")
	
	static_files["/static" + path] = SFile{data,0,int64(len(data)),mime.TypeByExtension(filepath.Ext(prefix + path)),f,f.ModTime().UTC().Format(http.TimeFormat)}
	return nil
}