package routes

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"../common"
)

var cacheControlMaxAge = "max-age=" + strconv.Itoa(common.Day) // TODO: Make this a common.Config value

// GET functions
func StaticFile(w http.ResponseWriter, r *http.Request) {
	file, ok := common.StaticFiles.Get(r.URL.Path)
	if !ok {
		common.DebugLogf("Failed to find '%s'", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	h := w.Header()

	// Surely, there's a more efficient way of doing this?
	t, err := time.Parse(http.TimeFormat, h.Get("If-Modified-Since"))
	if err == nil && file.Info.ModTime().Before(t.Add(1*time.Second)) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	h.Set("Last-Modified", file.FormattedModTime)
	h.Set("Content-Type", file.Mimetype)
	h.Set("Cache-Control", cacheControlMaxAge) //Cache-Control: max-age=31536000
	h.Set("Vary", "Accept-Encoding")
	if strings.Contains(h.Get("Accept-Encoding"), "gzip") {
		h.Set("Content-Encoding", "gzip")
		h.Set("Content-Length", strconv.FormatInt(file.GzipLength, 10))
		io.Copy(w, bytes.NewReader(file.GzipData)) // Use w.Write instead?
	} else {
		h.Set("Content-Length", strconv.FormatInt(file.Length, 10)) // Avoid doing a type conversion every time?
		io.Copy(w, bytes.NewReader(file.Data))
	}
	// Other options instead of io.Copy: io.CopyN(), w.Write(), http.ServeContent()
}

func Overview(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Zone = "overview"

	pi := common.Page{common.GetTitlePhrase("overview"), user, headerVars, tList, nil}
	if common.RunPreRenderHook("pre_render_overview", w, r, &user, &pi) {
		return nil
	}
	err := common.Templates.ExecuteTemplate(w, "overview.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func CustomPage(w http.ResponseWriter, r *http.Request, user common.User, name string) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Zone = "custom_page"

	// ! Is this safe?
	if common.Templates.Lookup("page_"+name+".html") == nil {
		return common.NotFound(w, r, headerVars)
	}

	pi := common.Page{common.GetTitlePhrase("page"), user, headerVars, tList, nil}
	// TODO: Pass the page name to the pre-render hook?
	if common.RunPreRenderHook("pre_render_custom_page", w, r, &user, &pi) {
		return nil
	}

	err := common.Templates.ExecuteTemplate(w, "page_"+name+".html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}
