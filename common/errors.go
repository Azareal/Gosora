package common

import (
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/Azareal/Gosora/common/phrases"
)

type ErrorItem struct {
	error
	Stack []byte
}

// ! The errorBuffer uses o(n) memory, we should probably do something about that
// TODO: Use the errorBuffer variable to construct the system log in the Control Panel. Should we log errors caused by users too? Or just collect statistics on those or do nothing? Intercept recover()? Could we intercept the logger instead here? We might get too much information, if we intercept the logger, maybe make it part of the Debug page?
// ? - Should we pass Header / HeaderLite rather than forcing the errors to pull the global Header instance?
var errorBufferMutex sync.RWMutex
var errorBuffer []ErrorItem

//var notfoundCountPerSecond int
//var nopermsCountPerSecond int

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

// WIP, a new system to propagate errors up from routes
type RouteError interface {
	Type() string
	Error() string
	Cause() string
	JSON() bool
	Handled() bool

	Wrap(string)
}

type RouteErrorImpl struct {
	userText string
	sysText  string
	system   bool
	json     bool
	handled  bool
}

func (err *RouteErrorImpl) Type() string {
	// System errors may contain sensitive information we don't want the user to see
	if err.system {
		return "system"
	}
	return "user"
}

func (err *RouteErrorImpl) Error() string {
	return err.userText
}

func (err *RouteErrorImpl) Cause() string {
	if err.sysText == "" {
		return err.Error()
	}
	return err.sysText
}

// Respond with JSON?
func (err *RouteErrorImpl) JSON() bool {
	return err.json
}

// Has this error been dealt with elsewhere?
func (err *RouteErrorImpl) Handled() bool {
	return err.handled
}

// Move the current error into the system error slot and add a new one to the user error slot to show the user
func (err *RouteErrorImpl) Wrap(userErr string) {
	err.sysText = err.userText
	err.userText = userErr
}

func HandledRouteError() RouteError {
	return &RouteErrorImpl{"", "", false, false, true}
}

func Error(errmsg string) RouteError {
	return &RouteErrorImpl{errmsg, "", false, false, false}
}

func FromError(err error) RouteError {
	return &RouteErrorImpl{err.Error(), "", false, false, false}
}

func ErrorJSQ(errmsg string, isJs bool) RouteError {
	return &RouteErrorImpl{errmsg, "", false, isJs, false}
}

func SysError(errmsg string) RouteError {
	return &RouteErrorImpl{errmsg, errmsg, true, false, false}
}

// LogError logs internal handler errors which can't be handled with InternalError() as a wrapper for log.Fatal(), we might do more with it in the future.
// TODO: Clean-up extra as a way of passing additional context
func LogError(err error, extra ...string) {
	LogWarning(err, extra...)
	log.Fatal("")
}

func LogWarning(err error, extra ...string) {
	var errmsg string
	for _, extraBit := range extra {
		errmsg += extraBit + "\n"
	}
	if err == nil {
		errmsg += "Unknown error"
	} else {
		errmsg += err.Error()
	}
	stack := debug.Stack()
	log.Print(errmsg+"\n", string(stack))
	errorBufferMutex.Lock()
	defer errorBufferMutex.Unlock()
	errorBuffer = append(errorBuffer, ErrorItem{err, stack})
}

func errorHeader(w http.ResponseWriter, user User, title string) *Header {
	header := DefaultHeader(w, user)
	header.Title = title
	return header
}

// TODO: Dump the request?
// InternalError is the main function for handling internal errors, while simultaneously printing out a page for the end-user to let them know that *something* has gone wrong
// ? - Add a user parameter?
func InternalError(err error, w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(500)
	pi := ErrorPage{errorHeader(w, GuestUser, phrases.GetErrorPhrase("internal_error_title")), phrases.GetErrorPhrase("internal_error_body")}
	handleErrorTemplate(w, r, pi)
	LogError(err)
	return HandledRouteError()
}

// InternalErrorJSQ is the JSON "maybe" version of InternalError which can handle both JSON and normal requests
// ? - Add a user parameter?
func InternalErrorJSQ(err error, w http.ResponseWriter, r *http.Request, isJs bool) RouteError {
	if !isJs {
		return InternalError(err, w, r)
	}
	return InternalErrorJS(err, w, r)
}

