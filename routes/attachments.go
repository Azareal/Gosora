package routes

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
)

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
