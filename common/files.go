package common

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	tmpl "github.com/Azareal/Gosora/tmpl_client"
)

type SFileList map[string]SFile

var StaticFiles SFileList = make(map[string]SFile)
var staticFileMutex sync.RWMutex

type SFile struct {
	Data             []byte
	GzipData         []byte
	Sha256           string
	OName            string
	Pos              int64
	Length           int64
	StrLength        string
	GzipLength       int64
	StrGzipLength    string
	Mimetype         string
	Info             os.FileInfo
	FormattedModTime string
}

type CSSData struct {
	Phrases map[string]string
}

func (list SFileList) JSTmplInit() error {
	DebugLog("Initialising the client side templates")
	return filepath.Walk("./tmpl_client", func(path string, f os.FileInfo, err error) error {
		if f.IsDir() || strings.HasSuffix(path, "template_list.go") || strings.HasSuffix(path, "stub.go") {
			return nil
		}
		path = strings.Replace(path, "\\", "/", -1)
		DebugLog("Processing client template " + path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		path = strings.TrimPrefix(path, "tmpl_client/")
		tmplName := strings.TrimSuffix(path, ".jgo")
		shortName := strings.TrimPrefix(tmplName, "template_")

		replace := func(data []byte, replaceThis, withThis string) []byte {
			return bytes.Replace(data, []byte(replaceThis), []byte(withThis), -1)
		}
		rep := func(replaceThis, withThis string) {
			data = replace(data, replaceThis, withThis)
		}

		startIndex, hasFunc := skipAllUntilCharsExist(data, 0, []byte("if(tmplInits===undefined)"))
		if !hasFunc {
			return errors.New("no init map found")
		}
		data = data[startIndex-len([]byte("if(tmplInits===undefined)")):]
		rep("// nolint", "")
		rep("func ", "function ")
		rep(" error {\n", " {\nlet o = \"\"\n")
		funcIndex, hasFunc := skipAllUntilCharsExist(data, 0, []byte("function Template_"))
		if !hasFunc {
			return errors.New("no template function found")
		}
		spaceIndex, hasSpace := skipUntilIfExists(data, funcIndex, ' ')
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
		rep(string(data[spaceIndex:endBrace]), "")
		rep("))\n", "  \n")
		endBrace -= preLen - len(data) // Offset it as we've deleted portions
		fmt.Println("new endBrace: ", endBrace)
		fmt.Println("data: ", string(data))

		/*showPos := func(data []byte, index int) (out string) {
			out = "["
			for j, char := range data {
				if index == j {
					out += "[" + string(char) + "] "
				} else {
					out += string(char) + " "
				}
			}
			return out + "]"
		}*/

		// ? Can we just use a regex? I'm thinking of going more efficient, or just outright rolling wasm, this is a temp hack in a place where performance doesn't particularly matter
		each := func(phrase string, h func(index int)) {
			//fmt.Println("find each '" + phrase + "'")
			index := endBrace
			if index < 0 {
				panic("index under zero: " + strconv.Itoa(index))
			}
			var foundIt bool
			for {
				//fmt.Println("in index: ", index)
				//fmt.Println("pos: ", showPos(data, index))
				index, foundIt = skipAllUntilCharsExist(data, index, []byte(phrase))
				if !foundIt {
					break
				}
				h(index)
			}
		}
		each("strconv.Itoa(", func(index int) {
			braceAt, hasEndBrace := skipUntilIfExistsOrLine(data, index, ')')
			if hasEndBrace {
				data[braceAt] = ' ' // Blank it
			}
		})
		each("[]byte(", func(index int) {
			braceAt, hasEndBrace := skipUntilIfExistsOrLine(data, index, ')')
			if hasEndBrace {
				data[braceAt] = ' ' // Blank it
			}
		})
		each("StringToBytes(", func(index int) {
			braceAt, hasEndBrace := skipUntilIfExistsOrLine(data, index, ')')
			if hasEndBrace {
				data[braceAt] = ' ' // Blank it
			}
		})
		each("w.Write(", func(index int) {
			braceAt, hasEndBrace := skipUntilIfExistsOrLine(data, index, ')')
			if hasEndBrace {
				data[braceAt] = ' ' // Blank it
			}
		})
		each("RelativeTime(", func(index int) {
			braceAt, _ := skipUntilIfExistsOrLine(data, index, 10)
			if data[braceAt-1] == ' ' {
				data[braceAt-1] = ' ' // Blank it
			}
		})
		each("if ", func(index int) {
			//fmt.Println("if index: ", index)
			braceAt, hasBrace := skipUntilIfExistsOrLine(data, index, '{')
			if hasBrace {
				if data[braceAt-1] != ' ' {
					panic("couldn't find space before brace, found ' " + string(data[braceAt-1]) + "' instead")
				}
				data[braceAt-1] = ')' // Drop a brace here to satisfy JS
			}
		})
		each("for _, item := range ", func(index int) {
			//fmt.Println("for index: ", index)
			braceAt, hasBrace := skipUntilIfExists(data, index, '{')
			if hasBrace {
				if data[braceAt-1] != ' ' {
					panic("couldn't find space before brace, found ' " + string(data[braceAt-1]) + "' instead")
				}
				data[braceAt-1] = ')' // Drop a brace here to satisfy JS
			}
		})
		rep("for _, item := range ", "for(item of ")
		rep("w.Write([]byte(", "o += ")
		rep("w.Write(StringToBytes(", "o += ")
		rep("w.Write(", "o += ")
		rep("+= c.", "+= ")
		rep("strconv.Itoa(", "")
		rep("strconv.FormatInt(", "")
		rep("	c.", "")
		rep("phrases.", "")
		rep(", 10;", "")

		//rep("var plist = GetTmplPhrasesBytes("+shortName+"_tmpl_phrase_id)", "const plist = tmplPhrases[\""+tmplName+"\"];")
		//rep("//var plist = GetTmplPhrasesBytes("+shortName+"_tmpl_phrase_id)", "const "+shortName+"_phrase_arr = tmplPhrases[\""+tmplName+"\"];")
		rep("//var plist = GetTmplPhrasesBytes("+shortName+"_tmpl_phrase_id)", "const pl=tmplPhrases[\""+tmplName+"\"];")
		rep(shortName+"_phrase_arr", "pl")
		rep("tmpl_"+shortName+"_vars", "t_vars")

		rep("var c_v_", "let c_v_")
		rep(`t_vars, ok := tmpl_i.`, `/*`)
		rep("[]byte(", "")
		rep("StringToBytes(", "")
		rep("RelativeTime(t_vars.", "t_vars.Relative")
		// TODO: Format dates properly on the client side
		rep(".Format(\"2006-01-02 15:04:05\"", "")
		rep(", 10", "")
		rep("if ", "if(")
		rep("return nil", "return o")
		rep(" )", ")")
		rep(" \n", "\n")
		rep("\n", ";\n")
		rep("{;", "{")
		rep("};", "}")
		rep("[;", "[")
		rep(";;", ";")
		rep(",;", ",")
		rep("=;", "=")
		rep(`,
	});
}`, "\n\t];")
		rep(`=
}`, "=[]")
		rep("o += ", "o+=")

		fragset := tmpl.GetFrag(shortName)
		if fragset != nil {
			sfrags := []byte("let " + shortName + "_frags=[\n")
			for _, frags := range fragset {
				//sfrags = append(sfrags, []byte(shortName+"_frags.push(`"+string(frags)+"`);\n")...)
				sfrags = append(sfrags, []byte("`"+string(frags)+"`,\n")...)
			}
			sfrags = append(sfrags, []byte("];\n")...)
			data = append(sfrags, data...)
		}
		rep("\n;", "\n")

		for name, _ := range Themes {
			if strings.HasSuffix(shortName, "_"+name) {
				data = append(data, "\nvar Template_"+strings.TrimSuffix(shortName, "_"+name)+"=Template_"+shortName+";"...)
				break
			}
		}

		path = tmplName + ".js"
		DebugLog("js path: ", path)
		ext := filepath.Ext("/tmpl_client/" + path)
		gzipData, err := CompressBytesGzip(data)
		if err != nil {
			return err
		}
		// Don't use Gzip if we get meagre gains from it as it takes longer to process the responses
		if len(gzipData) >= (len(data) + 120) {
			gzipData = nil
		} else {
			diff := len(data) - len(gzipData)
			if diff <= len(data)/100 {
				gzipData = nil
			}
		}

		// Get a checksum for CSPs and cache busting
		hasher := sha256.New()
		hasher.Write(data)
		checksum := hex.EncodeToString(hasher.Sum(nil))

		list.Set("/s/"+path, SFile{data, gzipData, checksum, path + "?h=" + checksum, 0, int64(len(data)), strconv.Itoa(len(data)), int64(len(gzipData)), strconv.Itoa(len(gzipData)), mime.TypeByExtension(ext), f, f.ModTime().UTC().Format(http.TimeFormat)})

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
		ext := filepath.Ext("/public/" + path)
		mimetype := mime.TypeByExtension(ext)

		// Get a checksum for CSPs and cache busting
		hasher := sha256.New()
		hasher.Write(data)
		checksum := hex.EncodeToString(hasher.Sum(nil))

		// Avoid double-compressing images
		var gzipData []byte
		if mimetype != "image/jpeg" && mimetype != "image/png" && mimetype != "image/gif" {
			gzipData, err = CompressBytesGzip(data)
			if err != nil {
				return err
			}
			// Don't use Gzip if we get meagre gains from it as it takes longer to process the responses
			if len(gzipData) >= (len(data) + 150) {
				gzipData = nil
			} else {
				diff := len(data) - len(gzipData)
				if diff <= len(data)/100 {
					gzipData = nil
				}
			}
		}

		list.Set("/s/"+path, SFile{data, gzipData, checksum, path + "?h=" + checksum, 0, int64(len(data)), strconv.Itoa(len(data)), int64(len(gzipData)), strconv.Itoa(len(gzipData)), mimetype, f, f.ModTime().UTC().Format(http.TimeFormat)})

		DebugLogf("Added the '%s' static file.", path)
		return nil
	})
}

func (list SFileList) Add(path, prefix string) error {
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

	ext := filepath.Ext(path)
	path = strings.TrimPrefix(path, prefix)
	gzipData, err := CompressBytesGzip(data)
	if err != nil {
		return err
	}
	// Don't use Gzip if we get meagre gains from it as it takes longer to process the responses
	if len(gzipData) >= (len(data) + 150) {
		gzipData = nil
	} else {
		diff := len(data) - len(gzipData)
		if diff <= len(data)/100 {
			gzipData = nil
		}
	}

	// Get a checksum for CSPs and cache busting
	hasher := sha256.New()
	hasher.Write(data)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	list.Set("/s"+path, SFile{data, gzipData, checksum, path + "?h=" + checksum, 0, int64(len(data)), strconv.Itoa(len(data)), int64(len(gzipData)), strconv.Itoa(len(gzipData)), mime.TypeByExtension(ext), f, f.ModTime().UTC().Format(http.TimeFormat)})

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

func CompressBytesGzip(in []byte) ([]byte, error) {
	var buff bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buff, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = gz.Write(in)
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}
