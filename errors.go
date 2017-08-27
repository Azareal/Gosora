package main

import "fmt"
import "log"
import "bytes"
import "sync"
import "net/http"
import "runtime/debug"

// TO-DO: Use the error_buffer variable to construct the system log in the Control Panel. Should we log errors caused by users too? Or just collect statistics on those or do nothing? Intercept recover()? Could we intercept the logger instead here? We might get too much information, if we intercept the logger, maybe make it part of the Debug page?
var error_buffer_mutex sync.RWMutex
var error_buffer []error
//var notfound_count_per_second int
//var noperms_count_per_second int
var error_internal []byte
var error_notfound []byte

func init_errors() error {
	var b bytes.Buffer
	user := User{0,"guest","Guest","",0,false,false,false,false,false,false,GuestPerms,nil,"",false,"","","","","",0,0,"0.0.0.0.0",0}
	pi := Page{"Internal Server Error",user,hvars,tList,"A problem has occurred in the system."}
	err := templates.ExecuteTemplate(&b,"error.html", pi)
	if err != nil {
		return err
	}
	error_internal = b.Bytes()

	b.Reset()
	pi = Page{"Not Found",user,hvars,tList,"The requested page doesn't exist."}
	err = templates.ExecuteTemplate(&b,"error.html", pi)
	if err != nil {
		return err
	}
	error_notfound = b.Bytes()
	return nil
}

func LogError(err error) {
	log.Print(err)
	debug.PrintStack()
	error_buffer_mutex.Lock()
	defer error_buffer_mutex.Unlock()
	error_buffer = append(error_buffer,err)
	log.Fatal("")
}

func InternalError(err error, w http.ResponseWriter) {
	w.Write(error_internal)
	log.Print(err)
	debug.PrintStack()
	error_buffer_mutex.Lock()
	defer error_buffer_mutex.Unlock()
	error_buffer = append(error_buffer,err)
	log.Fatal("")
}

func InternalErrorJSQ(err error, w http.ResponseWriter, r *http.Request, is_js string) {
	w.WriteHeader(500)
	if is_js == "0" {
		w.Write(error_internal)
	} else {
		w.Write([]byte(`{"errmsg":"A problem has occured in the system."}`))
	}
	log.Print(err)
	debug.PrintStack()
	error_buffer_mutex.Lock()
	defer error_buffer_mutex.Unlock()
	error_buffer = append(error_buffer,err)
	log.Fatal("")
}

func InternalErrorJS(err error, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	w.Write([]byte(`{"errmsg":"A problem has occured in the system."}`))
	error_buffer_mutex.Lock()
	defer error_buffer_mutex.Unlock()
	error_buffer = append(error_buffer,err)
	log.Fatal(err)
}

func PreError(errmsg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	user := User{ID:0,Group:6,Perms:GuestPerms,}
	pi := Page{"Error",user,hvars,tList,errmsg}
	if pre_render_hooks["pre_render_error"] != nil {
		if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html",pi)
	fmt.Fprintln(w,b.String())
}

func LocalError(errmsg string, w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(500)
	pi := Page{"Local Error",user,hvars,tList,errmsg}
	if pre_render_hooks["pre_render_error"] != nil {
		if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html",pi)
	fmt.Fprintln(w,b.String())
}

func LoginRequired(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(401)
	pi := Page{"Local Error",user,hvars,tList,"You need to login to do that."}
	if pre_render_hooks["pre_render_error"] != nil {
		if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html",pi)
	fmt.Fprintln(w,b.String())
}

func PreErrorJS(errmsg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
}

func PreErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, is_js string) {
	w.WriteHeader(500)
	if is_js == "0" {
		user := User{ID:0,Group:6,Perms:GuestPerms,}
		pi := Page{"Local Error",user,hvars,tList,errmsg}
		if pre_render_hooks["pre_render_error"] != nil {
			if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
	}
}

func LocalErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, user User, is_js string) {
	w.WriteHeader(500)
	if is_js == "0" {
		pi := Page{"Local Error",user,hvars,tList,errmsg}
		if pre_render_hooks["pre_render_error"] != nil {
			if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
	}
}

func LocalErrorJS(errmsg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	w.Write([]byte(`{'errmsg': '` + errmsg + `'}`))
}

func NoPermissions(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(403)
	pi := Page{"Local Error",user,hvars,tList,"You don't have permission to do that."}
	if pre_render_hooks["pre_render_error"] != nil {
		if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	errpage := b.String()
	fmt.Fprintln(w,errpage)
}

func NoPermissionsJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	w.WriteHeader(403)
	if is_js == "0" {
		pi := Page{"Local Error",user,hvars,tList,"You don't have permission to do that."}
		if pre_render_hooks["pre_render_error"] != nil {
			if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte(`{"errmsg":"You don't have permission to do that."}`))
	}
}

func Banned(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(403)
	pi := Page{"Banned",user,hvars,tList,"You have been banned from this site."}
	if pre_render_hooks["pre_render_error"] != nil {
		if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	fmt.Fprintln(w,b.String())
}

func BannedJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	w.WriteHeader(403)
	if is_js == "0" {
		pi := Page{"Banned",user,hvars,tList,"You have been banned from this site."}
		if pre_render_hooks["pre_render_error"] != nil {
			if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte(`{"errmsg":"You have been banned from this site."}`))
	}
}

func LoginRequiredJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	w.WriteHeader(401)
	if is_js == "0" {
		pi := Page{"Local Error",user,hvars,tList,"You need to login to do that."}
		if pre_render_hooks["pre_render_error"] != nil {
			if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte(`{"errmsg":"You need to login to do that."}`))
	}
}

func SecurityError(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(403)
	pi := Page{"Security Error",user,hvars,tList,"There was a security issue with your request."}
	if pre_render_hooks["pre_render_security_error"] != nil {
		if run_pre_render_hook("pre_render_security_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	fmt.Fprintln(w,b.String())
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write(error_notfound)
}

func CustomError(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(errcode)
	pi := Page{errtitle,user,hvars,tList,errmsg}
	if pre_render_hooks["pre_render_error"] != nil {
		if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
			return
		}
	}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	fmt.Fprintln(w,b.String())
}

func CustomErrorJSQ(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, user User, is_js string) {
	w.WriteHeader(errcode)
	if is_js == "0" {
		pi := Page{errtitle,user,hvars,tList,errmsg}
		if pre_render_hooks["pre_render_error"] != nil {
			if run_pre_render_hook("pre_render_error", w, r, &user, &pi) {
				return
			}
		}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte(`{"errmsg":"` + errmsg + `"}`))
	}
}
