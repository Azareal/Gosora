package main
import "fmt"
import "log"
import "bytes"
import "net/http"

var error_internal []byte
var error_notfound []byte
func init_errors() error {
	var b bytes.Buffer
	user := User{0,"Guest","",0,false,false,false,false,false,false,GuestPerms,"",false,"","","","","",0,0,"0.0.0.0.0"}
	pi := Page{"Internal Server Error",user,nList,tList,"A problem has occurred in the system."}
	err := templates.ExecuteTemplate(&b,"error.html", pi)
	if err != nil {
		return err
	}
	error_internal = b.Bytes()
	
	b.Reset()
	pi = Page{"Not Found",user,nList,tList,"The requested page doesn't exist."}
	err = templates.ExecuteTemplate(&b,"error.html", pi)
	if err != nil {
		return err
	}
	error_notfound = b.Bytes()
	return nil
}

func InternalError(err error, w http.ResponseWriter, r *http.Request) {
	w.Write(error_internal)
	log.Fatal(err)
}

func InternalErrorJSQ(err error, w http.ResponseWriter, r *http.Request, is_js string) {
	w.WriteHeader(500)
	if is_js == "0" {
		w.Write(error_internal)
	} else {
		w.Write([]byte(`{'errmsg': 'A problem has occured in the system.'}`))
	}
	log.Fatal(err)
}

func InternalErrorJS(err error, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	w.Write([]byte(`{'errmsg': 'A problem has occured in the system.'}`))
	log.Fatal(err)
}

func PreError(errmsg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	user := User{ID:0,Group:6,Perms:GuestPerms,}
	pi := Page{"Error",user,nList,tList,errmsg}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html",pi)
	fmt.Fprintln(w,b.String())
}

func LocalError(errmsg string, w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(500)
	pi := Page{"Local Error",user,nList,tList,errmsg}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html",pi)
	fmt.Fprintln(w,b.String())
}

func LoginRequired(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(401)
	pi := Page{"Local Error",user,nList,tList,"You need to login to do that."}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html",pi)
	fmt.Fprintln(w,b.String())
}

func PreErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, is_js string) {
	w.WriteHeader(500)
	if is_js == "0" {
		user := User{ID:0,Group:6,Perms:GuestPerms,}
		pi := Page{"Local Error",user,nList,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte(`{'errmsg': '` + errmsg + `'}`))
	}
}

func LocalErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, user User, is_js string) {
	w.WriteHeader(500)
	if is_js == "0" {
		pi := Page{"Local Error",user,nList,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte(`{'errmsg': '` + errmsg + `'}`))
	}
}

func LocalErrorJS(errmsg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	w.Write([]byte(`{'errmsg': '` + errmsg + `'}`))
}

func NoPermissions(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(403)
	pi := Page{"Local Error",user,nList,tList,"You don't have permission to do that."}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	errpage := b.String()
	fmt.Fprintln(w,errpage)
}

func NoPermissionsJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	w.WriteHeader(403)
	if is_js == "0" {
		pi := Page{"Local Error",user,nList,tList,"You don't have permission to do that."}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte("{'errmsg': 'You don't have permission to do that.'}"))
	}
}

func Banned(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(403)
	pi := Page{"Banned",user,nList,tList,"You have been banned from this site."}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	fmt.Fprintln(w,b.String())
}

func BannedJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	w.WriteHeader(403)
	if is_js == "0" {
		pi := Page{"Banned",user,nList,tList,"You have been banned from this site."}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte("{'errmsg': 'You have been banned from this site.'}"))
	}
}

func LoginRequiredJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	w.WriteHeader(401)
	if is_js == "0" {
		pi := Page{"Local Error",user,nList,tList,"You need to login to do that."}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte("{'errmsg': 'You need to login to do that.'}"))
	}
}

func SecurityError(w http.ResponseWriter, r *http.Request, user User) {
	w.WriteHeader(403)
	pi := Page{"Security Error",user,nList,tList,"There was a security issue with your request."}
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
	pi := Page{errtitle,user,nList,tList,errmsg}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	fmt.Fprintln(w,b.String())
}

func CustomErrorJSQ(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, user User, is_js string) {
	w.WriteHeader(errcode)
	if is_js == "0" {
		pi := Page{errtitle,user,nList,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		fmt.Fprintln(w,b.String())
	} else {
		w.Write([]byte(`{'errmsg': '` + errmsg + `'}`))
	}
}
