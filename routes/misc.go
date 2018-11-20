package routes

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
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
	if common.Templates.Lookup("page_"+name+".html") == nil {
		return common.NotFound(w, r, header)
	}

	header.Title = phrases.GetTitlePhrase("page")
	pi := common.Page{header, tList, nil}
	// TODO: Pass the page name to the pre-render hook?
	if common.RunPreRenderHook("pre_render_tmpl_page", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "page_"+name+".html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

type AttachmentStmts struct {
	get *sql.Stmt
}

var attachmentStmts AttachmentStmts

// TODO: Abstract this with an attachment store
func init() {
	common.DbInits.Add(func(acc *qgen.Accumulator) error {
		attachmentStmts = AttachmentStmts{
			get: acc.Select("attachments").Columns("sectionID, sectionTable, originID, originTable, uploadedBy, path").Where("path = ? AND sectionID = ? AND sectionTable = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

func ShowAttachment(w http.ResponseWriter, r *http.Request, user common.User, filename string) common.RouteError {
	filename = common.Stripslashes(filename)
	var ext = filepath.Ext("./attachs/" + filename)
	if !common.AllowedFileExts.Contains(strings.TrimPrefix(ext, ".")) {
		return common.LocalError("Bad extension", w, r, user)
	}

	sectionID, err := strconv.Atoi(r.FormValue("sectionID"))
	if err != nil {
		return common.LocalError("The sectionID is not an integer", w, r, user)
	}
	var sectionTable = r.FormValue("sectionType")

	var originTable string
	var originID, uploadedBy int
	err = attachmentStmts.get.QueryRow(filename, sectionID, sectionTable).Scan(&sectionID, &sectionTable, &originID, &originTable, &uploadedBy, &filename)
	if err == sql.ErrNoRows {
		return common.NotFound(w, r, nil)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if sectionTable == "forums" {
		_, ferr := common.SimpleForumUserCheck(w, r, &user, sectionID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic {
			return common.NoPermissions(w, r, user)
		}
	} else {
		return common.LocalError("Unknown section", w, r, user)
	}

	if originTable != "topics" && originTable != "replies" {
		return common.LocalError("Unknown origin", w, r, user)
	}

	if !user.Loggedin {
		w.Header().Set("Cache-Control", "max-age="+strconv.Itoa(int(common.Year)))
	} else {
		guest := common.GuestUser
		_, ferr := common.SimpleForumUserCheck(w, r, &guest, sectionID)
		if ferr != nil {
			return ferr
		}
		if guest.Perms.ViewTopic {
			w.Header().Set("Cache-Control", "max-age="+strconv.Itoa(int(common.Year)))
		} else {
			w.Header().Set("Cache-Control", "private")
		}
	}

	// TODO: Fix the problem where non-existent files aren't greeted with custom 404s on ServeFile()'s side
	http.ServeFile(w, r, "./attachs/"+filename)
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
