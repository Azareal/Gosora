package routes

import (
	"database/sql"
	"math"
	"net/http"
	"time"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
	qgen "github.com/Azareal/Gosora/query_gen"
)

type ProfileStmts struct {
	getReplies *sql.Stmt
}

var profileStmts ProfileStmts

// TODO: Move these DbInits into some sort of abstraction
func init() {
	c.DbInits.Add(func(acc *qgen.Accumulator) error {
		profileStmts = ProfileStmts{
			getReplies: acc.SimpleLeftJoin("users_replies", "users", "users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.group", "users_replies.createdBy=users.uid", "users_replies.uid=?", "", ""),
		}
		return acc.FirstError()
	})
}

// TODO: Remove the View part of the name?
func ViewProfile(w http.ResponseWriter, r *http.Request, user *c.User, h *c.Header) c.RouteError {
	var reCreatedAt time.Time
	var reContent, reCreatedByName, reAvatar string
	var rid, reCreatedBy, reLastEdit, reLastEditBy, reGroup int
	var reList []*c.ReplyUser

	// TODO: Do a 301 if it's the wrong username? Do a canonical too?
	_, pid, err := ParseSEOURL(r.URL.Path[len("/user/"):])
	if err != nil {
		return c.SimpleError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, h)
	}
	if pid == 0 {
		return c.NotFound(w, r, h)
	}

	var puser *c.User
	if pid == user.ID {
		user.IsMod = true
		puser = user
	} else {
		// Fetch the user data
		// TODO: Add a shared function for checking for ErrNoRows and internal erroring if it's not that case?
		puser, err = c.Users.Get(pid)
		if err == sql.ErrNoRows {
			return c.NotFound(w, r, h)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
	}
	h.Title = phrases.GetTitlePhrasef("profile", puser.Name)
	h.Path = c.BuildProfileURL(c.NameToSlug(puser.Name), puser.ID)
	// TODO: Preload this?
	h.AddSheet(h.Theme.Name + "/profile.css")
	if user.Loggedin {
		h.AddScriptAsync("profile_member.js")
	}

	// Get the replies..
	rows, err := profileStmts.getReplies.Query(puser.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&rid, &reContent, &reCreatedBy, &reCreatedAt, &reLastEdit, &reLastEditBy, &reAvatar, &reCreatedByName, &reGroup)
		if err != nil {
			return c.InternalError(err, w, r)
		}

		reLiked := false
		reLikeCount := 0
		ru := &c.ReplyUser{Reply: c.Reply{rid, puser.ID, reContent, reCreatedBy /*, reGroup*/, reCreatedAt, reLastEdit, reLastEditBy, 0, "", reLiked, reLikeCount, 0, ""}, ContentHtml: c.ParseMessage(reContent, 0, "", user.ParseSettings, user), CreatedByName: reCreatedByName, Avatar: reAvatar, Group: reGroup, Level: 0}
		_, err = ru.Init(user)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		if puser.ID == ru.CreatedBy {
			ru.Tag = phrases.GetTmplPhrase("profile.owner_tag")
		}

		// TODO: Add a hook here
		reList = append(reList, ru)
	}
	if err := rows.Err(); err != nil {
		return c.InternalError(err, w, r)
	}

	// Normalise the score so that the user sees their relative progress to the next level rather than showing them their total score
	prevScore := c.GetLevelScore(puser.Level)
	score := puser.Score
	//score = 23
	currentScore := score - prevScore
	nextScore := c.GetLevelScore(puser.Level+1) - prevScore
	perc := int(math.Floor((float64(currentScore) / float64(nextScore)) * 100))
	var blocked, blockedInv bool
	if user.Loggedin {
		blocked, err = c.UserBlocks.IsBlockedBy(user.ID, puser.ID)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		blockedInv, err = c.UserBlocks.IsBlockedBy(puser.ID, user.ID)
		if err != nil {
			return c.InternalError(err, w, r)
		}
	}

	canMessage := (!blockedInv && user.Perms.UseConvos) || (!blockedInv && puser.IsSuperMod && user.Perms.UseConvosOnlyWithMod) || user.IsSuperMod
	canComment := !blockedInv && user.Perms.CreateProfileReply
	showComments := c.PrivacyCommentsShow(puser, user)
	if !showComments {
		canComment = false
	}
	if !c.PrivacyAllowMessage(puser, user) {
		canMessage = false
	}

	ppage := c.ProfilePage{h, reList, *puser, currentScore, nextScore, perc, blocked, canMessage, canComment, showComments}
	return renderTemplate("profile", w, r, h, ppage)
}
