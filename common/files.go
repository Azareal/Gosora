package common

import (
	"bytes"
	"errors"
	"fmt"
	"mime"
	"strings"
	"sync"
	//"errors"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"../tmpl_client"
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
	Phrases map[string]string
}

func (list SFileList) JSTmplInit() error {
	var fragMap = make(map[string][][]byte)
	fragMap["alert"] = tmpl.GetFrag("alert")
	fmt.Println("fragMap: ", fragMap)
	return filepath.Walk("./tmpl_client", func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "template_list.go") || strings.HasSuffix(path, "stub.go") {
			return nil
		}

		path = strings.Replace(path, "\\", "/", -1)
		DebugLog("Processing client template " + path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var replace = func(data []byte, replaceThis string, withThis string) []byte {
			return bytes.Replace(data, []byte(replaceThis), []byte(withThis), -1)
		}

		startIndex, hasFunc := skipAllUntilCharsExist(data, 0, []byte("func Template"))
		if !hasFunc {
			return errors.New("no template function found")
		}
		data = data[startIndex-len([]byte("func Template")):]
		data = replace(data, "func ", "function ")
		data = replace(data, " error {\n", " {\nlet out = \"\"\n")
		spaceIndex, hasSpace := skipUntilIfExists(data, 10, ' ')
		if !hasSpace {
			return errors.New("no spaces found after the template function name")
		}
		endBrace, hasBrace := skipUntilIfExists(data, spaceIndex, ')')
		if !hasBrace {
			return errors.New("no right brace found after the template function name")
		}
		fmt.Println("spaceIndex: ", spaceIndex)
		fmt.Println("endBrace: ", endBrace)
		fmt.Println("string(data[spaceIndex:endBrace]): ", string(data[spaceIndex:endBrace]))
		preLen := len(data)
		data = replace(data, string(data[spaceIndex:endBrace]), "")
		data = replace(data, "))\n", "\n")
		endBrace -= preLen - len(data) // Offset it as we've deleted portions

		var showPos = func(data []byte, index int) (out string) {
			out = "["
			for j, char := range data {
				if index == j {
					out += "[" + string(char) + "] "
				} else {
					out += string(char) + " "
				}
			}
			return out + "]"
		}

		// ? Can we just use a regex? I'm thinking of going more efficient, or just outright rolling wasm, this is a temp hack in a place where performance doesn't particularly matter
		var each = func(phrase string, handle func(index int)) {
			fmt.Println("find each '" + phrase + "'")
			var index = endBrace
			var foundIt bool
			for {
				fmt.Println("in index: ", index)
				fmt.Println("pos: ", showPos(data, index))
				index, foundIt = skipAllUntilCharsExist(data, index, []byte(phrase))
				if !foundIt {
					break
				}
				handle(index)
			}
		}
		each("strconv.Itoa(", func(index int) {
			braceAt, hasEndBrace := skipUntilIfExists(data, index, ')')
			// TODO: Make sure we don't go onto the next line in case someone misplaced a brace
			if hasEndBrace {
				data[braceAt] = ' ' // Blank it
			}
		})
		each("w.Write([]byte(", func(index int) {
			braceAt, hasEndBrace := skipUntilIfExists(data, index, ')')
			// TODO: Make sure we don't go onto the next line in case someone misplaced a brace
			if hasEndBrace {
				data[braceAt] = ' ' // Blank it
			}
			braceAt, hasEndBrace = skipUntilIfExists(data, braceAt, ')')
			if hasEndBrace {
				data[braceAt] = ' ' // Blank this one too
			}
		})
		each("w.Write(", func(index int) {
			braceAt, hasEndBrace := skipUntilIfExists(data, index, ')')
			// TODO: Make sure we don't go onto the next line in case someone misplaced a brace
			if hasEndBrace {
				data[braceAt] = ' ' // Blank it
			}
		})
		each("if ", func(index int) {
			fmt.Println("if index: ", index)
			braceAt, hasBrace := skipUntilIfExists(data, index, '{')
			if hasBrace {
				if data[braceAt-1] != ' ' {
					panic("couldn't find space before brace, found ' " + string(data[braceAt-1]) + "' instead")
				}
				data[braceAt-1] = ')' // Drop a brace here to satisfy JS
			}
		})
		data = replace(data, "w.Write([]byte(", "out += ")
		data = replace(data, "w.Write(", "out += ")
		data = replace(data, "strconv.Itoa(", "")
		data = replace(data, "if ", "if(")
		data = replace(data, "return nil", "return out")
		data = replace(data, " )", ")")
		data = replace(data, " \n", "\n")
		data = replace(data, "\n", ";\n")
		data = replace(data, "{;", "{")
		data = replace(data, "};", "}")
		data = replace(data, ";;", ";")

		path = strings.TrimPrefix(path, "tmpl_client/")
		tmplName := strings.TrimSuffix(path, ".go")
		fragset, ok := fragMap[strings.TrimPrefix(tmplName, "template_")]
		if !ok {
			fmt.Println("tmplName: ", tmplName)
			return errors.New("couldn't find template in fragmap")
		}

		var sfrags = []byte("let alert_frags = [];\n")
		for _, frags := range fragset {
			sfrags = append(sfrags, []byte("alert_frags.push(`"+string(frags)+"`);\n")...)
		}
		data = append(sfrags, data...)
		data = replace(data, "\n;", "\n")

		path = tmplName + ".js"
		DebugLog("js path: ", path)
		var ext = filepath.Ext("/tmpl_client/" + path)
		gzipData := compressBytesGzip(data)

		list.Set("/static/"+path, SFile{data, gzipData, 0, int64(len(data)), int64(len(gzipData)), mime.TypeByExtension(ext), f, f.ModTime().UTC().Format(http.TimeFormat)})

		DebugLogf("Added the '%s' static file.", path)
		return nil
	})
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
