package common

import (
	"database/sql"
)

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

const SaltLength int = 32
const SessionLength int = 80

var TmplPtrMap = make(map[string]interface{})

// ErrNoRows is an alias of sql.ErrNoRows, just in case we end up with non-database/sql datastores
var ErrNoRows = sql.ErrNoRows

// ? - Make this more customisable?
var ExternalSites = map[string]string{
	"YT": "https://www.youtube.com/",
}

type StringList []string

// ? - Should we allow users to upload .php or .go files? It could cause security issues. We could store them with a mangled extension to render them inert
// TODO: Let admins manage this from the Control Panel
var AllowedFileExts = StringList{
	"png", "jpg", "jpeg", "svg", "bmp", "gif", "tif", "webp", "apng", // images

	"txt", "xml", "json", "yaml", "toml", "ini", "md", "html", "rtf", "js", "py", "rb", "css", "scss", "less", "eqcss", "pcss", "java", "ts", "cs", "c", "cc", "cpp", "cxx", "C", "c++", "h", "hh", "hpp", "hxx", "h++", "rs", "rlib", "htaccess", "gitignore", // text

	"mp3", "mp4", "avi", "wmv", "webm", // video

	"otf", "woff2", "woff", "ttf", "eot", // fonts
}
var ImageFileExts = StringList{
	"png", "jpg", "jpeg", "svg", "bmp", "gif", "tif", "webp", "apng",
}
var ArchiveFileExts = StringList{
	"bz2", "zip", "gz", "7z", "tar", "cab",
}
var ExecutableFileExts = StringList{
	"exe", "jar", "phar", "shar", "iso",
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

type dbInits []func() error

var DbInits dbInits

func (inits dbInits) Run() error {
	for _, init := range inits {
		err := init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (inits dbInits) Add(init ...func() error) {
	DbInits = dbInits(append(DbInits, init...))
}
