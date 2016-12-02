package main
import "log"
import "bytes"
import "net/http"

func InternalError(err error, w http.ResponseWriter, r *http.Request, user User) {
	log.Fatal(err)
	pi := Page{"Internal Server Error","error",user,tList,"A problem has occured in the system."}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	errpage := b.String()
	http.Error(w,errpage,500)
}

func InternalErrorJSQ(err error, w http.ResponseWriter, r *http.Request, user User, is_js string) {
	log.Fatal(err)
	errmsg := "A problem has occured in the system."
	if is_js == "0" {
		pi := Page{"Internal Server Error","error",user,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		http.Error(w,errpage,500)
	} else {
		http.Error(w,"{'errmsg': '" + errmsg + "'}",500)
	}
}

func LocalError(errmsg string, w http.ResponseWriter, r *http.Request, user User) {
	pi := Page{"Local Error","error",user,tList,errmsg}
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	errpage := b.String()
	http.Error(w,errpage,500)
}

func LoginRequired(w http.ResponseWriter, r *http.Request, user User) {
	errmsg := "You need to login to do that."
	pi := Page{"Local Error","error",user,tList,errmsg}
	
	var b bytes.Buffer
	templates.ExecuteTemplate(&b,"error.html", pi)
	errpage := b.String()
	http.Error(w,errpage,401)
}

func LocalErrorJSQ(errmsg string, w http.ResponseWriter, r *http.Request, user User, is_js string) {
	if is_js == "0" {
		pi := Page{"Local Error","error",user,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		http.Error(w,errpage,500)
	} else {
		http.Error(w,"{'errmsg': '" + errmsg + "'}",500)
	}
}

func NoPermissionsJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	errmsg := "You don't have permission to do that."
	if is_js == "0" {
		pi := Page{"Local Error","error",user,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		http.Error(w,errpage,403)
	} else {
		http.Error(w,"{'errmsg': '" + errmsg + "'}",403)
	}
}

func LoginRequiredJSQ(w http.ResponseWriter, r *http.Request, user User, is_js string) {
	errmsg := "You need to login to do that."
	if is_js == "0" {
		pi := Page{"Local Error","error",user,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		http.Error(w,errpage,401)
	} else {
		http.Error(w,"{'errmsg': '" + errmsg + "'}",401)
	}
}

func CustomErrorJSQ(errmsg string, errcode int, errtitle string, w http.ResponseWriter, r *http.Request, user User, is_js string) {
	if is_js == "0" {
		pi := Page{errtitle,"error",user,tList,errmsg}
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		http.Error(w,errpage,500)
	} else {
		http.Error(w,"{'errmsg': '" + errmsg + "'}",500)
	}
}
