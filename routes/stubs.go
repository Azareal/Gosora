package routes

import "net/http"

// HTTPSRedirect is a connection handler which redirects all HTTP requests to HTTPS
type HTTPSRedirect struct {
}

func (red *HTTPSRedirect) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	dest := "https://" + req.Host + req.URL.String()
	http.Redirect(w, req, dest, http.StatusTemporaryRedirect)
}

// Temporary stubs for view tracking
func DynamicRoute() {
}
func UploadedFile() {
}
func BadRoute() {
}
