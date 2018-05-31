package common

import (
	"log"
	"net/http"
	"runtime/debug"
	"sync"
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
	JSON() bool
	Handled() bool
}

type RouteErrorImpl struct {
	text    string
	system  bool
	json    bool
	handled bool
}

func (err *RouteErrorImpl) Type() string {
	// System errors may contain sensitive information we don't want the user to see
	if err.system {
		return "system"
	}
	return "user"
}

func (err *RouteErrorImpl) Error() string {
	return err.text
}

// Respond with JSON?
func (err *RouteErrorImpl) JSON() bool {
	return err.json
}

// Has this error been dealt with elsewhere?
func (err *RouteErrorImpl) Handled() bool {
	return err.handled
}

func HandledRouteError() RouteError {
	return &RouteErrorImpl{"", false, false, true}
}

// LogError logs internal handler errors which can't be handled with InternalError() as a wrapper for log.Fatal(), we might do more with it in the future.
func LogError(err error) {
	LogWarning(err)
	log.Fatal("")
}

func LogWarning(err error) {
	stack := debug.Stack()
	log.Print(err.Error()+"\n", string(stack))
	errorBufferMutex.Lock()
	defer errorBufferMutex.Unlock()
	errorBuffer = append(errorBuffer, ErrorItem{err, stack})
}

// TODO: Dump the request?
// InternalError is the main function for handling internal errors, while simultaneously printing out a page for the end-user to let them know that *something* has gone wrong
// ? - Add a user parameter?
func InternalError(err error, w http.ResponseWriter, r *http.Request) RouteError {
	pi := Page{"Internal Server Error", GuestUser, DefaultHeader(w), tList, "A problem has occurred in the system."}
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
	_, _ = w.Write([]byte(`{"errmsg":"A problem has occurred in the system."}`))
	LogError(err)
	return HandledRouteError()
}

var xmlInternalError = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<error>A problem has occured</error>`)

func InternalErrorXML(err error, w http.ResponseWriter, r *http.Request) RouteError {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(500)
	w.Write(xmlInternalError)
	LogError(err)
	return HandledRouteError()
}

// TODO: Stop killing the instance upon hitting an error with InternalError* and deprecate this
func SilentInternalErrorXML(err error, w http.ResponseWriter, r *http.Request) RouteError {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(500)
	w.Write(xmlInternalError)
	log.Print("InternalError: ", err)
	return HandledRouteError()
}

func PreError(errmsg string, w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(500)
	pi := Page{"Error", GuestUser, DefaultHeader(w), tList, errmsg}
	handleErrorTemplate(w, r, pi)
	return HandledRouteError()
}

func PreErrorJS(errmsg string, w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(500)
	_, _ = w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
	return HandledRouteError()
}

func PreErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, isJs bool) RouteError {
	if !isJs {
		return PreError(errmsg, w, r)
	}
	return PreErrorJS(errmsg, w, r)
}

// LocalError is an error shown to the end-user when something goes wrong and it's not the software's fault
func LocalError(errmsg string, w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(500)
	pi := Page{"Local Error", user, DefaultHeader(w), tList, errmsg}
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
	_, _ = w.Write([]byte(`{"errmsg": "` + errmsg + `"}`))
	return HandledRouteError()
}

// TODO: We might want to centralise the error logic in the future and just return what the error handler needs to construct the response rather than handling it here
// NoPermissions is an error shown to the end-user when they try to access an area which they aren't authorised to access
func NoPermissions(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(403)
	pi := Page{"Local Error", user, DefaultHeader(w), tList, "You don't have permission to do that."}
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
	_, _ = w.Write([]byte(`{"errmsg":"You don't have permission to do that."}`))
	return HandledRouteError()
}

// ? - Is this actually used? Should it be used? A ban in Gosora should be more of a permission revocation to stop them posting rather than something which spits up an error page, right?
func Banned(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(403)
	pi := Page{"Banned", user, DefaultHeader(w), tList, "You have been banned from this site."}
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
	_, _ = w.Write([]byte(`{"errmsg":"You have been banned from this site."}`))
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
	pi := Page{"Local Error", user, DefaultHeader(w), tList, "You need to login to do that."}
	handleErrorTemplate(w, r, pi)
	return HandledRouteError()
}

// nolint
func LoginRequiredJS(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(401)
	_, _ = w.Write([]byte(`{"errmsg":"You need to login to do that."}`))
	return HandledRouteError()
}

// SecurityError is used whenever a session mismatch is found
// ? - Should we add JS and JSQ versions of this?
func SecurityError(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(403)
	pi := Page{"Security Error", user, DefaultHeader(w), tList, "There was a security issue with your request."}
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
// ? - Add a JSQ and JS version of this?
// ? - Add a user parameter?
func NotFound(w http.ResponseWriter, r *http.Request, header *Header) RouteError {
	return CustomError("The requested page doesn't exist.", 404, "Not Found", w, r, header, GuestUser)
}

// CustomError lets us make custom error types which aren't covered by the generic functions above
func CustomError(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, header *Header, user User) RouteError {
	if header == nil {
		header = DefaultHeader(w)
	}
	w.WriteHeader(errcode)
	pi := Page{errtitle, user, header, tList, errmsg}
	handleErrorTemplate(w, r, pi)
	return HandledRouteError()
}

// CustomErrorJSQ is a version of CustomError which lets us handle both JSON and regular pages depending on how it's being accessed
func CustomErrorJSQ(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, header *Header, user User, isJs bool) RouteError {
	if !isJs {
		if header == nil {
			header = DefaultHeader(w)
		}
		return CustomError(errmsg, errcode, errtitle, w, r, header, user)
	}
	return CustomErrorJS(errmsg, errcode, w, r, user)
}

// CustomErrorJS is the pure JSON version of CustomError
func CustomErrorJS(errmsg string, errcode int, w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(errcode)
	_, _ = w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
	return HandledRouteError()
}

func handleErrorTemplate(w http.ResponseWriter, r *http.Request, pi Page) {
	// TODO: What to do about this hook?
	if RunPreRenderHook("pre_render_error", w, r, &pi.Header.CurrentUser, &pi) {
		return
	}
	err := RunThemeTemplate(pi.Header.Theme.Name, "error", pi, w)
	if err != nil {
		LogError(err)
	}
}
