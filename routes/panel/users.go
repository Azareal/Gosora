package panel

import (
	"database/sql"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	c "github.com/Azareal/Gosora/common"
)

func Users(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, u, "users", "users")
	if ferr != nil {
		return ferr
	}

	name := r.FormValue("s-name")
	email := r.FormValue("s-email")
	hasParam := name != "" || email != ""

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 15
	userCount := basePage.Stats.Users
	if hasParam {
		userCount = c.Users.CountSearch(name, email)
	}
	offset, page, lastPage := c.PageOffset(userCount, page, perPage)

	var users []*c.User
	var e error
	if hasParam {
		users, e = c.Users.SearchOffset(name, email, offset, perPage)
	} else {
		users, e = c.Users.GetOffset(offset, perPage)
	}
	if e != nil {
		return c.InternalError(e, w, r)
	}

	name = url.QueryEscape(name)
	email = url.QueryEscape(email)
	search := c.PanelUserPageSearch{name, email}

	var params string
	if hasParam {
		if name != "" {
			params += "s-name=" + name + "&"
		}
		if email != "" {
			params += "s-email=" + email + "&"
		}
	}
	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelUserPage{basePage, users, search, c.PaginatorMod{template.URL(params), pageList, page, lastPage}}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_users", &pi})
}

func UsersEdit(w http.ResponseWriter, r *http.Request, u *c.User, suid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, u, "edit_user", "users")
	if ferr != nil {
		return ferr
	}
	if !u.Perms.EditUser {
		return c.NoPermissions(w, r, u)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return c.LocalError("The provided UserID is not a valid number.", w, r, u)
	}
	targetUser, err := c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to edit doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if targetUser.IsAdmin && !u.IsAdmin {
		return c.LocalError("Only administrators can edit the account of an administrator.", w, r, u)
	}

	// ? - Should we stop admins from deleting all the groups? Maybe, protect the group they're currently using?
	groups, err := c.Groups.GetRange(1, 0) // ? - 0 = Go to the end
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var groupList []*c.Group
	for _, group := range groups {
		if !u.Perms.EditUserGroupAdmin && group.IsAdmin {
			continue
		}
		if !u.Perms.EditUserGroupSuperMod && group.IsMod {
			continue
		}
		groupList = append(groupList, group)
	}

	if r.FormValue("updated") == "1" {
		basePage.AddNotice("panel_user_updated")
	}
	showEmail := r.FormValue("show-email") == "1"

	pi := c.PanelUserEditPage{basePage, groupList, targetUser, showEmail}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_user_edit", &pi})
}

