package panel

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"../../common"
)

func Forums(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	// TODO: Paginate this?
	var forumList []interface{}
	forums, err := common.Forums.GetAll()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// ? - Should we generate something similar to the forumView? It might be a little overkill for a page which is rarely loaded in comparison to /forums/
	for _, forum := range forums {
		if forum.Name != "" && forum.ParentID == 0 {
			fadmin := common.ForumAdmin{forum.ID, forum.Name, forum.Desc, forum.Active, forum.Preset, forum.TopicCount, common.PresetToLang(forum.Preset)}
			if fadmin.Preset == "" {
				fadmin.Preset = "custom"
			}
			forumList = append(forumList, fadmin)
		}
	}

	if r.FormValue("created") == "1" {
		headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("panel_forum_created"))
	} else if r.FormValue("deleted") == "1" {
		headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("panel_forum_deleted"))
	} else if r.FormValue("updated") == "1" {
		headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("panel_forum_updated"))
	}

	pi := common.PanelPage{common.GetTitlePhrase("panel_forums"), user, headerVars, stats, "forums", forumList, nil}
	return panelRenderTemplate("panel_forums", w, r, user, &pi)
}

func ForumsCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	fname := r.PostFormValue("forum-name")
	fdesc := r.PostFormValue("forum-desc")
	fpreset := common.StripInvalidPreset(r.PostFormValue("forum-preset"))
	factive := r.PostFormValue("forum-active")
	active := (factive == "on" || factive == "1")

	_, err := common.Forums.Create(fname, fdesc, active, fpreset)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/forums/?created=1", http.StatusSeeOther)
	return nil
}

// TODO: Revamp this
func ForumsDelete(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	forum, err := common.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return common.LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Make this a phrase
	confirmMsg := "Are you sure you want to delete the '" + forum.Name + "' forum?"
	yousure := common.AreYouSure{"/panel/forums/delete/submit/" + strconv.Itoa(fid), confirmMsg}

	pi := common.PanelPage{common.GetTitlePhrase("panel_delete_forum"), user, headerVars, stats, "forums", tList, yousure}
	if common.RunPreRenderHook("pre_render_panel_delete_forum", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "are_you_sure.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func ForumsDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	err = common.Forums.Delete(fid)
	if err == sql.ErrNoRows {
		return common.LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/forums/?deleted=1", http.StatusSeeOther)
	return nil
}

func ForumsEdit(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	forum, err := common.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return common.LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if forum.Preset == "" {
		forum.Preset = "custom"
	}

	glist, err := common.Groups.GetAll()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var gplist []common.GroupForumPermPreset
	for gid, group := range glist {
		if gid == 0 {
			continue
		}
		forumPerms, err := common.FPStore.Get(fid, group.ID)
		if err == sql.ErrNoRows {
			forumPerms = common.BlankForumPerms()
		} else if err != nil {
			return common.InternalError(err, w, r)
		}
		gplist = append(gplist, common.GroupForumPermPreset{group, common.ForumPermsToGroupForumPreset(forumPerms)})
	}

	if r.FormValue("updated") == "1" {
		headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("panel_forum_updated"))
	}

	pi := common.PanelEditForumPage{common.GetTitlePhrase("panel_edit_forum"), user, headerVars, stats, "forums", forum.ID, forum.Name, forum.Desc, forum.Active, forum.Preset, gplist}
	if common.RunPreRenderHook("pre_render_panel_edit_forum", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel-forum-edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	return nil
}

func ForumsEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, user, isJs)
	}

	forum, err := common.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return common.LocalErrorJSQ("The forum you're trying to edit doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	forumName := r.PostFormValue("forum_name")
	forumDesc := r.PostFormValue("forum_desc")
	forumPreset := common.StripInvalidPreset(r.PostFormValue("forum_preset"))
	forumActive := r.PostFormValue("forum_active")

	var active = false
	if forumActive == "" {
		active = forum.Active
	} else if forumActive == "1" || forumActive == "Show" {
		active = true
	}

	err = forum.Update(forumName, forumDesc, active, forumPreset)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	// ? Should we redirect to the forum editor instead?
	return panelSuccessRedirect("/panel/forums/", w, r, isJs)
}

func ForumsEditPermsSubmit(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, user, isJs)
	}

	gid, err := strconv.Atoi(r.PostFormValue("gid"))
	if err != nil {
		return common.LocalErrorJSQ("Invalid Group ID", w, r, user, isJs)
	}

	forum, err := common.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return common.LocalErrorJSQ("This forum doesn't exist", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	permPreset := common.StripInvalidGroupForumPreset(r.PostFormValue("perm_preset"))
	err = forum.SetPreset(permPreset, gid)
	if err != nil {
		return common.LocalErrorJSQ(err.Error(), w, r, user, isJs)
	}

	return panelSuccessRedirect("/panel/forums/edit/"+strconv.Itoa(fid)+"?updated=1", w, r, isJs)
}

