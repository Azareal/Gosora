package routes

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
)

type AttachmentStmts struct {
	get *sql.Stmt
}

var attachmentStmts AttachmentStmts

// TODO: Abstract this with an attachment store
func init() {
	c.DbInits.Add(func(acc *qgen.Accumulator) error {
		attachmentStmts = AttachmentStmts{
			get: acc.Select("attachments").Columns("sectionID, sectionTable, originID, originTable, uploadedBy, path").Where("path=? AND sectionID=? AND sectionTable=?").Prepare(),
		}
		return acc.FirstError()
	})
}

var maxAgeYear = "max-age=" + strconv.Itoa(int(c.Year))

func ShowAttachment(w http.ResponseWriter, r *http.Request, user *c.User, filename string) c.RouteError {
	filename = c.Stripslashes(filename)
	ext := filepath.Ext("./attachs/" + filename)
	if !c.AllowedFileExts.Contains(strings.TrimPrefix(ext, ".")) {
		return c.LocalError("Bad extension", w, r, user)
	}

	sid, err := strconv.Atoi(r.FormValue("sid"))
	if err != nil {
		return c.LocalError("The sid is not an integer", w, r, user)
	}
	sectionTable := r.FormValue("stype")

	var originTable string
	var originID, uploadedBy int
	err = attachmentStmts.get.QueryRow(filename, sid, sectionTable).Scan(&sid, &sectionTable, &originID, &originTable, &uploadedBy, &filename)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	if sectionTable == "forums" {
		_, ferr := c.SimpleForumUserCheck(w, r, user, sid)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic {
			return c.NoPermissions(w, r, user)
		}
	} else {
		return c.LocalError("Unknown section", w, r, user)
	}

	if originTable != "topics" && originTable != "replies" {
		return c.LocalError("Unknown origin", w, r, user)
	}

	if !user.Loggedin {
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
	err := c.DeleteAttachment(aid)
	if err == sql.ErrNoRows {
		return c.NotFoundJSQ(w, r, nil, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
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
