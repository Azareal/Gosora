/*
*
*	Gosora Common Resources
*	Copyright Azareal 2018 - 2020
*
 */
package common // import "github.com/Azareal/Gosora/common"

import (
	"database/sql"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	meta "github.com/Azareal/Gosora/common/meta"
	qgen "github.com/Azareal/Gosora/query_gen"
)

var SoftwareVersion = Version{Major: 0, Minor: 3, Patch: 0, Tag: "dev"}

var Meta meta.MetaStore

// nolint I don't want to write comments for each of these o.o
const Hour int = 60 * 60
const Day = Hour * 24
const Week = Day * 7
const Month = Day * 30
const Year = Day * 365
const Kilobyte int = 1024
const Megabyte = Kilobyte * 1024
const Gigabyte = Megabyte * 1024
const Terabyte = Gigabyte * 1024
const Petabyte = Terabyte * 1024

var StartTime time.Time
var GzipStartEtag string
var StartEtag string
var TmplPtrMap = make(map[string]interface{})

// Anti-spam token with rotated key
var JSTokenBox atomic.Value              // TODO: Move this and some of these other globals somewhere else
var SessionSigningKeyBox atomic.Value    // For MFA to avoid hitting the database unneccessarily
var OldSessionSigningKeyBox atomic.Value // Just in case we've signed with a key that's about to go stale so we don't annoy the user too much
var IsDBDown int32 = 0                   // 0 = false, 1 = true. this is value which should be manipulated with package atomic for representing whether the database is down so we don't spam the log with lots of redundant errors

// ErrNoRows is an alias of sql.ErrNoRows, just in case we end up with non-database/sql datastores
var ErrNoRows = sql.ErrNoRows

//var StrSlicePool sync.Pool

// ? - Make this more customisable?
/*var ExternalSites = map[string]string{
	"YT": "https://www.youtube.com/",
}*/

// TODO: Make this more customisable
var SpammyDomainBits = []string{"porn", "sex", "acup", "nude", "milf", "tits", "vape", "busty", "kink", "lingerie", "strapon", "problog", "fet", "xblog", "blogin", "blognetwork", "relayblog"}

var Chrome, Firefox int // ! Temporary Hack for http push
var SimpleBots []int    // ! Temporary hack to stop semrush, ahrefs, python bots and other from wasting resources

type StringList []string

// ? - Should we allow users to upload .php or .go files? It could cause security issues. We could store them with a mangled extension to render them inert
// TODO: Let admins manage this from the Control Panel
// apng is commented out for now, as we have no way of re-encoding it into a smaller file
var AllowedFileExts = StringList{
	"png", "jpg", "jpe", "jpeg", "jif", "jfi", "jfif", "svg", "bmp", "gif", "tiff", "tif", "webp", "apng", "avif", "flif", "heif", "heic", "bpg", // images (encodable) + apng (browser support) + bpg + avif + flif + heif / heic

	"txt", "xml", "json", "yaml", "toml", "ini", "md", "html", "rtf", "js", "py", "rb", "css", "scss", "less", "eqcss", "pcss", "java", "ts", "cs", "c", "cc", "cpp", "cxx", "C", "c++", "h", "hh", "hpp", "hxx", "h++", "rs", "rlib", "htaccess", "gitignore", /*"go","php",*/ // text

	"wav", "mp3", "oga", "m4a", "flac", "ac3", "aac", "opus", // audio

	"mp4", "avi", "ogg", "ogv", "ogx", "wmv", "webm", "flv", "f4v", "xvid", "mov", "movie", "qt", // video

	"otf", "woff2", "woff", "ttf", "eot", // fonts

	"bz2", "zip", "zipx", "gz", "7z", "tar", "cab", "rar", "kgb", "pea", "xz", "zz", "tgz", "xpi", // archives

	"docx", "pdf", // documents
}
var ImageFileExts = StringList{
	"png", "jpg", "jpe", "jpeg", "jif", "jfi", "jfif", "svg", "bmp", "gif", "tiff", "tif", "webp", /* "apng", "bpg", "avif", */
}
var TextFileExts = StringList{
	"txt", "xml", "json", "yaml", "toml", "ini", "md", "html", "rtf", "js", "py", "rb", "css", "scss", "less", "eqcss", "pcss", "java", "ts", "cs", "c", "cc", "cpp", "cxx", "C", "c++", "h", "hh", "hpp", "hxx", "h++", "rs", "rlib", "htaccess", "gitignore", /*"go","php",*/
}
var VideoFileExts = StringList{
	"mp4", "avi", "ogg", "ogv", "ogx", "wmv", "webm", "flv", "f4v", "xvid", "mov", "movie", "qt",
}
var WebVideoFileExts = StringList{
	"mp4", "avi", "ogg", "ogv", "webm",
}
var WebAudioFileExts = StringList{
	"wav", "mp3", "oga", "m4a", "flac",
}
var ArchiveFileExts = StringList{
	"bz2", "zip", "zipx", "gz", "7z", "tar", "cab", "rar", "kgb", "pea", "xz", "zz", "tgz", "xpi",
}
var ExecutableFileExts = StringList{
	"exe", "jar", "phar", "shar", "iso", "apk", "deb",
}