// InternalErrorJS is the JSON version of InternalError on routes we know will only be requested via JSON. E.g. An API.
// ? - Add a user parameter?
func InternalErrorJS(err error, w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(500)
	writeJsonError(phrases.GetErrorPhrase("internal_error_body"), w)
	LogError(err)
	return HandledRouteError()
}

// When the task system detects if the database is down, some database errors might lip by this
func DatabaseError(w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(500)
	pi := ErrorPage{errorHeader(w, GuestUser, phrases.GetErrorPhrase("internal_error_title")), phrases.GetErrorPhrase("internal_error_body")}
	handleErrorTemplate(w, r, pi)
	return HandledRouteError()
}

func InternalErrorXML(err error, w http.ResponseWriter, r *http.Request) RouteError {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(500)
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<error>` + phrases.GetErrorPhrase("internal_error_body") + `</error>`))
	LogError(err)
	return HandledRouteError()
}

// TODO: Stop killing the instance upon hitting an error with InternalError* and deprecate this
func SilentInternalErrorXML(err error, w http.ResponseWriter, r *http.Request) RouteError {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(500)
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<error>` + phrases.GetErrorPhrase("internal_error_body") + `</error>`))
	log.Print("InternalError: ", err)
	return HandledRouteError()
}

func PreError(errmsg string, w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(500)
	pi := ErrorPage{errorHeader(w, GuestUser, phrases.GetErrorPhrase("error_title")), errmsg}
	handleErrorTemplate(w, r, pi)
	return HandledRouteError()
}

func PreErrorJS(errmsg string, w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(500)
	writeJsonError(errmsg, w)
	return HandledRouteError()
}

func PreErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, isJs bool) RouteError {
	if !isJs {
		return PreError(errmsg, w, r)
	}
	return PreErrorJS(errmsg, w, r)
}

// LocalError is an error shown to the end-user when something goes wrong and it's not the software's fault
// TODO: Pass header in for this and similar errors instead of having to pass in both user and w? Would also allow for more stateful things, although this could be a problem
func LocalError(errmsg string, w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(500)
	pi := ErrorPage{errorHeader(w, user, phrases.GetErrorPhrase("local_error_title")), errmsg}
	handleErrorTemplate(w, r, pi)
	return HandledRouteError()
}

func LocalErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, user User, isJs bool) RouteError {
	if !isJs {
		return LocalError(errmsg, w, r, user)
	}
	return LocalErrorJS(errmsg, w, r)
}

func LocalErrorJS(errmsg string, w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(500)
	writeJsonError(errmsg, w)
	return HandledRouteError()
}

// TODO: We might want to centralise the error logic in the future and just return what the error handler needs to construct the response rather than handling it here
// NoPermissions is an error shown to the end-user when they try to access an area which they aren't authorised to access
func NoPermissions(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(403)
	pi := ErrorPage{errorHeader(w, user, phrases.GetErrorPhrase("no_permissions_title")), phrases.GetErrorPhrase("no_permissions_body")}
	handleErrorTemplate(w, r, pi)
	return HandledRouteError()
}

func NoPermissionsJSQ(w http.ResponseWriter, r *http.Request, user User, isJs bool) RouteError {
	if !isJs {
		return NoPermissions(w, r, user)
	}
	return NoPermissionsJS(w, r, user)
}

func NoPermissionsJS(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(403)
	writeJsonError(phrases.GetErrorPhrase("no_permissions_body"), w)
	return HandledRouteError()
}

// ? - Is this actually used? Should it be used? A ban in Gosora should be more of a permission revocation to stop them posting rather than something which spits up an error page, right?
func Banned(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(403)
	pi := ErrorPage{errorHeader(w, user, phrases.GetErrorPhrase("banned_title")), phrases.GetErrorPhrase("banned_body")}
	handleErrorTemplate(w, r, pi)
	return HandledRouteError()
}

