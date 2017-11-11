package common

import "log"

import "sync"
import "net/http"
import "runtime/debug"

// TODO: Use the error_buffer variable to construct the system log in the Control Panel. Should we log errors caused by users too? Or just collect statistics on those or do nothing? Intercept recover()? Could we intercept the logger instead here? We might get too much information, if we intercept the logger, maybe make it part of the Debug page?
// ? - Should we pass HeaderVars / HeaderLite rather than forcing the errors to pull the global HeaderVars instance?
var errorBufferMutex sync.RWMutex
var errorBuffer []error

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

/*func NewRouteError(msg string, system bool, json bool) RouteError {
	return &RouteErrorImpl{msg, system, json, false}
}*/

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
	log.Print(err)
	debug.PrintStack()
	errorBufferMutex.Lock()
	defer errorBufferMutex.Unlock()
	errorBuffer = append(errorBuffer, err)
}

// TODO: Dump the request?
// InternalError is the main function for handling internal errors, while simultaneously printing out a page for the end-user to let them know that *something* has gone wrong
// ? - Add a user parameter?
func InternalError(err error, w http.ResponseWriter, r *http.Request) RouteError {
	log.Print(err)
	debug.PrintStack()

	// TODO: Centralise the user struct somewhere else
	user := User{0, "guest", "Guest", "", 0, false, false, false, false, false, false, GuestPerms, nil, "", false, "", "", "", "", "", 0, 0, "0.0.0.0.0", 0}
	pi := Page{"Internal Server Error", user, DefaultHeaderVar(), tList, "A problem has occurred in the system."}
	err = Templates.ExecuteTemplate(w, "error.html", pi)
	if err != nil {
		log.Print(err)
	}

	errorBufferMutex.Lock()
	defer errorBufferMutex.Unlock()
	errorBuffer = append(errorBuffer, err)
	log.Fatal("")
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
	errorBufferMutex.Lock()
	defer errorBufferMutex.Unlock()
	errorBuffer = append(errorBuffer, err)
	log.Fatal(err)
	return HandledRouteError()
}

func PreError(errmsg string, w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(500)
	user := User{ID: 0, Group: 6, Perms: GuestPerms}
	pi := Page{"Error", user, DefaultHeaderVar(), tList, errmsg}
	if PreRenderHooks["pre_render_error"] != nil {
		if RunPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return nil
		}
	}
	err := Templates.ExecuteTemplate(w, "error.html", pi)
	if err != nil {
		LogError(err)
	}
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
	pi := Page{"Local Error", user, DefaultHeaderVar(), tList, errmsg}
	if PreRenderHooks["pre_render_error"] != nil {
		if RunPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return nil
		}
	}
	err := Templates.ExecuteTemplate(w, "error.html", pi)
	if err != nil {
		LogError(err)
	}
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
	_, _ = w.Write([]byte(`{'errmsg': '` + errmsg + `'}`))
	return HandledRouteError()
}

// TODO: We might want to centralise the error logic in the future and just return what the error handler needs to construct the response rather than handling it here
// NoPermissions is an error shown to the end-user when they try to access an area which they aren't authorised to access
func NoPermissions(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(403)
	pi := Page{"Local Error", user, DefaultHeaderVar(), tList, "You don't have permission to do that."}
	// TODO: What to do about this hook?
	if PreRenderHooks["pre_render_error"] != nil {
		if RunPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return nil
		}
	}
	err := Templates.ExecuteTemplate(w, "error.html", pi)
	if err != nil {
		LogError(err)
	}
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
	pi := Page{"Banned", user, DefaultHeaderVar(), tList, "You have been banned from this site."}
	if PreRenderHooks["pre_render_error"] != nil {
		if RunPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return nil
		}
	}
	err := Templates.ExecuteTemplate(w, "error.html", pi)
	if err != nil {
		LogError(err)
	}
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
	pi := Page{"Local Error", user, DefaultHeaderVar(), tList, "You need to login to do that."}
	if PreRenderHooks["pre_render_error"] != nil {
		if RunPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return nil
		}
	}
	err := Templates.ExecuteTemplate(w, "error.html", pi)
	if err != nil {
		LogError(err)
	}
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
	pi := Page{"Security Error", user, DefaultHeaderVar(), tList, "There was a security issue with your request."}
	if PreRenderHooks["pre_render_security_error"] != nil {
		if RunPreRenderHook("pre_render_security_error", w, r, &user, &pi) {
			return nil
		}
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
func NotFound(w http.ResponseWriter, r *http.Request) RouteError {
	w.WriteHeader(404)
	// TODO: Centralise the user struct somewhere else
	user := User{0, "guest", "Guest", "", 0, false, false, false, false, false, false, GuestPerms, nil, "", false, "", "", "", "", "", 0, 0, "0.0.0.0.0", 0}
	pi := Page{"Not Found", user, DefaultHeaderVar(), tList, "The requested page doesn't exist."}
	err := Templates.ExecuteTemplate(w, "error.html", pi)
	if err != nil {
		LogError(err)
	}
	return HandledRouteError()
}

// CustomError lets us make custom error types which aren't covered by the generic functions above
func CustomError(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(errcode)
	pi := Page{errtitle, user, DefaultHeaderVar(), tList, errmsg}
	if PreRenderHooks["pre_render_error"] != nil {
		if RunPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return nil
		}
	}
	err := Templates.ExecuteTemplate(w, "error.html", pi)
	if err != nil {
		LogError(err)
	}
	return HandledRouteError()
}

// CustomErrorJSQ is a version of CustomError which lets us handle both JSON and regular pages depending on how it's being accessed
func CustomErrorJSQ(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, user User, isJs bool) RouteError {
	if !isJs {
		return CustomError(errmsg, errcode, errtitle, w, r, user)
	}
	return CustomErrorJS(errmsg, errcode, errtitle, w, r, user)
}

// CustomErrorJS is the pure JSON version of CustomError
func CustomErrorJS(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.WriteHeader(errcode)
	_, _ = w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
	return HandledRouteError()
}
