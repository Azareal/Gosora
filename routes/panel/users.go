package panel

import (
	"database/sql"
	"net/http"
	"strconv"

	c "github.com/Azareal/Gosora/common"
)

func Users(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "users", "users")
	if ferr != nil {
		return ferr
	}
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 15
	offset, page, lastPage := c.PageOffset(basePage.Stats.Users, page, perPage)

	users, err := c.Users.GetOffset(offset, perPage)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelUserPage{basePage, users, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_users", &pi})
}

func UsersEdit(w http.ResponseWriter, r *http.Request, user c.User, suid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "edit_user", "users")
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
		return c.LocalError("Only administrators can edit the account of an administrator.", w, r, user)
	}

	// ? - Should we stop admins from deleting all the groups? Maybe, protect the group they're currently using?
	groups, err := c.Groups.GetRange(1, 0) // ? - 0 = Go to the end
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var groupList []*c.Group
	for _, group := range groups {
		if !user.Perms.EditUserGroupAdmin && group.IsAdmin {
			continue
		}
		if !user.Perms.EditUserGroupSuperMod && group.IsMod {
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

func UsersEditSubmit(w http.ResponseWriter, r *http.Request, user c.User, suid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
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

	newname := c.SanitiseSingleLine(r.PostFormValue("name"))
	if newname == "" {
		return c.LocalError("You didn't put in a name.", w, r, user)
	}

	// TODO: How should activation factor into admin set emails?
	// TODO: How should we handle secondary emails? Do we even have secondary emails implemented?
	newemail := c.SanitiseSingleLine(r.PostFormValue("email"))
	if newemail == "" && targetUser.Email != "" {
		return c.LocalError("You didn't put in an email address.", w, r, user)
	}
	if newemail == "-1" {
		newemail = targetUser.Email
	}
	if (newemail != targetUser.Email) && !user.Perms.EditUserEmail {
		return c.LocalError("You need the EditUserEmail permission to edit the email address of a user.", w, r, user)
	}

	newpassword := r.PostFormValue("password")
	if newpassword != "" && !user.Perms.EditUserPassword {
		return c.LocalError("You need the EditUserPassword permission to edit the password of a user.", w, r, user)
	}

	newgroup, err := strconv.Atoi(r.PostFormValue("group"))
	if err != nil {
		return c.LocalError("You need to provide a whole number for the group ID", w, r, user)
	}
	group, err := c.Groups.Get(newgroup)
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

	err = targetUser.Update(newname, newemail, newgroup)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	red := false
	if newpassword != "" {
		c.SetPassword(targetUser.ID, newpassword)
		// Log the user out as a safety precaution
		c.Auth.ForceLogout(targetUser.ID)
		red = true
	}
	targetUser.CacheRemove()

	err = c.AdminLogs.Create("edit", targetUser.ID, "user", user.LastIP, user.ID)
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

func UsersAvatarSubmit(w http.ResponseWriter, r *http.Request, user c.User, suid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	// TODO: Check the UploadAvatars permission too?
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

	ext, ferr := c.UploadAvatar(w, r, user, targetUser.ID)
	if ferr != nil {
		return ferr
	}
	ferr = c.ChangeAvatar("."+ext, w, r, *targetUser)
	if ferr != nil {
		return ferr
	}
	// TODO: Only schedule a resize if the avatar isn't tiny
	err = targetUser.ScheduleAvatarResize()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	err = c.AdminLogs.Create("edit", targetUser.ID, "user", user.LastIP, user.ID)
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

func UsersAvatarRemoveSubmit(w http.ResponseWriter, r *http.Request, user c.User, suid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
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
	ferr = c.ChangeAvatar("", w, r, *targetUser)
	if ferr != nil {
		return ferr
	}

	err = c.AdminLogs.Create("edit", targetUser.ID, "user", user.LastIP, user.ID)
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
