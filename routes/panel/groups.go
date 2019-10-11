package panel

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	p "github.com/Azareal/Gosora/common/phrases"
)

func Groups(w http.ResponseWriter, r *http.Request, u c.User) c.RouteError {
	bPage, ferr := buildBasePage(w, r, &u, "groups", "groups")
	if ferr != nil {
		return ferr
	}
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 15
	offset, page, lastPage := c.PageOffset(bPage.Stats.Groups, page, perPage)

	// Skip the 'Unknown' group
	offset++

	var count int
	var groupList []c.GroupAdmin
	groups, _ := c.Groups.GetRange(offset, 0)
	for _, g := range groups {
		if count == perPage {
			break
		}
		var rank string
		var rankClass string
		canDelete := false

		// TODO: Localise this
		switch {
		case g.IsAdmin:
			rank = "Admin"
			rankClass = "admin"
		case g.IsMod:
			rank = "Mod"
			rankClass = "mod"
		case g.IsBanned:
			rank = "Banned"
			rankClass = "banned"
		case g.ID == 6:
			rank = "Guest"
			rankClass = "guest"
		default:
			rank = "Member"
			rankClass = "member"
		}

		canEdit := u.Perms.EditGroup && (!g.IsAdmin || u.Perms.EditGroupAdmin) && (!g.IsMod || u.Perms.EditGroupSuperMod)
		groupList = append(groupList, c.GroupAdmin{g.ID, g.Name, rank, rankClass, canEdit, canDelete})
		count++
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelGroupPage{bPage, groupList, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, bPage.Header, c.Panel{bPage, "", "", "panel_groups", &pi})
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
		return c.LocalError(p.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}
	g, err := c.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters")
		return c.NotFound(w, r, basePage.Header)
	}
	ferr = groupCheck(w, r, user, g, err)
	if ferr != nil {
		return ferr
	}

	var rank string
	switch {
	case g.IsAdmin:
		rank = "Admin"
	case g.IsMod:
		rank = "Mod"
	case g.IsBanned:
		rank = "Banned"
	case g.ID == 6:
		rank = "Guest"
	default:
		rank = "Member"
	}
	disableRank := !user.Perms.EditGroupGlobalPerms || (g.ID == 6)

	pi := c.PanelEditGroupPage{basePage, g.ID, g.Name, g.Tag, rank, disableRank}
	return renderTemplate("panel_group_edit", w, r, basePage.Header, pi)
}

