package main

import "fmt"
import "log"
import "bytes"
import "sync"
import "net/http"
import "runtime/debug"

// TODO: Use the error_buffer variable to construct the system log in the Control Panel. Should we log errors caused by users too? Or just collect statistics on those or do nothing? Intercept recover()? Could we intercept the logger instead here? We might get too much information, if we intercept the logger, maybe make it part of the Debug page?
var errorBufferMutex sync.RWMutex
var errorBuffer []error

//var notfoundCountPerSecond int
//var nopermsCountPerSecond int
var errorInternal []byte
var errorNotfound []byte

func initErrors() error {
	var b bytes.Buffer
	user := User{0, "guest", "Guest", "", 0, false, false, false, false, false, false, GuestPerms, nil, "", false, "", "", "", "", "", 0, 0, "0.0.0.0.0", 0}
	pi := Page{"Internal Server Error", user, hvars, tList, "A problem has occurred in the system."}
	err := templates.ExecuteTemplate(&b, "error.html", pi)
	if err != nil {
		return err
	}
	errorInternal = b.Bytes()

	b.Reset()
	pi = Page{"Not Found", user, hvars, tList, "The requested page doesn't exist."}
	err = templates.ExecuteTemplate(&b, "error.html", pi)
	if err != nil {
		return err
	}
	errorNotfound = b.Bytes()
	return nil
}

// LogError logs internal handler errors which can't be handled with InternalError() as a wrapper for log.Fatal(), we might do more with it in the future
func LogError(err error) {
	log.Print(err)
	debug.PrintStack()
	errorBufferMutex.Lock()
	defer errorBufferMutex.Unlock()
	errorBuffer = append(errorBuffer, err)
	log.Fatal("")
}

// InternalError is the main function for handling internal errors, while simultaneously printing out a page for the end-user to let them know that *something* has gone wrong
func InternalError(err error, w http.ResponseWriter) {
	_, _ = w.Write(errorInternal)
	log.Print(err)
	debug.PrintStack()
	errorBufferMutex.Lock()
	defer errorBufferMutex.Unlock()
	errorBuffer = append(errorBuffer, err)
	log.Fatal("")
}

// InternalErrorJSQ is the JSON "maybe" version of InternalError which can handle both JSON and normal requests
func InternalErrorJSQ(err error, w http.ResponseWriter, r *http.Request, isJs bool) {
	w.WriteHeader(500)
	if !isJs {
		_, _ = w.Write(errorInternal)
	} else {
		_, _ = w.Write([]byte(`{"errmsg":"A problem has occured in the system."}`))
	}
	log.Print(err)
	debug.PrintStack()
	errorBufferMutex.Lock()
	defer errorBufferMutex.Unlock()
	errorBuffer = append(errorBuffer, err)
	log.Fatal("")
}

// InternalErrorJS is the JSON version of InternalError on routes we know will only be requested via JSON. E.g. An API.
func InternalErrorJS(err error, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	_, _ = w.Write([]byte(`{"errmsg":"A problem has occured in the system."}`))
	errorBufferMutex.Lock()
	defer errorBufferMutex.Unlock()
	errorBuffer = append(errorBuffer, err)
	log.Fatal(err)
}

