package panel

import (
	"database/sql"
	"net/http"
	"strconv"

	"../../common"
)

// routePanelGroups
func Groups(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = common.GetTitlePhrase("panel_groups")

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 9
	offset, page, lastPage := common.PageOffset(stats.Groups, page, perPage)

	// Skip the 'Unknown' group
	offset++

	var count int
	var groupList []common.GroupAdmin
	groups, _ := common.Groups.GetRange(offset, 0)
	for _, group := range groups {
		if count == perPage {
			break
		}

		var rank string
		var rankClass string
		var canEdit bool
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

		canEdit = user.Perms.EditGroup && (!group.IsAdmin || user.Perms.EditGroupAdmin) && (!group.IsMod || user.Perms.EditGroupSuperMod)
		groupList = append(groupList, common.GroupAdmin{group.ID, group.Name, rank, rankClass, canEdit, canDelete})
		count++
	}
	//log.Printf("groupList: %+v\n", groupList)

	pageList := common.Paginate(stats.Groups, perPage, 5)
	pi := common.PanelGroupPage{&common.BasePanelPage{header, stats, "groups", common.ReportForumID}, groupList, common.Paginator{pageList, page, lastPage}}
	return panelRenderTemplate("panel_groups", w, r, user, &pi)
}

//routePanelGroupsEdit
func GroupsEdit(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}
	header.Title = common.GetTitlePhrase("panel_edit_group")

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r, header)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
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

	pi := common.PanelEditGroupPage{&common.BasePanelPage{header, stats, "groups", common.ReportForumID}, group.ID, group.Name, group.Tag, rank, disableRank}
	if common.RunPreRenderHook("pre_render_panel_edit_group", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel_group_edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

//routePanelGroupsEditPerms
func GroupsEditPerms(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}
	header.Title = common.GetTitlePhrase("panel_edit_group")

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("The Group ID is not a valid integer.", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r, header)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	// TODO: Load the phrases in bulk for efficiency?
	var localPerms []common.NameLangToggle

	var addLocalPerm = func(permStr string, perm bool) {
		localPerms = append(localPerms, common.NameLangToggle{permStr, common.GetLocalPermPhrase(permStr), perm})
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

	var globalPerms []common.NameLangToggle
	var addGlobalPerm = func(permStr string, perm bool) {
		globalPerms = append(globalPerms, common.NameLangToggle{permStr, common.GetGlobalPermPhrase(permStr), perm})
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

	pi := common.PanelEditGroupPermsPage{&common.BasePanelPage{header, stats, "groups", common.ReportForumID}, group.ID, group.Name, localPerms, globalPerms}
	if common.RunPreRenderHook("pre_render_panel_edit_group_perms", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel_group_edit_perms.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

//routePanelGroupsEditSubmit
func GroupsEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r, nil)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	gname := r.FormValue("group-name")
	if gname == "" {
		return common.LocalError("The group name can't be left blank.", w, r, user)
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

	if rank != originalRank {
		if !user.Perms.EditGroupGlobalPerms {
			return common.LocalError("You need the EditGroupGlobalPerms permission to change the group type.", w, r, user)
		}

		switch rank {
		case "Admin":
			if !user.Perms.EditGroupAdmin {
				return common.LocalError("You need the EditGroupAdmin permission to designate this group as an admin group.", w, r, user)
			}
			err = group.ChangeRank(true, true, false)
		case "Mod":
			if !user.Perms.EditGroupSuperMod {
				return common.LocalError("You need the EditGroupSuperMod permission to designate this group as a super-mod group.", w, r, user)
			}
			err = group.ChangeRank(false, true, false)
		case "Banned":
			err = group.ChangeRank(false, false, true)
		case "Guest":
			return common.LocalError("You can't designate a group as a guest group.", w, r, user)
		case "Member":
			err = group.ChangeRank(false, false, false)
		default:
			return common.LocalError("Invalid group type.", w, r, user)
		}
		if err != nil {
			return common.InternalError(err, w, r)
		}
	}

	err = group.Update(gname, gtag)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

//routePanelGroupsEditPermsSubmit
func GroupsEditPermsSubmit(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("The Group ID is not a valid integer.", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters o.o")
		return common.NotFound(w, r, nil)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	var pmap = make(map[string]bool)

	if user.Perms.EditGroupLocalPerms {
		for _, perm := range common.LocalPermList {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	if user.Perms.EditGroupGlobalPerms {
		for _, perm := range common.GlobalPermList {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	err = group.UpdatePerms(pmap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/groups/edit/perms/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

//routePanelGroupsCreateSubmit
func GroupsCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	groupName := r.PostFormValue("group-name")
	if groupName == "" {
		return common.LocalError("You need a name for this group!", w, r, user)
	}
	groupTag := r.PostFormValue("group-tag")

	var isAdmin, isMod, isBanned bool
	if user.Perms.EditGroupGlobalPerms {
		groupType := r.PostFormValue("group-type")
		if groupType == "Admin" {
			if !user.Perms.EditGroupAdmin {
				return common.LocalError("You need the EditGroupAdmin permission to create admin groups", w, r, user)
			}
			isAdmin = true
			isMod = true
		} else if groupType == "Mod" {
			if !user.Perms.EditGroupSuperMod {
				return common.LocalError("You need the EditGroupSuperMod permission to create admin groups", w, r, user)
			}
			isMod = true
		} else if groupType == "Banned" {
			isBanned = true
		}
	}

	gid, err := common.Groups.Create(groupName, groupTag, isAdmin, isMod, isBanned)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}