func init() {
	JSTokenBox.Store("")
	SessionSigningKeyBox.Store("")
	OldSessionSigningKeyBox.Store("")
}

// TODO: Write a test for this
func (sl StringList) Contains(needle string) bool {
	for _, it := range sl {
		if it == needle {
			return true
		}
	}
	return false
}

/*var DbTables []string
var TableToID = make(map[string]int)
var IDToTable = make(map[int]string)

func InitTables(acc *qgen.Accumulator) error {
	stmt := acc.Select("tables").Columns("id,name").Prepare()
	if e := acc.FirstError(); e != nil {
		return e
	}
	return eachall(stmt, func(r *sql.Rows) error {
		var id int
		var name string
		if e := r.Scan(&id, &name); e != nil {
			return e
		}
		TableToID[name] = id
		IDToTable[id] = name
		return nil
	})
}*/

type dbInits []func(acc *qgen.Accumulator) error

var DbInits dbInits

func (inits dbInits) Run() error {
	for _, i := range inits {
		if e := i(qgen.NewAcc()); e != nil {
			return e
		}
	}
	return nil
}

func (inits dbInits) Add(i ...func(acc *qgen.Accumulator) error) {
	DbInits = dbInits(append(DbInits, i...))
}

// TODO: Add a graceful shutdown function
func StoppedServer(msg ...interface{}) {
	//log.Print("stopped server")
	StopServerChan <- msg
}

var StopServerChan = make(chan []interface{})

var LogWriter = io.MultiWriter(os.Stderr)

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

func Log(args ...interface{}) {
	log.Print(args...)
}

func Logf(str string, args ...interface{}) {
	log.Printf(str, args...)
}

func Countf(stmt *sql.Stmt, args ...interface{}) (count int) {
	err := stmt.QueryRow(args...).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func eachall(stmt *sql.Stmt, f func(r *sql.Rows) error) error {
	rows, e := stmt.Query()
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		if e := f(rows); e != nil {
			return e
		}
	}
	return rows.Err()
}

var qcache = []string{0: "?", 1: "?,?", 2: "?,?,?", 3: "?,?,?,?", 4: "?,?,?,?,?", 5: "?,?,?,?,?,?", 6: "?,?,?,?,?,?,?", 7: "?,?,?,?,?,?,?,?"}

func inqbuild(ids []int) ([]interface{}, string) {
	if len(ids) < 7 {
		idList := make([]interface{}, len(ids))
		for i, id := range ids {
			idList[i] = strconv.Itoa(id)
		}
		return idList, qcache[len(ids)-1]
	}

	var sb strings.Builder
	sb.Grow((len(ids) * 2) - 1)
	idList := make([]interface{}, len(ids))
	for i, id := range ids {
		idList[i] = strconv.Itoa(id)
		if i == 0 {
			sb.WriteRune('?')
		} else {
			sb.WriteString(",?")
		}
	}
	return idList, sb.String()
}

func inqbuild2(count int) string {
	if count <= 7 {
		return qcache[count-1]
	}
	var sb strings.Builder
	sb.Grow((count * 2) - 1)
	for i := 0; i <= count; i++ {
		if i == 0 {
			sb.WriteRune('?')
		} else {
			sb.WriteString(",?")
		}
	}
	return sb.String()
}

func inqbuildstr(strs []string) ([]interface{}, string) {
	if len(strs) < 7 {
		idList := make([]interface{}, len(strs))
		for i, id := range strs {
			idList[i] = id
		}
		return idList, qcache[len(strs)-1]
	}

	var sb strings.Builder
	sb.Grow((len(strs) * 2) - 1)
	idList := make([]interface{}, len(strs))
	for i, id := range strs {
		idList[i] = id
		if i == 0 {
			sb.WriteRune('?')
		} else {
			sb.WriteString(",?")
		}
	}
	return idList, sb.String()
}
