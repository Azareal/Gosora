package routes

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
)

var maxAgeYear = "max-age=" + strconv.Itoa(int(c.Year))

func ShowAttachment(w http.ResponseWriter, r *http.Request, u *c.User, filename string) c.RouteError {
	sid, err := strconv.Atoi(r.FormValue("sid"))
	if err != nil {
		return c.LocalError("The sid is not an integer", w, r, u)
	}
	sectionTable := r.FormValue("stype")

	filename = c.Stripslashes(filename)
	if filename == "" {
		return c.LocalError("Bad filename", w, r, u)
	}
	ext := filepath.Ext(filename)
	if ext == "" || !c.AllowedFileExts.Contains(strings.TrimPrefix(ext, ".")) {
		return c.LocalError("Bad extension", w, r, u)
	}

	// TODO: Use the same hook table as upstream
	hTbl := c.GetHookTable()
	skip, rerr := c.H_route_attach_start_hook(hTbl, w, r, u, filename)
	if skip || rerr != nil {
		return rerr
	}

	a, err := c.Attachments.GetForRenderRoute(filename, sid, sectionTable)
	// ErrCorruptAttachPath is a possibility now
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	skip, rerr = c.H_route_attach_post_get_hook(hTbl, w, r, u, a)
	if skip || rerr != nil {
		return rerr
	}

	if a.SectionTable == "forums" {
		_, ferr := c.SimpleForumUserCheck(w, r, u, sid)
		if ferr != nil {
			return ferr
		}
		if !u.Perms.ViewTopic {
			return c.NoPermissions(w, r, u)
		}
	} else {
		return c.LocalError("Unknown section", w, r, u)
	}

	if a.OriginTable != "topics" && a.OriginTable != "replies" {
		return c.LocalError("Unknown origin", w, r, u)
	}

	if !u.Loggedin {
		w.Header().Set("Cache-Control", maxAgeYear)
	} else {
		guest := c.GuestUser
		_, ferr := c.SimpleForumUserCheck(w, r, &guest, sid)
		if ferr != nil {
			return ferr
		}
		h := w.Header()
		if guest.Perms.ViewTopic {
			h.Set("Cache-Control", maxAgeYear)
		} else {
			h.Set("Cache-Control", "private")
		}
	}

	// TODO: Fix the problem where non-existent files aren't greeted with custom 404s on ServeFile()'s side
	http.ServeFile(w, r, "./attachs/"+filename)
	return nil
}

func deleteAttachment(w http.ResponseWriter, r *http.Request, u *c.User, aid int, js bool) c.RouteError {
	e := c.DeleteAttachment(aid)
	if e == sql.ErrNoRows {
		return c.NotFoundJSQ(w, r, nil, js)
	} else if e != nil {
		return c.InternalErrorJSQ(e, w, r, js)
	}
	return nil
}

// TODO: Stop duplicating this code
// TODO: Use a transaction here
// TODO: Move this function to neutral ground
func uploadAttachment(w http.ResponseWriter, r *http.Request, u *c.User, sid int, stable string, oid int, otable, extra string) (pathMap map[string]string, rerr c.RouteError) {
	pathMap = make(map[string]string)
	files, rerr := uploadFilesWithHash(w, r, u, "./attachs/")
	if rerr != nil {
		return nil, rerr
	}

	for _, filename := range files {
		aid, err := c.Attachments.Add(sid, stable, oid, otable, u.ID, filename, extra)
		if err != nil {
			return nil, c.InternalError(err, w, r)
		}

		_, ok := pathMap[filename]
		if ok {
			pathMap[filename] += "," + strconv.Itoa(aid)
		} else {
			pathMap[filename] = strconv.Itoa(aid)
		}

		err = c.Attachments.AddLinked(otable, oid)
		if err != nil {
			return nil, c.InternalError(err, w, r)
		}
	}

	return pathMap, nil
}