func GroupsEditPromotions(w http.ResponseWriter, r *http.Request, user c.User, sgid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "edit_group", "groups")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return c.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}
	g, err := c.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters")
		return c.NotFound(w, r, basePage.Header)
	}
	ferr = groupCheck(w, r, user, g, err)
	if ferr != nil {
		return ferr
	}

	promotions, err := c.GroupPromotions.GetByGroup(g.ID)
	if err != sql.ErrNoRows && err != nil {
		return c.InternalError(err, w, r)
	}
	promoteExt := make([]*c.GroupPromotionExtend, len(promotions))
	for i, promote := range promotions {
		fg, err := c.Groups.Get(promote.From)
		if err == sql.ErrNoRows {
			fg = &c.Group{Name: "Deleted Group"}
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
		tg, err := c.Groups.Get(promote.To)
		if err == sql.ErrNoRows {
			tg = &c.Group{Name: "Deleted Group"}
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
		promoteExt[i] = &c.GroupPromotionExtend{promote, fg, tg}
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

	pi := c.PanelEditGroupPromotionsPage{basePage, g.ID, g.Name, promoteExt, groupList}
	return renderTemplate("panel_group_edit_promotions", w, r, basePage.Header, pi)
}

func groupCheck(w http.ResponseWriter, r *http.Request, user c.User, g *c.Group, err error) c.RouteError {
	if err == sql.ErrNoRows {
		return c.LocalError("No such group.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if g.IsAdmin && !user.Perms.EditGroupAdmin {
		return c.LocalError(p.GetErrorPhrase("panel_groups_cannot_edit_admin"), w, r, user)
	}
	if g.IsMod && !user.Perms.EditGroupSuperMod {
		return c.LocalError(p.GetErrorPhrase("panel_groups_cannot_edit_supermod"), w, r, user)
	}
	return nil
}

func GroupsPromotionsCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User, sgid string) c.RouteError {
	if !user.Perms.EditGroup {
		return c.NoPermissions(w, r, user)
	}
	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	from, err := strconv.Atoi(r.FormValue("from"))
	if err != nil {
		return c.LocalError("from must be integer", w, r, user)
	}
	to, err := strconv.Atoi(r.FormValue("to"))
	if err != nil {
		return c.LocalError("to must be integer", w, r, user)
	}
	if from == to {
		return c.LocalError("the from group and to group cannot be the same", w, r, user)
	}
	twoWay := r.FormValue("two-way") == "1"

	level, err := strconv.Atoi(r.FormValue("level"))
	if err != nil {
		return c.LocalError("level must be integer", w, r, user)
	}
	posts, err := strconv.Atoi(r.FormValue("posts"))
	if err != nil {
		return c.LocalError("posts must be integer", w, r, user)
	}

	g, err := c.Groups.Get(from)
	ferr := groupCheck(w, r, user, g, err)
	if err != nil {
		return ferr
	}
	g, err = c.Groups.Get(to)
	ferr = groupCheck(w, r, user, g, err)
	if err != nil {
		return ferr
	}
	_, err = c.GroupPromotions.Create(from, to, twoWay, level, posts)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/groups/edit/promotions/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func GroupsPromotionsDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, sspl string) c.RouteError {
	if !user.Perms.EditGroup {
		return c.NoPermissions(w, r, user)
	}
	spl := strings.Split(sspl, "-")
	if len(spl) < 2 {
		return c.LocalError("need two params", w, r, user)
	}
	gid, err := strconv.Atoi(spl[0])
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}
	pid, err := strconv.Atoi(spl[1])
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	pro, err := c.GroupPromotions.Get(pid)
	if err == sql.ErrNoRows {
		return c.LocalError("That group promotion doesn't exist", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	g, err := c.Groups.Get(pro.From)
	ferr := groupCheck(w, r, user, g, err)
	if err != nil {
		return ferr
	}
	g, err = c.Groups.Get(pro.To)
	ferr = groupCheck(w, r, user, g, err)
	if err != nil {
		return ferr
	}
	err = c.GroupPromotions.Delete(pid)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/groups/edit/promotions/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
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
		return c.LocalError(p.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	g, err := c.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters")
		return c.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if g.IsAdmin && !user.Perms.EditGroupAdmin {
		return c.LocalError(p.GetErrorPhrase("panel_groups_cannot_edit_admin"), w, r, user)
	}
	if g.IsMod && !user.Perms.EditGroupSuperMod {
		return c.LocalError(p.GetErrorPhrase("panel_groups_cannot_edit_supermod"), w, r, user)
	}

	// TODO: Load the phrases in bulk for efficiency?
	var localPerms []c.NameLangToggle
	addPerm := func(permStr string, perm bool) {
		localPerms = append(localPerms, c.NameLangToggle{permStr, p.GetPermPhrase(permStr), perm})
	}

	addPerm("ViewTopic", g.Perms.ViewTopic)
	addPerm("LikeItem", g.Perms.LikeItem)
	addPerm("CreateTopic", g.Perms.CreateTopic)
	//<--
	addPerm("EditTopic", g.Perms.EditTopic)
	addPerm("DeleteTopic", g.Perms.DeleteTopic)
	addPerm("CreateReply", g.Perms.CreateReply)
	addPerm("EditReply", g.Perms.EditReply)
	addPerm("DeleteReply", g.Perms.DeleteReply)
	addPerm("PinTopic", g.Perms.PinTopic)
	addPerm("CloseTopic", g.Perms.CloseTopic)
	addPerm("MoveTopic", g.Perms.MoveTopic)

	var globalPerms []c.NameLangToggle
	addPerm = func(permStr string, perm bool) {
		globalPerms = append(globalPerms, c.NameLangToggle{permStr, p.GetPermPhrase(permStr), perm})
	}

	addPerm("BanUsers", g.Perms.BanUsers)
	addPerm("ActivateUsers", g.Perms.ActivateUsers)
	addPerm("EditUser", g.Perms.EditUser)
	addPerm("EditUserEmail", g.Perms.EditUserEmail)
	addPerm("EditUserPassword", g.Perms.EditUserPassword)
	addPerm("EditUserGroup", g.Perms.EditUserGroup)
	addPerm("EditUserGroupSuperMod", g.Perms.EditUserGroupSuperMod)
	addPerm("EditUserGroupAdmin", g.Perms.EditUserGroupAdmin)
	addPerm("EditGroup", g.Perms.EditGroup)
	addPerm("EditGroupLocalPerms", g.Perms.EditGroupLocalPerms)
	addPerm("EditGroupGlobalPerms", g.Perms.EditGroupGlobalPerms)
	addPerm("EditGroupSuperMod", g.Perms.EditGroupSuperMod)
	addPerm("EditGroupAdmin", g.Perms.EditGroupAdmin)
	addPerm("ManageForums", g.Perms.ManageForums)
	addPerm("EditSettings", g.Perms.EditSettings)
	addPerm("ManageThemes", g.Perms.ManageThemes)
	addPerm("ManagePlugins", g.Perms.ManagePlugins)
	addPerm("ViewAdminLogs", g.Perms.ViewAdminLogs)
	addPerm("ViewIPs", g.Perms.ViewIPs)
	addPerm("UploadFiles", g.Perms.UploadFiles)
	addPerm("UploadAvatars", g.Perms.UploadAvatars)
	addPerm("UseConvos", g.Perms.UseConvos)

	pi := c.PanelEditGroupPermsPage{basePage, g.ID, g.Name, localPerms, globalPerms}
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
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}

	group, err := c.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters")
		return c.NotFound(w, r, nil)
	}
	ferr = groupCheck(w, r, user, group, err)
	if ferr != nil {
		return ferr
	}

	gname := r.FormValue("group-name")
	if gname == "" {
		return c.LocalError(p.GetErrorPhrase("panel_groups_need_name"), w, r, user)
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
			return c.LocalError(p.GetErrorPhrase("panel_groups_cannot_edit_group_type"), w, r, user)
		}

		switch rank {
		case "Admin":
			if !user.Perms.EditGroupAdmin {
				return c.LocalError(p.GetErrorPhrase("panel_groups_edit_cannot_designate_admin"), w, r, user)
			}
			err = group.ChangeRank(true, true, false)
		case "Mod":
			if !user.Perms.EditGroupSuperMod {
				return c.LocalError(p.GetErrorPhrase("panel_groups_edit_cannot_designate_supermod"), w, r, user)
			}
			err = group.ChangeRank(false, true, false)
		case "Banned":
			err = group.ChangeRank(false, false, true)
		case "Guest":
			return c.LocalError(p.GetErrorPhrase("panel_groups_cannot_be_guest"), w, r, user)
		case "Member":
			err = group.ChangeRank(false, false, false)
		default:
			return c.LocalError(p.GetErrorPhrase("panel_groups_invalid_group_type"), w, r, user)
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
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}

	group, err := c.Groups.Get(gid)
	if err == sql.ErrNoRows {
		//log.Print("aaaaa monsters o.o")
		return c.NotFound(w, r, nil)
	}
	ferr = groupCheck(w, r, user, group, err)
	if ferr != nil {
		return ferr
	}

	// TODO: Don't unset perms we don't have permission to set?
	pmap := make(map[string]bool)
	pCheck := func(hasPerm bool, perms []string) {
		if hasPerm {
			for _, perm := range perms {
				pvalue := r.PostFormValue("group-perm-" + perm)
				pmap[perm] = (pvalue == "1")
			}
		}
	}
	pCheck(user.Perms.EditGroupLocalPerms, c.LocalPermList)
	pCheck(user.Perms.EditGroupGlobalPerms, c.GlobalPermList)

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
		return c.LocalError(p.GetErrorPhrase("panel_groups_need_name"), w, r, user)
	}
	groupTag := r.PostFormValue("group-tag")

	var isAdmin, isMod, isBanned bool
	if user.Perms.EditGroupGlobalPerms {
		switch r.PostFormValue("group-type") {
		case "Admin":
			if !user.Perms.EditGroupAdmin {
				return c.LocalError(p.GetErrorPhrase("panel_groups_create_cannot_designate_admin"), w, r, user)
			}
			isAdmin = true
			isMod = true
		case "Mod":
			if !user.Perms.EditGroupSuperMod {
				return c.LocalError(p.GetErrorPhrase("panel_groups_create_cannot_designate_supermod"), w, r, user)
			}
			isMod = true
		case "Banned":
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
