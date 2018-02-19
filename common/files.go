package common

import (
	"bytes"
	"mime"
	"strings"
	"sync"
	//"errors"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type SFileList map[string]SFile

var StaticFiles SFileList = make(map[string]SFile)
var staticFileMutex sync.RWMutex

type SFile struct {
	Data             []byte
	GzipData         []byte
	Pos              int64
	Length           int64
	GzipLength       int64
	Mimetype         string
	Info             os.FileInfo
	FormattedModTime string
}

type CSSData struct {
	ComingSoon string
}

func (list SFileList) Init() error {
	return filepath.Walk("./public", func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}

		path = strings.Replace(path, "\\", "/", -1)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		path = strings.TrimPrefix(path, "public/")
		var ext = filepath.Ext("/public/" + path)
		gzipData := compressBytesGzip(data)

		list.Set("/static/"+path, SFile{data, gzipData, 0, int64(len(data)), int64(len(gzipData)), mime.TypeByExtension(ext), f, f.ModTime().UTC().Format(http.TimeFormat)})

		DebugLogf("Added the '%s' static file.", path)
		return nil
	})
}

func (list SFileList) Add(path string, prefix string) error {
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

	var ext = filepath.Ext(path)
	path = strings.TrimPrefix(path, prefix)
	gzipData := compressBytesGzip(data)

	list.Set("/static"+path, SFile{data, gzipData, 0, int64(len(data)), int64(len(gzipData)), mime.TypeByExtension(ext), f, f.ModTime().UTC().Format(http.TimeFormat)})

	DebugLogf("Added the '%s' static file", path)
	return nil
}

func (list SFileList) Get(name string) (file SFile, exists bool) {
	staticFileMutex.RLock()
	defer staticFileMutex.RUnlock()
	file, exists = list[name]
	return file, exists
}

func (list SFileList) Set(name string, data SFile) {
	staticFileMutex.Lock()
	defer staticFileMutex.Unlock()
	list[name] = data
}

func compressBytesGzip(in []byte) []byte {
	var buff bytes.Buffer
	gz := gzip.NewWriter(&buff)
	_, _ = gz.Write(in) // TODO: What if this errors? What circumstances could it error under? Should we add a second return value?
	_ = gz.Close()
	return buff.Bytes()
}
