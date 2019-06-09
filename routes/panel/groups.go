package panel

import (
	"database/sql"
	"net/http"
	"strconv"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func Groups(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "groups", "groups")
	if ferr != nil {
		return ferr
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 15
	offset, page, lastPage := c.PageOffset(basePage.Stats.Groups, page, perPage)

	// Skip the 'Unknown' group
	offset++

	var count int
	var groupList []c.GroupAdmin
	groups, _ := c.Groups.GetRange(offset, 0)
	for _, group := range groups {
		if count == perPage {
			break
		}
		var rank string
		var rankClass string
		var canDelete = false

		// TODO: Use a switch for this
		// TODO: Localise this
		if group.IsAdmin {
			rank = "Admin"
			rankClass = "admin"
		} else if group.IsMod {
			rank = "Mod"
			rankClass = "mod"
		} else if group.IsBanned {
			rank = "Banned"
			rankClass = "banned"
		} else if group.ID == 6 {
			rank = "Guest"
			rankClass = "guest"
		} else {
			rank = "Member"
			rankClass = "member"
		}

		canEdit := user.Perms.EditGroup && (!group.IsAdmin || user.Perms.EditGroupAdmin) && (!group.IsMod || user.Perms.EditGroupSuperMod)
		groupList = append(groupList, c.GroupAdmin{group.ID, group.Name, rank, rankClass, canEdit, canDelete})
		count++
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelGroupPage{basePage, groupList, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage,"","","panel_groups",&pi})
}

func GroupsEdit(w http.ResponseWriter, r *http.Request, user c.User, sgid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "edit_group", "groups")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return c.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return c.LocalError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	group, err := c.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters")
		return c.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return c.LocalError(phrases.GetErrorPhrase("panel_groups_cannot_edit_admin"), w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return c.LocalError(phrases.GetErrorPhrase("panel_groups_cannot_edit_supermod"), w, r, user)
	}

	var rank string
	switch {
	case group.IsAdmin:
		rank = "Admin"
	case group.IsMod:
		rank = "Mod"
	case group.IsBanned:
		rank = "Banned"
	case group.ID == 6:
		rank = "Guest"
	default:
		rank = "Member"
	}
	disableRank := !user.Perms.EditGroupGlobalPerms || (group.ID == 6)

	pi := c.PanelEditGroupPage{basePage, group.ID, group.Name, group.Tag, rank, disableRank}
	return renderTemplate("panel_group_edit", w, r, basePage.Header, pi)
}

func GroupsEditPerms(w http.ResponseWriter, r *http.Request, user c.User, sgid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "edit_group", "groups")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return c.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return c.LocalError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	group, err := c.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters")
		return c.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return c.LocalError(phrases.GetErrorPhrase("panel_groups_cannot_edit_admin"), w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return c.LocalError(phrases.GetErrorPhrase("panel_groups_cannot_edit_supermod"), w, r, user)
	}

	// TODO: Load the phrases in bulk for efficiency?
	var localPerms []c.NameLangToggle

	var addLocalPerm = func(permStr string, perm bool) {
		localPerms = append(localPerms, c.NameLangToggle{permStr, phrases.GetLocalPermPhrase(permStr), perm})
	}

	addLocalPerm("ViewTopic", group.Perms.ViewTopic)
	addLocalPerm("LikeItem", group.Perms.LikeItem)
	addLocalPerm("CreateTopic", group.Perms.CreateTopic)
	//<--
	addLocalPerm("EditTopic", group.Perms.EditTopic)
	addLocalPerm("DeleteTopic", group.Perms.DeleteTopic)
	addLocalPerm("CreateReply", group.Perms.CreateReply)
	addLocalPerm("EditReply", group.Perms.EditReply)
	addLocalPerm("DeleteReply", group.Perms.DeleteReply)
	addLocalPerm("PinTopic", group.Perms.PinTopic)
	addLocalPerm("CloseTopic", group.Perms.CloseTopic)
	addLocalPerm("MoveTopic", group.Perms.MoveTopic)

	var globalPerms []c.NameLangToggle
	var addGlobalPerm = func(permStr string, perm bool) {
		globalPerms = append(globalPerms, c.NameLangToggle{permStr, phrases.GetGlobalPermPhrase(permStr), perm})
	}

	addGlobalPerm("BanUsers", group.Perms.BanUsers)
	addGlobalPerm("ActivateUsers", group.Perms.ActivateUsers)
	addGlobalPerm("EditUser", group.Perms.EditUser)
	addGlobalPerm("EditUserEmail", group.Perms.EditUserEmail)
	addGlobalPerm("EditUserPassword", group.Perms.EditUserPassword)
	addGlobalPerm("EditUserGroup", group.Perms.EditUserGroup)
	addGlobalPerm("EditUserGroupSuperMod", group.Perms.EditUserGroupSuperMod)
	addGlobalPerm("EditUserGroupAdmin", group.Perms.EditUserGroupAdmin)
	addGlobalPerm("EditGroup", group.Perms.EditGroup)
	addGlobalPerm("EditGroupLocalPerms", group.Perms.EditGroupLocalPerms)
	addGlobalPerm("EditGroupGlobalPerms", group.Perms.EditGroupGlobalPerms)
	addGlobalPerm("EditGroupSuperMod", group.Perms.EditGroupSuperMod)
	addGlobalPerm("EditGroupAdmin", group.Perms.EditGroupAdmin)
	addGlobalPerm("ManageForums", group.Perms.ManageForums)
	addGlobalPerm("EditSettings", group.Perms.EditSettings)
	addGlobalPerm("ManageThemes", group.Perms.ManageThemes)
	addGlobalPerm("ManagePlugins", group.Perms.ManagePlugins)
	addGlobalPerm("ViewAdminLogs", group.Perms.ViewAdminLogs)
	addGlobalPerm("ViewIPs", group.Perms.ViewIPs)
	addGlobalPerm("UploadFiles", group.Perms.UploadFiles)
	addGlobalPerm("UploadAvatars", group.Perms.UploadAvatars)

	pi := c.PanelEditGroupPermsPage{basePage, group.ID, group.Name, localPerms, globalPerms}
	return renderTemplate("panel_group_edit_perms", w, r, basePage.Header, pi)
}

