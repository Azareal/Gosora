package routes

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"../common"
	"../query_gen/lib"
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
func ViewProfile(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	// TODO: Preload this?
	header.AddSheet(header.Theme.Name + "/profile.css")

	var err error
	var replyCreatedAt time.Time
	var replyContent, replyCreatedByName, replyRelativeCreatedAt, replyAvatar, replyMicroAvatar, replyTag, replyClassName string
	var rid, replyCreatedBy, replyLastEdit, replyLastEditBy, replyLines, replyGroup int
	var replyList []common.ReplyUser

	// SEO URLs...
	// TODO: Do a 301 if it's the wrong username? Do a canonical too?
	halves := strings.Split(r.URL.Path[len("/user/"):], ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	pid, err := strconv.Atoi(halves[1])
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
	header.Title = common.GetTitlePhrasef("profile", puser.Name)
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
			replyTag = common.GetTmplPhrase("profile_owner_tag")
		} else {
			replyTag = ""
		}

		replyLiked := false
		replyLikeCount := 0
		replyRelativeCreatedAt = common.RelativeTime(replyCreatedAt)

		// TODO: Add a hook here

		replyList = append(replyList, common.ReplyUser{rid, puser.ID, replyContent, common.ParseMessage(replyContent, 0, ""), replyCreatedBy, common.BuildProfileURL(common.NameToSlug(replyCreatedByName), replyCreatedBy), replyCreatedByName, replyGroup, replyCreatedAt, replyRelativeCreatedAt, replyLastEdit, replyLastEditBy, replyAvatar, replyMicroAvatar, replyClassName, replyLines, replyTag, "", "", "", 0, "", replyLiked, replyLikeCount, "", ""})
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
	if common.RunPreRenderHook("pre_render_profile", w, r, &user, &ppage) {
		return nil
	}

	err = common.RunThemeTemplate(header.Theme.Name, "profile", ppage, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}
