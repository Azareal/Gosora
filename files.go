package main

import (
	"log"
	"bytes"
	"strings"
	"mime"
	//"errors"
	"os"
	"io/ioutil"
	"path/filepath"
	"net/http"
	"compress/gzip"
)

type SFile struct
{
	Data []byte
	GzipData []byte
	Pos int64
	Length int64
	GzipLength int64
	Mimetype string
	Info os.FileInfo
	FormattedModTime string
}

type CssData struct
{
	ComingSoon string
}

func init_static_files() {
	log.Print("Loading the static files.")
	err := filepath.Walk("./public", func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}

		path = strings.Replace(path,"\\","/",-1)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		path = strings.TrimPrefix(path,"public/")
		var ext string = filepath.Ext("/public/" + path)
		gzip_data := compress_bytes_gzip(data)

		static_files["/static/" + path] = SFile{data,gzip_data,0,int64(len(data)),int64(len(gzip_data)),mime.TypeByExtension(ext),f,f.ModTime().UTC().Format(http.TimeFormat)}

		if debug_mode {
			log.Print("Added the '" + path + "' static file.")
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
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

	var ext string = filepath.Ext(path)
	path = strings.TrimPrefix(path, prefix)
	gzip_data := compress_bytes_gzip(data)

	static_files["/static" + path] = SFile{data,gzip_data,0,int64(len(data)),int64(len(gzip_data)),mime.TypeByExtension(ext),f,f.ModTime().UTC().Format(http.TimeFormat)}

	if debug_mode {
		log.Print("Added the '" + path + "' static file")
	}
	return nil
}

func compress_bytes_gzip(in []byte) []byte {
	var buff bytes.Buffer
	gz := gzip.NewWriter(&buff)
	gz.Write(in)
	gz.Close()
	return buff.Bytes()
}