func GroupsEditSubmit(w http.ResponseWriter, r *http.Request, user c.User, sgid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return c.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return c.LocalError(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}

	group, err := c.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters")
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return c.LocalError(phrases.GetErrorPhrase("panel_groups_cannot_edit_admin"), w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return c.LocalError(phrases.GetErrorPhrase("panel_groups_cannot_edit_supermod"), w, r, user)
	}

	gname := r.FormValue("group-name")
	if gname == "" {
		return c.LocalError(phrases.GetErrorPhrase("panel_groups_need_name"), w, r, user)
	}
	gtag := r.FormValue("group-tag")
	rank := r.FormValue("group-type")

	var originalRank string
	// TODO: Use a switch for this
	if group.IsAdmin {
		originalRank = "Admin"
	} else if group.IsMod {
		originalRank = "Mod"
	} else if group.IsBanned {
		originalRank = "Banned"
	} else if group.ID == 6 {
		originalRank = "Guest"
	} else {
		originalRank = "Member"
	}

	if rank != originalRank && originalRank != "Guest" {
		if !user.Perms.EditGroupGlobalPerms {
			return c.LocalError(phrases.GetErrorPhrase("panel_groups_cannot_edit_group_type"), w, r, user)
		}

		switch rank {
		case "Admin":
			if !user.Perms.EditGroupAdmin {
				return c.LocalError(phrases.GetErrorPhrase("panel_groups_edit_cannot_designate_admin"), w, r, user)
			}
			err = group.ChangeRank(true, true, false)
		case "Mod":
			if !user.Perms.EditGroupSuperMod {
				return c.LocalError(phrases.GetErrorPhrase("panel_groups_edit_cannot_designate_supermod"), w, r, user)
			}
			err = group.ChangeRank(false, true, false)
		case "Banned":
			err = group.ChangeRank(false, false, true)
		case "Guest":
			return c.LocalError(phrases.GetErrorPhrase("panel_groups_cannot_be_guest"), w, r, user)
		case "Member":
			err = group.ChangeRank(false, false, false)
		default:
			return c.LocalError(phrases.GetErrorPhrase("panel_groups_invalid_group_type"), w, r, user)
		}
		if err != nil {
			return c.InternalError(err, w, r)
		}
	}

	err = group.Update(gname, gtag)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func GroupsEditPermsSubmit(w http.ResponseWriter, r *http.Request, user c.User, sgid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return c.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return c.LocalError(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}

	group, err := c.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters o.o")
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return c.LocalError(phrases.GetErrorPhrase("panel_groups_cannot_edit_admin"), w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return c.LocalError(phrases.GetErrorPhrase("panel_groups_cannot_edit_supermod"), w, r, user)
	}

	var pmap = make(map[string]bool)

	if user.Perms.EditGroupLocalPerms {
		for _, perm := range c.LocalPermList {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	if user.Perms.EditGroupGlobalPerms {
		for _, perm := range c.GlobalPermList {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	err = group.UpdatePerms(pmap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/groups/edit/perms/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func GroupsCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return c.NoPermissions(w, r, user)
	}

	groupName := r.PostFormValue("group-name")
	if groupName == "" {
		return c.LocalError(phrases.GetErrorPhrase("panel_groups_need_name"), w, r, user)
	}
	groupTag := r.PostFormValue("group-tag")

	var isAdmin, isMod, isBanned bool
	if user.Perms.EditGroupGlobalPerms {
		groupType := r.PostFormValue("group-type")
		if groupType == "Admin" {
			if !user.Perms.EditGroupAdmin {
				return c.LocalError(phrases.GetErrorPhrase("panel_groups_create_cannot_designate_admin"), w, r, user)
			}
			isAdmin = true
			isMod = true
		} else if groupType == "Mod" {
			if !user.Perms.EditGroupSuperMod {
				return c.LocalError(phrases.GetErrorPhrase("panel_groups_create_cannot_designate_supermod"), w, r, user)
			}
			isMod = true
		} else if groupType == "Banned" {
			isBanned = true
		}
	}

	gid, err := c.Groups.Create(groupName, groupTag, isAdmin, isMod, isBanned)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}