// nolint
// BannedJSQ is the version of the banned error page which handles both JavaScript requests and normal page loads
func BannedJSQ(w http.ResponseWriter, r *http.Request, user User, isJs bool) RouteError {
	if !isJs {
		return Banned(w, r, user)
	}
	return BannedJS(w, r, user)
}

func BannedJS(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(403)
	writeJsonError(phrases.GetErrorPhrase("banned_body"), w)
	return HandledRouteError()
}

// nolint
func LoginRequiredJSQ(w http.ResponseWriter, r *http.Request, user User, isJs bool) RouteError {
	if !isJs {
		return LoginRequired(w, r, user)
	}
	return LoginRequiredJS(w, r, user)
}

// ? - Where is this used? Should we use it more?
// LoginRequired is an error shown to the end-user when they try to access an area which requires them to login
func LoginRequired(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(401)
	pi := ErrorPage{errorHeader(w, user, phrases.GetErrorPhrase("no_permissions_title")), phrases.GetErrorPhrase("login_required_body")}
	handleErrorTemplate(w, r, pi)
	return HandledRouteError()
}

// nolint
func LoginRequiredJS(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(401)
	writeJsonError(phrases.GetErrorPhrase("login_required_body"), w)
	return HandledRouteError()
}

// SecurityError is used whenever a session mismatch is found
// ? - Should we add JS and JSQ versions of this?
func SecurityError(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(403)
	pi := ErrorPage{errorHeader(w, user, phrases.GetErrorPhrase("security_error_title")), phrases.GetErrorPhrase("security_error_body")}
	if RunPreRenderHook("pre_render_security_error", w, r, &user, &pi) {
		return nil
	}
	err := Templates.ExecuteTemplate(w, "error.html", pi)
	if err != nil {
		LogError(err)
	}
	return HandledRouteError()
}

// NotFound is used when the requested page doesn't exist
// ? - Add a JSQ version of this?
// ? - Add a user parameter?
func NotFound(w http.ResponseWriter, r *http.Request, header *Header) RouteError {
	return CustomError(phrases.GetErrorPhrase("not_found_body"), 404, phrases.GetErrorPhrase("not_found_title"), w, r, header, GuestUser)
}

// ? - Add a user parameter?
func NotFoundJS(w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(401)
	writeJsonError(phrases.GetErrorPhrase("not_found_body"), w)
	return HandledRouteError()
}

func NotFoundJSQ(w http.ResponseWriter, r *http.Request, header *Header, js bool) RouteError {
	if js {
		return NotFoundJS(w, r)
	}
	if header == nil {
		header = DefaultHeader(w, GuestUser)
	}
	return NotFound(w, r, header)
}

// CustomError lets us make custom error types which aren't covered by the generic functions above
func CustomError(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, header *Header, user User) RouteError {
	if header == nil {
		header = DefaultHeader(w, user)
	}
	header.Title = errtitle
	w.WriteHeader(errcode)
	pi := ErrorPage{header, errmsg}
	handleErrorTemplate(w, r, pi)
	return HandledRouteError()
}

// CustomErrorJSQ is a version of CustomError which lets us handle both JSON and regular pages depending on how it's being accessed
func CustomErrorJSQ(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, header *Header, user User, isJs bool) RouteError {
	if !isJs {
		return CustomError(errmsg, errcode, errtitle, w, r, header, user)
	}
	return CustomErrorJS(errmsg, errcode, w, r, user)
}

// CustomErrorJS is the pure JSON version of CustomError
func CustomErrorJS(errmsg string, errcode int, w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(errcode)
	writeJsonError(errmsg, w)
	return HandledRouteError()
}

// TODO: Should we optimise this by caching these json strings?
func writeJsonError(errmsg string, w http.ResponseWriter) {
	_, _ = w.Write([]byte(`{"errmsg":"` + strings.Replace(errmsg, "\"", "", -1) + `"}`))
}

func handleErrorTemplate(w http.ResponseWriter, r *http.Request, pi ErrorPage) {
	// TODO: What to do about this hook?
	if RunPreRenderHook("pre_render_error", w, r, &pi.Header.CurrentUser, &pi) {
		return
	}
	err := pi.Header.Theme.RunTmpl("error", pi, w)
	if err != nil {
		LogError(err)
	}
}