// ? - Where is this used? Is it actually used? Should we use it more?
// LoginRequired is an error shown to the end-user when they try to access an area which requires them to login
func LoginRequired(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(401)
	pi := Page{"Local Error", user, hvars, tList, "You need to login to do that."}
	if preRenderHooks["pre_render_error"] != nil {
		if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	err := templates.ExecuteTemplate(&b, "error.html", pi)
	if err != nil {
		LogError(err)
	}
	fmt.Fprintln(w, b.String())
}

func PreError(errmsg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	user := User{ID: 0, Group: 6, Perms: GuestPerms}
	pi := Page{"Error", user, hvars, tList, errmsg}
	if preRenderHooks["pre_render_error"] != nil {
		if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	err := templates.ExecuteTemplate(&b, "error.html", pi)
	if err != nil {
		LogError(err)
	}
	fmt.Fprintln(w, b.String())
}

func PreErrorJS(errmsg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	_, _ = w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
}

func PreErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, isJs bool) {
	w.WriteHeader(500)
	if !isJs {
		user := User{ID: 0, Group: 6, Perms: GuestPerms}
		pi := Page{"Local Error", user, hvars, tList, errmsg}
		if preRenderHooks["pre_render_error"] != nil {
			if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		err := templates.ExecuteTemplate(&b, "error.html", pi)
		if err != nil {
			LogError(err)
		}
		fmt.Fprintln(w, b.String())
	} else {
		_, _ = w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
	}
}

// LocalError is an error shown to the end-user when something goes wrong and it's not the software's fault
func LocalError(errmsg string, w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(500)
	pi := Page{"Local Error", user, hvars, tList, errmsg}
	if preRenderHooks["pre_render_error"] != nil {
		if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	err := templates.ExecuteTemplate(&b, "error.html", pi)
	if err != nil {
		LogError(err)
	}
	fmt.Fprintln(w, b.String())
}

func LocalErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, user User, isJs bool) {
	w.WriteHeader(500)
	if !isJs {
		pi := Page{"Local Error", user, hvars, tList, errmsg}
		if preRenderHooks["pre_render_error"] != nil {
			if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		err := templates.ExecuteTemplate(&b, "error.html", pi)
		if err != nil {
			LogError(err)
		}
		fmt.Fprintln(w, b.String())
	} else {
		_, _ = w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
	}
}

func LocalErrorJS(errmsg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	_, _ = w.Write([]byte(`{'errmsg': '` + errmsg + `'}`))
}

// NoPermissions is an error shown to the end-user when they try to access an area which they aren't authorised to access
func NoPermissions(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(403)
	pi := Page{"Local Error", user, hvars, tList, "You don't have permission to do that."}
	if preRenderHooks["pre_render_error"] != nil {
		if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	err := templates.ExecuteTemplate(&b, "error.html", pi)
	if err != nil {
		LogError(err)
	}
	errpage := b.String()
	fmt.Fprintln(w, errpage)
}

func NoPermissionsJSQ(w http.ResponseWriter, r *http.Request, user User, isJs bool) {
	w.WriteHeader(403)
	if !isJs {
		pi := Page{"Local Error", user, hvars, tList, "You don't have permission to do that."}
		if preRenderHooks["pre_render_error"] != nil {
			if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		err := templates.ExecuteTemplate(&b, "error.html", pi)
		if err != nil {
			LogError(err)
		}
		fmt.Fprintln(w, b.String())
	} else {
		_, _ = w.Write([]byte(`{"errmsg":"You don't have permission to do that."}`))
	}
}

// ? - Is this actually used? Should it be used? A ban in Gosora should be more of a permission revocation to stop them posting rather than something which spits up an error page, right?
func Banned(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(403)
	pi := Page{"Banned", user, hvars, tList, "You have been banned from this site."}
	if preRenderHooks["pre_render_error"] != nil {
		if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	err := templates.ExecuteTemplate(&b, "error.html", pi)
	if err != nil {
		LogError(err)
	}
	fmt.Fprintln(w, b.String())
}

// nolint
// BannedJSQ is the version of the banned error page which handles both JavaScript requests and normal page loads
func BannedJSQ(w http.ResponseWriter, r *http.Request, user User, isJs bool) {
	w.WriteHeader(403)
	if !isJs {
		pi := Page{"Banned", user, hvars, tList, "You have been banned from this site."}
		if preRenderHooks["pre_render_error"] != nil {
			if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		err := templates.ExecuteTemplate(&b, "error.html", pi)
		if err != nil {
			LogError(err)
		}
		fmt.Fprintln(w, b.String())
	} else {
		_, _ = w.Write([]byte(`{"errmsg":"You have been banned from this site."}`))
	}
}

// nolint
func LoginRequiredJSQ(w http.ResponseWriter, r *http.Request, user User, isJs bool) {
	w.WriteHeader(401)
	if !isJs {
		pi := Page{"Local Error", user, hvars, tList, "You need to login to do that."}
		if preRenderHooks["pre_render_error"] != nil {
			if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		err := templates.ExecuteTemplate(&b, "error.html", pi)
		if err != nil {
			LogError(err)
		}
		fmt.Fprintln(w, b.String())
	} else {
		_, _ = w.Write([]byte(`{"errmsg":"You need to login to do that."}`))
	}
}

func SecurityError(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(403)
	pi := Page{"Security Error", user, hvars, tList, "There was a security issue with your request."}
	if preRenderHooks["pre_render_security_error"] != nil {
		if runPreRenderHook("pre_render_security_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	err := templates.ExecuteTemplate(&b, "error.html", pi)
	if err != nil {
		LogError(err)
	}
	fmt.Fprintln(w, b.String())
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	_, _ = w.Write(errorNotfound)
}

// nolint
func CustomError(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(errcode)
	pi := Page{errtitle, user, hvars, tList, errmsg}
	if preRenderHooks["pre_render_error"] != nil {
		if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	err := templates.ExecuteTemplate(&b, "error.html", pi)
	if err != nil {
		LogError(err)
	}
	fmt.Fprintln(w, b.String())
}

// nolint
func CustomErrorJSQ(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, user User, isJs bool) {
	w.WriteHeader(errcode)
	if !isJs {
		pi := Page{errtitle, user, hvars, tList, errmsg}
		if preRenderHooks["pre_render_error"] != nil {
			if runPreRenderHook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		err := templates.ExecuteTemplate(&b, "error.html", pi)
		if err != nil {
			LogError(err)
		}
		fmt.Fprintln(w, b.String())
	} else {
		_, _ = w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
	}
}
