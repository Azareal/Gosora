package main
import "fmt"
import "log"
import "bytes"
import "net/http"

func InternalError(err error, w http.ResponseWriter, r *http.Request, user User) {
	log.Fatal(err)
	pi := Page{"Internal Server Error","error",user,tList,"A problem has occured in the system."}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	errpage := b.String()
	w.WriteHeader(500)
	fmt.Fprintln(w,errpage)
}

func InternalErrorJSQ(err error, w http.ResponseWriter, r *http.Request, user User, is_js string) {
	log.Fatal(err)
	errmsg := "A problem has occured in the system."
	if is_js == "0" {
		pi := Page{"Internal Server Error","error",user,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
	} else {
		http.Error(w,"{'errmsg': '" + errmsg + "'}",500)
	}
}

func LocalError(errmsg string, w http.ResponseWriter, r *http.Request, user User) {
	pi := Page{"Local Error","error",user,tList,errmsg}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	errpage := b.String()
	w.WriteHeader(500)
	fmt.Fprintln(w,errpage)
}

func LoginRequired(w http.ResponseWriter, r *http.Request, user User) {
	pi := Page{"Local Error","error",user,tList,"You need to login to do that."}
	
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	errpage := b.String()
	w.WriteHeader(401)
	fmt.Fprintln(w,errpage)
}

func LocalErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, user User, is_js string) {
	if is_js == "0" {
		pi := Page{"Local Error","error",user,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
	} else {
		http.Error(w,"{'errmsg': '" + errmsg + "'}",500)
	}
}

func NoPermissions(w http.ResponseWriter, r *http.Request, user User) {
	errmsg := "You don't have permission to do that."
	pi := Page{"Local Error","error",user,tList,errmsg}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	errpage := b.String()
	w.WriteHeader(403)
	fmt.Fprintln(w,errpage)
}

func NoPermissionsJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	errmsg := "You don't have permission to do that."
	if is_js == "0" {
		pi := Page{"Local Error","error",user,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(403)
		fmt.Fprintln(w,errpage)
	} else {
		http.Error(w,"{'errmsg': '" + errmsg + "'}",403)
	}
}

func Banned(w http.ResponseWriter, r *http.Request, user User) {
	pi := Page{"Banned","error",user,tList,"You have been banned from this site."}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	errpage := b.String()
	w.WriteHeader(403)
	fmt.Fprintln(w,errpage)
}

func BannedJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	if is_js == "0" {
		pi := Page{"Banned","error",user,tList,"You have been banned from this site."}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(403)
		fmt.Fprintln(w,errpage)
	} else {
		http.Error(w,"{'errmsg': 'You have been banned from this site.'}",403)
	}
}

func LoginRequiredJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	if is_js == "0" {
		pi := Page{"Local Error","error",user,tList,"You need to login to do that."}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(401)
		fmt.Fprintln(w,errpage)
	} else {
		http.Error(w,"{'errmsg': 'You need to login to do that.'}",401)
	}
}

func CustomErrorJSQ(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, user User, is_js string) {
	if is_js == "0" {
		pi := Page{errtitle,"error",user,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
	} else {
		http.Error(w,"{'errmsg': '" + errmsg + "'}",500)
	}
}
