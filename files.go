package main

import (
	"bytes"
	"log"
	"mime"
	"strings"
	//"errors"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

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

type CssData struct {
	ComingSoon string
}

func initStaticFiles() error {
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

		staticFiles["/static/"+path] = SFile{data, gzipData, 0, int64(len(data)), int64(len(gzipData)), mime.TypeByExtension(ext), f, f.ModTime().UTC().Format(http.TimeFormat)}

		if dev.DebugMode {
			log.Print("Added the '" + path + "' static file.")
		}
		return nil
	})
}

func addStaticFile(path string, prefix string) error {
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

	staticFiles["/static"+path] = SFile{data, gzipData, 0, int64(len(data)), int64(len(gzipData)), mime.TypeByExtension(ext), f, f.ModTime().UTC().Format(http.TimeFormat)}

	if dev.DebugMode {
		log.Print("Added the '" + path + "' static file")
	}
	return nil
}

func compressBytesGzip(in []byte) []byte {
	var buff bytes.Buffer
	gz := gzip.NewWriter(&buff)
	_, _ = gz.Write(in) // TO-DO: What if this errors? What circumstances could it error under? Should we add a second return value?
	_ = gz.Close()
	return buff.Bytes()
}