// A helper function for the Advanced portion of the Forum Perms Editor
func forumPermsExtractDash(paramList string) (fid int, gid int, err error) {
	params := strings.Split(paramList, "-")
	if len(params) != 2 {
		return fid, gid, errors.New("Parameter count mismatch")
	}

	fid, err = strconv.Atoi(params[0])
	if err != nil {
		return fid, gid, errors.New("The provided Forum ID is not a valid number.")
	}

	gid, err = strconv.Atoi(params[1])
	if err != nil {
		err = errors.New("The provided Group ID is not a valid number.")
	}

	return fid, gid, err
}

func ForumsEditPermsAdvance(w http.ResponseWriter, r *http.Request, user common.User, paramList string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	fid, gid, err := forumPermsExtractDash(paramList)
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	forum, err := common.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return common.LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if forum.Preset == "" {
		forum.Preset = "custom"
	}

	forumPerms, err := common.FPStore.Get(fid, gid)
	if err == sql.ErrNoRows {
		forumPerms = common.BlankForumPerms()
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	var formattedPermList []common.NameLangToggle

	// TODO: Load the phrases in bulk for efficiency?
	// TODO: Reduce the amount of code duplication between this and the group editor. Also, can we grind this down into one line or use a code generator to stay current more easily?
	var addNameLangToggle = func(permStr string, perm bool) {
		formattedPermList = append(formattedPermList, common.NameLangToggle{permStr, common.GetLocalPermPhrase(permStr), perm})
	}
	addNameLangToggle("ViewTopic", forumPerms.ViewTopic)
	addNameLangToggle("LikeItem", forumPerms.LikeItem)
	addNameLangToggle("CreateTopic", forumPerms.CreateTopic)
	//<--
	addNameLangToggle("EditTopic", forumPerms.EditTopic)
	addNameLangToggle("DeleteTopic", forumPerms.DeleteTopic)
	addNameLangToggle("CreateReply", forumPerms.CreateReply)
	addNameLangToggle("EditReply", forumPerms.EditReply)
	addNameLangToggle("DeleteReply", forumPerms.DeleteReply)
	addNameLangToggle("PinTopic", forumPerms.PinTopic)
	addNameLangToggle("CloseTopic", forumPerms.CloseTopic)
	addNameLangToggle("MoveTopic", forumPerms.MoveTopic)

	if r.FormValue("updated") == "1" {
		headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("panel_forums_perms_updated"))
	}

	pi := common.PanelEditForumGroupPage{common.GetTitlePhrase("panel_edit_forum"), user, headerVars, stats, "forums", forum.ID, gid, forum.Name, forum.Desc, forum.Active, forum.Preset, formattedPermList}
	if common.RunPreRenderHook("pre_render_panel_edit_forum", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel-forum-edit-perms.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	return nil
}

func ForumsEditPermsAdvanceSubmit(w http.ResponseWriter, r *http.Request, user common.User, paramList string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, gid, err := forumPermsExtractDash(paramList)
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	forum, err := common.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return common.LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	forumPerms, err := common.FPStore.GetCopy(fid, gid)
	if err == sql.ErrNoRows {
		forumPerms = *common.BlankForumPerms()
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	var extractPerm = func(name string) bool {
		pvalue := r.PostFormValue("forum-perm-" + name)
		return (pvalue == "1")
	}

	// TODO: Generate this code?
	forumPerms.ViewTopic = extractPerm("ViewTopic")
	forumPerms.LikeItem = extractPerm("LikeItem")
	forumPerms.CreateTopic = extractPerm("CreateTopic")
	forumPerms.EditTopic = extractPerm("EditTopic")
	forumPerms.DeleteTopic = extractPerm("DeleteTopic")
	forumPerms.CreateReply = extractPerm("CreateReply")
	forumPerms.EditReply = extractPerm("EditReply")
	forumPerms.DeleteReply = extractPerm("DeleteReply")
	forumPerms.PinTopic = extractPerm("PinTopic")
	forumPerms.CloseTopic = extractPerm("CloseTopic")
	forumPerms.MoveTopic = extractPerm("MoveTopic")

	err = forum.SetPerms(&forumPerms, "custom", gid)
	if err != nil {
		return common.LocalErrorJSQ(err.Error(), w, r, user, isJs)
	}

	return panelSuccessRedirect("/panel/forums/edit/perms/"+strconv.Itoa(fid)+"-"+strconv.Itoa(gid)+"?updated=1", w, r, isJs)
}
