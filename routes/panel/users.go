package panel

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Azareal/Gosora/common"
)

func Users(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "users", "users")
	if ferr != nil {
		return ferr
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := common.PageOffset(basePage.Stats.Users, page, perPage)

	users, err := common.Users.GetOffset(offset, perPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pageList := common.Paginate(basePage.Stats.Users, perPage, 5)
	pi := common.PanelUserPage{basePage, users, common.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel_users", w, r, user, &pi)
}

func UsersEdit(w http.ResponseWriter, r *http.Request, user common.User, suid string) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "edit_user", "users")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditUser {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := common.Users.Get(uid)
	if err == sql.ErrNoRows {
		return common.LocalError("The user you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if targetUser.IsAdmin && !user.IsAdmin {
		return common.LocalError("Only administrators can edit the account of an administrator.", w, r, user)
	}

	// ? - Should we stop admins from deleting all the groups? Maybe, protect the group they're currently using?
	groups, err := common.Groups.GetRange(1, 0) // ? - 0 = Go to the end
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var groupList []interface{}
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

	pi := common.PanelPage{basePage, groupList, targetUser}
	if common.RunPreRenderHook("pre_render_panel_edit_user", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel_user_edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func UsersEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, suid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditUser {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := common.Users.Get(uid)
	if err == sql.ErrNoRows {
		return common.LocalError("The user you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if targetUser.IsAdmin && !user.IsAdmin {
		return common.LocalError("Only administrators can edit the account of other administrators.", w, r, user)
	}

	newname := common.SanitiseSingleLine(r.PostFormValue("user-name"))
	if newname == "" {
		return common.LocalError("You didn't put in a username.", w, r, user)
	}

	// TODO: How should activation factor into admin set emails?
	// TODO: How should we handle secondary emails? Do we even have secondary emails implemented?
	newemail := common.SanitiseSingleLine(r.PostFormValue("user-email"))
	if newemail == "" {
		return common.LocalError("You didn't put in an email address.", w, r, user)
	}
	if (newemail != targetUser.Email) && !user.Perms.EditUserEmail {
		return common.LocalError("You need the EditUserEmail permission to edit the email address of a user.", w, r, user)
	}

	newpassword := r.PostFormValue("user-password")
	if newpassword != "" && !user.Perms.EditUserPassword {
		return common.LocalError("You need the EditUserPassword permission to edit the password of a user.", w, r, user)
	}

	newgroup, err := strconv.Atoi(r.PostFormValue("user-group"))
	if err != nil {
		return common.LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := common.Groups.Get(newgroup)
	if err == sql.ErrNoRows {
		return common.LocalError("The group you're trying to place this user in doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if !user.Perms.EditUserGroupAdmin && group.IsAdmin {
		return common.LocalError("You need the EditUserGroupAdmin permission to assign someone to an administrator group.", w, r, user)
	}
	if !user.Perms.EditUserGroupSuperMod && group.IsMod {
		return common.LocalError("You need the EditUserGroupSuperMod permission to assign someone to a super mod group.", w, r, user)
	}

	err = targetUser.Update(newname, newemail, newgroup)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	if newpassword != "" {
		common.SetPassword(targetUser.ID, newpassword)
		// Log the user out as a safety precaution
		common.Auth.ForceLogout(targetUser.ID)
	}
	targetUser.CacheRemove()

	// If we're changing our own password, redirect to the index rather than to a noperms error due to the force logout
	if targetUser.ID == user.ID {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/panel/users/edit/"+strconv.Itoa(targetUser.ID)+"?updated=1", http.StatusSeeOther)
	}
	return nil
}
