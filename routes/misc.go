package routes

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

var cacheControlMaxAge = "max-age=" + strconv.Itoa(int(common.Day)) // TODO: Make this a common.Config value

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
	t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since"))
	if err == nil && file.Info.ModTime().Before(t.Add(1*time.Second)) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	h.Set("Last-Modified", file.FormattedModTime)
	h.Set("Content-Type", file.Mimetype)
	h.Set("Cache-Control", cacheControlMaxAge) //Cache-Control: max-age=31536000
	h.Set("Vary", "Accept-Encoding")

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && file.GzipLength > 0 {
		h.Set("Content-Encoding", "gzip")
		h.Set("Content-Length", strconv.FormatInt(file.GzipLength, 10))
		io.Copy(w, bytes.NewReader(file.GzipData)) // Use w.Write instead?
	} else {
		h.Set("Content-Length", strconv.FormatInt(file.Length, 10)) // Avoid doing a type conversion every time?
		io.Copy(w, bytes.NewReader(file.Data))
	}
	// Other options instead of io.Copy: io.CopyN(), w.Write(), http.ServeContent()
}

func Overview(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	header.Title = phrases.GetTitlePhrase("overview")
	header.Zone = "overview"
	pi := common.Page{header, tList, nil}
	return renderTemplate("overview", w, r, header, pi)
}

func CustomPage(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header, name string) common.RouteError {
	header.Zone = "custom_page"
	name = common.SanitiseSingleLine(name)
	page, err := common.Pages.GetByName(name)
	if err == nil {
		header.Title = page.Title
		pi := common.CustomPagePage{header, page}
		return renderTemplate("custom_page", w, r, header, pi)
	} else if err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}

	// ! Is this safe?
	if common.DefaultTemplates.Lookup("page_"+name+".html") == nil {
		return common.NotFound(w, r, header)
	}

	header.Title = phrases.GetTitlePhrase("page")
	pi := common.Page{header, tList, nil}
	// TODO: Pass the page name to the pre-render hook?
	if common.RunPreRenderHook("pre_render_tmpl_page", w, r, &user, &pi) {
		return nil
	}
	err = header.Theme.RunTmpl("page_"+name, pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

// TODO: Set the cookie domain
func ChangeTheme(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	//headerLite, _ := SimpleUserCheck(w, r, &user)
	// TODO: Rename isJs to something else, just in case we rewrite the JS side in WebAssembly?
	isJs := (r.PostFormValue("isJs") == "1")
	newTheme := common.SanitiseSingleLine(r.PostFormValue("newTheme"))

	theme, ok := common.Themes[newTheme]
	if !ok || theme.HideFromThemes {
		return common.LocalErrorJSQ("That theme doesn't exist", w, r, user, isJs)
	}

	cookie := http.Cookie{Name: "current_theme", Value: newTheme, Path: "/", MaxAge: int(common.Year)}
	http.SetCookie(w, &cookie)

	if !isJs {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}
