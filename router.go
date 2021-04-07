// Now home to the parts of gen_router.go which aren't expected to change from generation to generation
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	c "github.com/Azareal/Gosora/common"
)

type WriterIntercept struct {
	http.ResponseWriter
}

func NewWriterIntercept(w http.ResponseWriter) *WriterIntercept {
	return &WriterIntercept{w}
}

var wiMaxAge = "max-age=" + strconv.Itoa(int(c.Day))

func (wi *WriterIntercept) WriteHeader(code int) {
	if code == 200 {
		h := wi.ResponseWriter.Header()
		h.Set("Cache-Control", wiMaxAge)
		h.Set("Vary", "Accept-Encoding")
	}
	wi.ResponseWriter.WriteHeader(code)
}

type GenRouter struct {
	UploadHandler func(http.ResponseWriter, *http.Request)
	extraRoutes   map[string]func(http.ResponseWriter, *http.Request, *c.User) c.RouteError

	reqLogger *log.Logger

	reqLog2 *RouterLog
	suspLog *RouterLog

	sync.RWMutex
}

type RouterLogLog struct {
	File *os.File
	Log  *log.Logger
}
type RouterLog struct {
	FileVal atomic.Value
	LogVal  atomic.Value

	sync.RWMutex
}

func (r *GenRouter) DailyTick() error {
	currentTime := time.Now()
	rotateLog := func(l *RouterLog, name string) error {
		l.Lock()
		defer l.Unlock()

		f := l.FileVal.Load().(*os.File)
		stat, e := f.Stat()
		if e != nil {
			return nil
		}
		if (stat.Size() < int64(c.Megabyte)) && (currentTime.Sub(c.StartTime).Hours() >= (24 * 7)) {
			return nil
		}
		if e = f.Close(); e != nil {
			return e
		}

		stimestr := strconv.FormatInt(currentTime.Unix(), 10)
		f, e = os.OpenFile(c.Config.LogDir+name+stimestr+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
		if e != nil {
			return e
		}
		lval := log.New(f, "", log.LstdFlags)
		l.FileVal.Store(f)
		l.LogVal.Store(lval)
		return nil
	}

	if !c.Config.DisableSuspLog {
		err := rotateLog(r.suspLog, "reqs-susp-")
		if err != nil {
			return err
		}
	}
	return rotateLog(r.reqLog2, "reqs-")
}

func NewGenRouter(uploads http.Handler) (*GenRouter, error) {
	stimestr := strconv.FormatInt(c.StartTime.Unix(), 10)
	createLog := func(name, stimestr string) (*RouterLog, error) {
		f, err := os.OpenFile(c.Config.LogDir+name+"-"+stimestr+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
		if err != nil {
			return nil, err
		}
		l := log.New(f, "", log.LstdFlags)
		var aVal atomic.Value
		var aVal2 atomic.Value
		aVal.Store(f)
		aVal2.Store(l)
		return &RouterLog{FileVal: aVal, LogVal: aVal2}, nil
	}
	reqLog, err := createLog("reqs", stimestr)
	if err != nil {
		return nil, err
	}
	var suspReqLog *RouterLog
	if !c.Config.DisableSuspLog {
		suspReqLog, err = createLog("reqs-susp", stimestr)
		if err != nil {
			return nil, err
		}
	}
	f3, err := os.OpenFile(c.Config.LogDir+"reqs-misc-"+stimestr+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	reqMiscLog := log.New(f3, "", log.LstdFlags)

	return &GenRouter{
		UploadHandler: func(w http.ResponseWriter, r *http.Request) {
			writ := NewWriterIntercept(w)
			http.StripPrefix("/uploads/", uploads).ServeHTTP(writ, r)
		},
		extraRoutes: make(map[string]func(http.ResponseWriter, *http.Request, *c.User) c.RouteError),

		reqLogger: reqMiscLog,
		reqLog2:   reqLog,
		suspLog:   suspReqLog,
	}, nil
}

func (r *GenRouter) handleError(err c.RouteError, w http.ResponseWriter, req *http.Request, u *c.User) {
	if err.Handled() {
		return
	}
	if err.Type() == "system" {
		c.InternalErrorJSQ(err, w, req, err.JSON())
		return
	}
	c.LocalErrorJSQ(err.Error(), w, req, u, err.JSON())
}

func (r *GenRouter) Handle(_ string, _ http.Handler) {
}

func (r *GenRouter) HandleFunc(pattern string, h func(http.ResponseWriter, *http.Request, *c.User) c.RouteError) {
	r.Lock()
	defer r.Unlock()
	r.extraRoutes[pattern] = h
}

func (r *GenRouter) RemoveFunc(pattern string) error {
	r.Lock()
	defer r.Unlock()
	_, ok := r.extraRoutes[pattern]
	if !ok {
		return ErrNoRoute
	}
	delete(r.extraRoutes, pattern)
	return nil
}

func (r *GenRouter) dumpRequest(req *http.Request, pre string, log *RouterLog) {
	var sb strings.Builder
	r.ddumpRequest(req, pre, log, &sb)
}

// TODO: Some of these sanitisations may be redundant
var dumpReqLen = len("\nUA: \n Host: \nIP: \n") + 7
var dumpReqLen2 = len("\nHead : ") + 2

func (r *GenRouter) ddumpRequest(req *http.Request, pre string, l *RouterLog, sb *strings.Builder) {
	nfield := func(label, val string) {
		sb.WriteString(label)
		sb.WriteString(val)
	}
	field := func(label, val string) {
		nfield(label, c.SanitiseSingleLine(val))
	}
	ua := req.UserAgent()

	sb.Grow(dumpReqLen + len(pre) + len(ua) + len(req.Method) + len(req.Host) + (dumpReqLen2 * len(req.Header)))
	sb.WriteString(pre)
	sb.WriteString("\n")
	sb.WriteString(c.SanitiseSingleLine(req.Method))
	sb.WriteRune(' ')
	sb.WriteString(c.SanitiseSingleLine(req.URL.Path))
	field("\nUA: ", ua)

	for key, val := range req.Header {
		// Avoid logging this for security reasons
		if key == "Cookie" {
			continue
		}
		for _, vvalue := range val {
			sb.WriteString("\nHead ")
			sb.WriteString(c.SanitiseSingleLine(key))
			sb.WriteString(": ")
			sb.WriteString(c.SanitiseSingleLine(vvalue))
		}
	}
	field("\nHost: ", req.Host)
	if rawQuery := req.URL.RawQuery; rawQuery != "" {
		field("\nURL.RawQuery: ", rawQuery)
	}
	if ref := req.Referer(); ref != "" {
		field("\nRef: ", ref)
	}
	nfield("\nIP: ", req.RemoteAddr)
	sb.WriteString("\n")

	str := sb.String()
	l.RLock()
	l.LogVal.Load().(*log.Logger).Print(str)
	l.RUnlock()
}

func (r *GenRouter) DumpRequest(req *http.Request, pre string) {
	r.dumpRequest(req, pre, r.reqLog2)
}

func (r *GenRouter) unknownUA(req *http.Request) {
	if c.Dev.DebugMode {
		var presb strings.Builder
		presb.WriteString("Unknown UA: ")
		for _, ch := range req.UserAgent() {
			presb.WriteString(strconv.Itoa(int(ch)))
			presb.WriteRune(' ')
		}
		r.ddumpRequest(req, "", r.reqLog2, &presb)
	} else {
		r.reqLogger.Print("unknown ua: ", c.SanitiseSingleLine(req.UserAgent()))
	}
}

func (r *GenRouter) susp1(req *http.Request) bool {
	if !strings.Contains(req.URL.Path, ".") {
		return false
	}
	if strings.Contains(req.URL.Path, "..") /* || strings.Contains(req.URL.Path,"--")*/ {
		return true
	}
	lp := strings.ToLower(req.URL.Path)
	// TODO: Flag any requests which has a dot with anything but a number after that
	// TODO: Use HasSuffix to avoid over-scanning?
	return strings.Contains(lp, ".php") || strings.Contains(lp, ".asp") || strings.Contains(lp, ".cgi") || strings.Contains(lp, ".py") || strings.Contains(lp, ".sql") || strings.Contains(lp, ".act") //.action
}

func (r *GenRouter) suspScan(req *http.Request) {
	if c.Config.DisableSuspLog {
		if c.Dev.FullReqLog {
			r.DumpRequest(req, "")
		}
		return
	}

	// TODO: Cover more suspicious strings and at a lower layer than this
	var ch rune
	var susp bool
	for _, ch = range req.URL.Path { //char
		if ch != '&' && !(ch > 44 && ch < 58) && ch != '=' && ch != '?' && !(ch > 64 && ch < 91) && ch != '\\' && ch != '_' && !(ch > 96 && ch < 123) {
			susp = true
			break
		}
	}

	// Avoid logging the same request multiple times
	susp2 := r.susp1(req)
	if susp && susp2 {
		r.SuspiciousRequest(req, "Bad char '"+string(ch)+"' in path\nBad snippet in path")
	} else if susp {
		r.SuspiciousRequest(req, "Bad char '"+string(ch)+"' in path")
	} else if susp2 {
		r.SuspiciousRequest(req, "Bad snippet in path")
	} else if c.Dev.FullReqLog {
		r.DumpRequest(req, "")
	}
}

func isLocalHost(h string) bool {
	return h == "localhost" || h == "127.0.0.1" || h == "::1"
}

//var brPool = sync.Pool{}
var gzipPool = sync.Pool{}

//var uaBufPool = sync.Pool{}

func (r *GenRouter) responseWriter(w http.ResponseWriter) http.ResponseWriter {
	/*if bzw, ok := w.(c.BrResponseWriter); ok {
		w = bzw.ResponseWriter
		w.Header().Del("Content-Encoding")
	} else */if gzw, ok := w.(c.GzipResponseWriter); ok {
		w = gzw.ResponseWriter
		w.Header().Del("Content-Encoding")
	}
	return w
}
