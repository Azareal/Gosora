package routes

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
)

type ProfileStmts struct {
	getReplies *sql.Stmt
}

var profileStmts ProfileStmts

// TODO: Move these DbInits into some sort of abstraction
func init() {
	common.DbInits.Add(func(acc *qgen.Accumulator) error {
		profileStmts = ProfileStmts{
			getReplies: acc.SimpleLeftJoin("users_replies", "users", "users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.group", "users_replies.createdBy = users.uid", "users_replies.uid = ?", "", ""),
		}
		return acc.FirstError()
	})
}

// TODO: Remove the View part of the name?
func ViewProfile(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	// TODO: Preload this?
	header.AddSheet(header.Theme.Name + "/profile.css")
	if user.Loggedin {
		header.AddScript("profile_member.js")
	}

	var err error
	var replyCreatedAt time.Time
	var replyContent, replyCreatedByName, replyAvatar, replyMicroAvatar, replyTag, replyClassName string
	var rid, replyCreatedBy, replyLastEdit, replyLastEditBy, replyLines, replyGroup int
	var replyList []common.ReplyUser

	// TODO: Do a 301 if it's the wrong username? Do a canonical too?
	_, pid, err := ParseSEOURL(r.URL.Path[len("/user/"):])
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	var puser *common.User
	if pid == user.ID {
		user.IsMod = true
		puser = &user
	} else {
		// Fetch the user data
		// TODO: Add a shared function for checking for ErrNoRows and internal erroring if it's not that case?
		puser, err = common.Users.Get(pid)
		if err == sql.ErrNoRows {
			return common.NotFound(w, r, header)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}
		puser.Init()
	}
	header.Title = phrases.GetTitlePhrasef("profile", puser.Name)
	header.Path = common.BuildProfileURL(common.NameToSlug(puser.Name), puser.ID)

	// Get the replies..
	rows, err := profileStmts.getReplies.Query(puser.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&rid, &replyContent, &replyCreatedBy, &replyCreatedAt, &replyLastEdit, &replyLastEditBy, &replyAvatar, &replyCreatedByName, &replyGroup)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		group, err := common.Groups.Get(replyGroup)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		replyLines = strings.Count(replyContent, "\n")
		if group.IsMod {
			replyClassName = common.Config.StaffCSS
		} else {
			replyClassName = ""
		}
		replyAvatar, replyMicroAvatar = common.BuildAvatar(replyCreatedBy, replyAvatar)

		if group.Tag != "" {
			replyTag = group.Tag
		} else if puser.ID == replyCreatedBy {
			replyTag = phrases.GetTmplPhrase("profile_owner_tag")
		} else {
			replyTag = ""
		}

		replyLiked := false
		replyLikeCount := 0
		// TODO: Add a hook here

		replyList = append(replyList, common.ReplyUser{rid, puser.ID, replyContent, common.ParseMessage(replyContent, 0, ""), replyCreatedBy, common.BuildProfileURL(common.NameToSlug(replyCreatedByName), replyCreatedBy), replyCreatedByName, replyGroup, replyCreatedAt, replyLastEdit, replyLastEditBy, replyAvatar, replyMicroAvatar, replyClassName, replyLines, replyTag, "", "", "", 0, "", replyLiked, replyLikeCount, 0, "", "", nil})
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Normalise the score so that the user sees their relative progress to the next level rather than showing them their total score
	prevScore := common.GetLevelScore(puser.Level)
	currentScore := puser.Score - prevScore
	nextScore := common.GetLevelScore(puser.Level+1) - prevScore

	ppage := common.ProfilePage{header, replyList, *puser, currentScore, nextScore}
	return renderTemplate("profile", w, r, header, ppage)
}
