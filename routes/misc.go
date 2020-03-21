package routes

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	//"fmt"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

var cacheControlMaxAge = "max-age=" + strconv.Itoa(int(c.Day))      // TODO: Make this a c.Config value
var cacheControlMaxAgeWeek = "max-age=" + strconv.Itoa(int(c.Week)) // TODO: Make this a c.Config value

// GET functions
func StaticFile(w http.ResponseWriter, r *http.Request) {
	file, ok := c.StaticFiles.Get(r.URL.Path)
	if !ok {
		//c.DebugLogf("Failed to find '%s'", r.URL.Path) // TODO: Use MicroNotFound? Might be better than the unneccessary overhead of sprintf
		w.WriteHeader(http.StatusNotFound)
		return
	}
	h := w.Header()

	if file.Length > 300 {
		if h.Get("Range") != "" {
			h.Set("Vary", "Accept-Encoding")
			if len(file.Sha256) != 0 {
				h.Set("Cache-Control", cacheControlMaxAgeWeek)
			} else {
				h.Set("Cache-Control", cacheControlMaxAge) //Cache-Control: max-age=31536000
			}
			
			if file.GzipLength > 300 && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				h.Set("Content-Encoding", "gzip")
				h.Set("Content-Length", file.StrGzipLength)
				http.ServeContent(w, r, r.URL.Path, file.Info.ModTime(), bytes.NewReader(file.GzipData))
				return
			}

			h.Set("Content-Length", file.StrLength)
			http.ServeContent(w, r, r.URL.Path, file.Info.ModTime(), bytes.NewReader(file.Data))
			return
		}
	}

	// Surely, there's a more efficient way of doing this?
	t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since"))
	if err == nil && file.Info.ModTime().Before(t.Add(1*time.Second)) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	h.Set("Last-Modified", file.FormattedModTime)
	h.Set("Content-Type", file.Mimetype)
	if len(file.Sha256) != 0 {
		h.Set("Cache-Control", cacheControlMaxAgeWeek)
	} else {
		h.Set("Cache-Control", cacheControlMaxAge) //Cache-Control: max-age=31536000
	}
	h.Set("Vary", "Accept-Encoding")

	if file.GzipLength > 0 && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		h.Set("Content-Encoding", "gzip")
		h.Set("Content-Length", file.StrGzipLength)
		io.Copy(w, bytes.NewReader(file.GzipData)) // Use w.Write instead?
	} else {
		h.Set("Content-Length", file.StrLength)
		io.Copy(w, bytes.NewReader(file.Data))
	}
	// Other options instead of io.Copy: io.CopyN(), w.Write(), http.ServeContent()
}

func Overview(w http.ResponseWriter, r *http.Request, u *c.User, h *c.Header) c.RouteError {
	h.Title = phrases.GetTitlePhrase("overview")
	h.Zone = "overview"
	return renderTemplate("overview", w, r, h, c.Page{h, tList, nil})
}

func CustomPage(w http.ResponseWriter, r *http.Request, u *c.User, h *c.Header, name string) c.RouteError {
	h.Zone = "custom_page"
	name = c.SanitiseSingleLine(name)
	page, err := c.Pages.GetByName(name)
	if err == nil {
		h.Title = page.Title
		return renderTemplate("custom_page", w, r, h, c.CustomPagePage{h, page})
	} else if err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	h.Title = phrases.GetTitlePhrase("page")

	// TODO: Pass the page name to the pre-render hook?
	err = renderTemplate3("page_"+name, "tmpl_page", w, r, h, c.Page{h, tList, nil})
	if err == c.ErrBadDefaultTemplate {
		return c.NotFound(w, r, h)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	return nil
}

// TODO: Set the cookie domain
func ChangeTheme(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	//headerLite, _ := SimpleUserCheck(w, r, u)
	// TODO: Rename js to something else, just in case we rewrite the JS side in WebAssembly?
	js := r.PostFormValue("js") == "1"
	newTheme := c.SanitiseSingleLine(r.PostFormValue("theme"))
	//fmt.Printf("newTheme: %+v\n", newTheme)

	theme, ok := c.Themes[newTheme]
	if !ok || theme.HideFromThemes {
		return c.LocalErrorJSQ("That theme doesn't exist", w, r, u, js)
	}

	cookie := http.Cookie{Name: "current_theme", Value: newTheme, Path: "/", MaxAge: int(c.Year)}
	http.SetCookie(w, &cookie)

	if !js {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}
