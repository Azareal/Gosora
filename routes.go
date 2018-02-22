/*
*
*	Gosora Route Handlers
*	Copyright Azareal 2016 - 2018
*
 */
package main

import (
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"./common"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

//var nList []string
var successJSONBytes = []byte(`{"success":"1"}`)

// HTTPSRedirect is a connection handler which redirects all HTTP requests to HTTPS
type HTTPSRedirect struct {
}

func (red *HTTPSRedirect) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	dest := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		dest += "?" + req.URL.RawQuery
	}
	http.Redirect(w, req, dest, http.StatusTemporaryRedirect)
}

// Temporary stubs for view tracking
func routeDynamic() {
}
func routeUploads() {
}
func BadRoute() {
}

func routeForums(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Zone = "forums"
	headerVars.MetaDesc = headerVars.Settings["meta_desc"].(string)

	var err error
	var forumList []common.Forum
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = common.Forums.GetAllVisibleIDs()
		if err != nil {
			return common.InternalError(err, w, r)
		}
	} else {
		group, err := common.Groups.Get(user.Group)
		if err != nil {
			log.Printf("Group #%d doesn't exist despite being used by common.User #%d", user.Group, user.ID)
			return common.LocalError("Something weird happened", w, r, user)
		}
		canSee = group.CanSee
	}

	for _, fid := range canSee {
		// Avoid data races by copying the struct into something we can freely mold without worrying about breaking something somewhere else
		var forum = common.Forums.DirtyGet(fid).Copy()
		if forum.ParentID == 0 && forum.Name != "" && forum.Active {
			if forum.LastTopicID != 0 {
				if forum.LastTopic.ID != 0 && forum.LastReplyer.ID != 0 {
					forum.LastTopicTime = common.RelativeTime(forum.LastTopic.LastReplyAt)
				} else {
					forum.LastTopicTime = ""
				}
			} else {
				forum.LastTopicTime = ""
			}
			common.RunHook("forums_frow_assign", &forum)
			forumList = append(forumList, forum)
		}
	}

	pi := common.ForumsPage{common.GetTitlePhrase("forums"), user, headerVars, forumList}
	if common.RunPreRenderHook("pre_render_forum_list", w, r, &user, &pi) {
		return nil
	}
	err = common.RunThemeTemplate(headerVars.Theme.Name, "forums", pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeProfile(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var err error
	var replyCreatedAt time.Time
	var replyContent, replyCreatedByName, replyRelativeCreatedAt, replyAvatar, replyTag, replyClassName string
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
		if err == ErrNoRows {
			return common.NotFound(w, r, headerVars)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}
	}

	// Get the replies..
	rows, err := stmts.getProfileReplies.Query(puser.ID)
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
		if group.IsMod || group.IsAdmin {
			replyClassName = common.Config.StaffCSS
		} else {
			replyClassName = ""
		}
		replyAvatar = common.BuildAvatar(replyCreatedBy, replyAvatar)

		if group.Tag != "" {
			replyTag = group.Tag
		} else if puser.ID == replyCreatedBy {
			replyTag = "Profile Owner"
		} else {
			replyTag = ""
		}

		replyLiked := false
		replyLikeCount := 0
		replyRelativeCreatedAt = common.RelativeTime(replyCreatedAt)

		// TODO: Add a hook here

		replyList = append(replyList, common.ReplyUser{rid, puser.ID, replyContent, common.ParseMessage(replyContent, 0, ""), replyCreatedBy, common.BuildProfileURL(common.NameToSlug(replyCreatedByName), replyCreatedBy), replyCreatedByName, replyGroup, replyCreatedAt, replyRelativeCreatedAt, replyLastEdit, replyLastEditBy, replyAvatar, replyClassName, replyLines, replyTag, "", "", "", 0, "", replyLiked, replyLikeCount, "", ""})
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add a phrase for this title
	ppage := common.ProfilePage{puser.Name + "'s Profile", user, headerVars, replyList, *puser}
	if common.RunPreRenderHook("pre_render_profile", w, r, &user, &ppage) {
		return nil
	}

	err = common.RunThemeTemplate(headerVars.Theme.Name, "profile", ppage, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

// TODO: Set the cookie domain
func routeChangeTheme(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	//headerLite, _ := SimpleUserCheck(w, r, &user)
	// TODO: Rename isJs to something else, just in case we rewrite the JS side in WebAssembly?
	isJs := (r.PostFormValue("isJs") == "1")
	newTheme := html.EscapeString(r.PostFormValue("newTheme"))

	theme, ok := common.Themes[newTheme]
	if !ok || theme.HideFromThemes {
		return common.LocalErrorJSQ("That theme doesn't exist", w, r, user, isJs)
	}

	cookie := http.Cookie{Name: "current_theme", Value: newTheme, Path: "/", MaxAge: common.Year}
	http.SetCookie(w, &cookie)

	if !isJs {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}

// TODO: Refactor this
var phraseLoginAlerts = []byte(`{"msgs":[{"msg":"Login to see your alerts","path":"/accounts/login"}]}`)

// TODO: Refactor this endpoint
func routeAPI(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	// TODO: Don't make this too JSON dependent so that we can swap in newer more efficient formats
	w.Header().Set("Content-Type", "application/json")
	err := r.ParseForm()
	if err != nil {
		return common.PreErrorJS("Bad Form", w, r)
	}

	action := r.FormValue("action")
	if action != "get" && action != "set" {
		return common.PreErrorJS("Invalid Action", w, r)
	}

	module := r.FormValue("module")
	switch module {
	case "dismiss-alert":
		asid, err := strconv.Atoi(r.FormValue("asid"))
		if err != nil {
			return common.PreErrorJS("Invalid asid", w, r)
		}
		_, err = stmts.deleteActivityStreamMatch.Exec(user.ID, asid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
	case "alerts": // A feed of events tailored for a specific user
		if !user.Loggedin {
			w.Write(phraseLoginAlerts)
			return nil
		}

		var msglist, event, elementType string
		var asid, actorID, targetUserID, elementID int
		var msgCount int

		err = stmts.getActivityCountByWatcher.QueryRow(user.ID).Scan(&msgCount)
		if err == ErrNoRows {
			return common.PreErrorJS("Couldn't find the parent topic", w, r)
		} else if err != nil {
			return common.InternalErrorJS(err, w, r)
		}

		rows, err := stmts.getActivityFeedByWatcher.Query(user.ID)
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}
		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&asid, &actorID, &targetUserID, &event, &elementType, &elementID)
			if err != nil {
				return common.InternalErrorJS(err, w, r)
			}
			res, err := buildAlert(asid, event, elementType, actorID, targetUserID, elementID, user)
			if err != nil {
				return common.LocalErrorJS(err.Error(), w, r)
			}
			msglist += res + ","
		}
		err = rows.Err()
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}

		if len(msglist) != 0 {
			msglist = msglist[0 : len(msglist)-1]
		}
		_, _ = w.Write([]byte(`{"msgs":[` + msglist + `],"msgCount":` + strconv.Itoa(msgCount) + `}`))
	default:
		return common.PreErrorJS("Invalid Module", w, r)
	}
	return nil
}