func UsersEditSubmit(w http.ResponseWriter, r *http.Request, user *c.User, suid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditUser {
		return c.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return c.LocalError("The provided UserID is not a valid number.", w, r, user)
	}
	targetUser, err := c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if targetUser.IsAdmin && !user.IsAdmin {
		return c.LocalError("Only administrators can edit the account of other administrators.", w, r, user)
	}

	newName := c.SanitiseSingleLine(r.PostFormValue("name"))
	if newName == "" {
		return c.LocalError("You didn't put in a name.", w, r, user)
	}

	// TODO: How should activation factor into admin set emails?
	// TODO: How should we handle secondary emails? Do we even have secondary emails implemented?
	newEmail := c.SanitiseSingleLine(r.PostFormValue("email"))
	if newEmail == "" && targetUser.Email != "" {
		return c.LocalError("You didn't put in an email address.", w, r, user)
	}
	if newEmail == "-1" {
		newEmail = targetUser.Email
	}
	if (newEmail != targetUser.Email) && !user.Perms.EditUserEmail {
		return c.LocalError("You need the EditUserEmail permission to edit the email address of a user.", w, r, user)
	}

	newPassword := r.PostFormValue("password")
	if newPassword != "" && !user.Perms.EditUserPassword {
		return c.LocalError("You need the EditUserPassword permission to edit the password of a user.", w, r, user)
	}

	newGroup, err := strconv.Atoi(r.PostFormValue("group"))
	if err != nil {
		return c.LocalError("You need to provide a whole number for the group ID", w, r, user)
	}
	group, err := c.Groups.Get(newGroup)
	if err == sql.ErrNoRows {
		return c.LocalError("The group you're trying to place this user in doesn't exist.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if !user.Perms.EditUserGroupAdmin && group.IsAdmin {
		return c.LocalError("You need the EditUserGroupAdmin permission to assign someone to an administrator group.", w, r, user)
	}
	if !user.Perms.EditUserGroupSuperMod && group.IsMod {
		return c.LocalError("You need the EditUserGroupSuperMod permission to assign someone to a super mod group.", w, r, user)
	}

	err = targetUser.Update(newName, c.CanonEmail(newEmail), newGroup)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	red := false
	if newPassword != "" {
		c.SetPassword(targetUser.ID, newPassword)
		// Log the user out as a safety precaution
		c.Auth.ForceLogout(targetUser.ID)
		red = true
	}
	targetUser.CacheRemove()

	targetUser, err = c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	err = c.GroupPromotions.PromoteIfEligible(targetUser, targetUser.Level, targetUser.Posts, targetUser.CreatedAt)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	targetUser.CacheRemove()

	err = c.AdminLogs.Create("edit", targetUser.ID, "user", user.GetIP(), user.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// If we're changing our own password, redirect to the index rather than to a noperms error due to the force logout
	if targetUser.ID == user.ID && red {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		var se string
		if r.PostFormValue("show-email") == "1" {
			se = "&show-email=1"
		}
		http.Redirect(w, r, "/panel/users/edit/"+strconv.Itoa(targetUser.ID)+"?updated=1"+se, http.StatusSeeOther)
	}
	return nil
}

func UsersAvatarSubmit(w http.ResponseWriter, r *http.Request, u *c.User, suid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	// TODO: Check the UploadAvatars permission too?
	if !u.Perms.EditUser {
		return c.NoPermissions(w, r, u)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return c.LocalError("The provided UserID is not a valid number.", w, r, u)
	}
	targetUser, err := c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to edit doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if targetUser.IsAdmin && !u.IsAdmin {
		return c.LocalError("Only administrators can edit the account of other administrators.", w, r, u)
	}

	ext, ferr := c.UploadAvatar(w, r, u, targetUser.ID)
	if ferr != nil {
		return ferr
	}
	ferr = c.ChangeAvatar("."+ext, w, r, targetUser)
	if ferr != nil {
		return ferr
	}
	// TODO: Only schedule a resize if the avatar isn't tiny
	err = targetUser.ScheduleAvatarResize()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	err = c.AdminLogs.Create("edit", targetUser.ID, "user", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var se string
	if r.PostFormValue("show-email") == "1" {
		se = "&show-email=1"
	}
	http.Redirect(w, r, "/panel/users/edit/"+strconv.Itoa(targetUser.ID)+"?updated=1"+se, http.StatusSeeOther)
	return nil
}

func UsersAvatarRemoveSubmit(w http.ResponseWriter, r *http.Request, u *c.User, suid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.EditUser {
		return c.NoPermissions(w, r, u)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return c.LocalError("The provided UserID is not a valid number.", w, r, u)
	}
	targetUser, err := c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to edit doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if targetUser.IsAdmin && !u.IsAdmin {
		return c.LocalError("Only administrators can edit the account of other administrators.", w, r, u)
	}
	ferr = c.ChangeAvatar("", w, r, targetUser)
	if ferr != nil {
		return ferr
	}

	err = c.AdminLogs.Create("edit", targetUser.ID, "user", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var se string
	if r.PostFormValue("show-email") == "1" {
		se = "&show-email=1"
	}
	http.Redirect(w, r, "/panel/users/edit/"+strconv.Itoa(targetUser.ID)+"?updated=1"+se, http.StatusSeeOther)
	return nil
}
