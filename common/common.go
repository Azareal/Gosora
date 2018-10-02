package common

import (
	"database/sql"
	"log"
	"sync/atomic"
	"time"

	"../query_gen/lib"
)

var SoftwareVersion = Version{Major: 0, Minor: 2, Patch: 0, Tag: "dev"}

// nolint I don't want to write comments for each of these o.o
const Hour int = 60 * 60
const Day int = Hour * 24
const Week int = Day * 7
const Month int = Day * 30
const Year int = Day * 365
const Kilobyte int = 1024
const Megabyte int = Kilobyte * 1024
const Gigabyte int = Megabyte * 1024
const Terabyte int = Gigabyte * 1024
const Petabyte int = Terabyte * 1024

var StartTime time.Time
var TmplPtrMap = make(map[string]interface{})

// Anti-spam token with rotated key
var JSTokenBox atomic.Value              // TODO: Move this and some of these other globals somewhere else
var SessionSigningKeyBox atomic.Value    // For MFA to avoid hitting the database unneccesarily
var OldSessionSigningKeyBox atomic.Value // Just in case we've signed with a key that's about to go stale so we don't annoy the user too much
var IsDBDown int32 = 0                   // 0 = false, 1 = true. this is value which should be manipulated with package atomic for representing whether the database is down so we don't spam the log with lots of redundant errors

// ErrNoRows is an alias of sql.ErrNoRows, just in case we end up with non-database/sql datastores
var ErrNoRows = sql.ErrNoRows

// ? - Make this more customisable?
var ExternalSites = map[string]string{
	"YT": "https://www.youtube.com/",
}

type StringList []string

// ? - Should we allow users to upload .php or .go files? It could cause security issues. We could store them with a mangled extension to render them inert
// TODO: Let admins manage this from the Control Panel
// apng is commented out for now, as we have no way of re-encoding it into a smaller file
var AllowedFileExts = StringList{
	"png", "jpg", "jpeg", "svg", "bmp", "gif", "tif", "webp", /*"apng",*/ // images

	"txt", "xml", "json", "yaml", "toml", "ini", "md", "html", "rtf", "js", "py", "rb", "css", "scss", "less", "eqcss", "pcss", "java", "ts", "cs", "c", "cc", "cpp", "cxx", "C", "c++", "h", "hh", "hpp", "hxx", "h++", "rs", "rlib", "htaccess", "gitignore", /*"go","php",*/ // text

	"mp3", "mp4", "avi", "wmv", "webm", // video

	"otf", "woff2", "woff", "ttf", "eot", // fonts
}
var ImageFileExts = StringList{
	"png", "jpg", "jpeg", "svg", "bmp", "gif", "tif", "webp", /* "apng",*/
}
var ArchiveFileExts = StringList{
	"bz2", "zip", "gz", "7z", "tar", "cab",
}
var ExecutableFileExts = StringList{
	"exe", "jar", "phar", "shar", "iso",
}

func init() {
	JSTokenBox.Store("")
	SessionSigningKeyBox.Store("")
	OldSessionSigningKeyBox.Store("")
}

// TODO: Write a test for this
func (slice StringList) Contains(needle string) bool {
	for _, item := range slice {
		if item == needle {
			return true
		}
	}
	return false
}

type dbInits []func(acc *qgen.Accumulator) error

var DbInits dbInits

func (inits dbInits) Run() error {
	for _, init := range inits {
		err := init(qgen.NewAcc())
		if err != nil {
			return err
		}
	}
	return nil
}

func (inits dbInits) Add(init ...func(acc *qgen.Accumulator) error) {
	DbInits = dbInits(append(DbInits, init...))
}

// TODO: Add a graceful shutdown function
func StoppedServer(msg ...interface{}) {
	//log.Print("stopped server")
	StopServerChan <- msg
}

var StopServerChan = make(chan []interface{})

func DebugDetail(args ...interface{}) {
	if Dev.SuperDebug {
		log.Print(args...)
	}
}

func DebugDetailf(str string, args ...interface{}) {
	if Dev.SuperDebug {
		log.Printf(str, args...)
	}
}

func DebugLog(args ...interface{}) {
	if Dev.DebugMode {
		log.Print(args...)
	}
}

func DebugLogf(str string, args ...interface{}) {
	if Dev.DebugMode {
		log.Printf(str, args...)
	}
}
