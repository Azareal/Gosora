package panel

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func Forums(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "forums", "forums")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return c.NoPermissions(w, r, user)
	}
	basePage.Header.AddScript("Sortable-1.4.0/Sortable.min.js")
	basePage.Header.AddScriptAsync("panel_forums.js")

	// TODO: Paginate this?
	var forumList []interface{}
	forums, err := c.Forums.GetAll()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// ? - Should we generate something similar to the forumView? It might be a little overkill for a page which is rarely loaded in comparison to /forums/
	for _, forum := range forums {
		if forum.Name != "" && forum.ParentID == 0 {
			fadmin := c.ForumAdmin{forum.ID, forum.Name, forum.Desc, forum.Active, forum.Preset, forum.TopicCount, c.PresetToLang(forum.Preset)}
			if fadmin.Preset == "" {
				fadmin.Preset = "custom"
			}
			forumList = append(forumList, fadmin)
		}
	}

	if r.FormValue("created") == "1" {
		basePage.AddNotice("panel_forum_created")
	} else if r.FormValue("deleted") == "1" {
		basePage.AddNotice("panel_forum_deleted")
	} else if r.FormValue("updated") == "1" {
		basePage.AddNotice("panel_forum_updated")
	}

	pi := c.PanelPage{basePage, forumList, nil}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage,"","","panel_forums",&pi})
}

func ForumsCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return c.NoPermissions(w, r, user)
	}

	fname := r.PostFormValue("forum-name")
	fdesc := r.PostFormValue("forum-desc")
	fpreset := c.StripInvalidPreset(r.PostFormValue("forum-preset"))
	factive := r.PostFormValue("forum-active")
	active := (factive == "on" || factive == "1")

	_, err := c.Forums.Create(fname, fdesc, active, fpreset)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/forums/?created=1", http.StatusSeeOther)
	return nil
}

// TODO: Revamp this
func ForumsDelete(w http.ResponseWriter, r *http.Request, user c.User, sfid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "delete_forum", "forums")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return c.NoPermissions(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	forum, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	confirmMsg := phrases.GetTmplPhrasef("panel_forum_delete_are_you_sure", forum.Name)
	yousure := c.AreYouSure{"/panel/forums/delete/submit/" + strconv.Itoa(fid), confirmMsg}

	pi := c.PanelPage{basePage, tList, yousure}
	if c.RunPreRenderHook("pre_render_panel_delete_forum", w, r, &user, &pi) {
		return nil
	}
	return renderTemplate("panel_are_you_sure", w, r, basePage.Header, &pi)
}

func ForumsDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, sfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return c.NoPermissions(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}
	err = c.Forums.Delete(fid)
	if err == sql.ErrNoRows {
		return c.LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/forums/?deleted=1", http.StatusSeeOther)
	return nil
}

func ForumsOrderSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageForums {
		return c.NoPermissionsJSQ(w, r, user, isJs)
	}
	sitems := strings.TrimSuffix(strings.TrimPrefix(r.PostFormValue("items"), "{"), "}")
	//fmt.Printf("sitems: %+v\n", sitems)

	var updateMap = make(map[int]int)
	for index, sfid := range strings.Split(sitems, ",") {
		fid, err := strconv.Atoi(sfid)
		if err != nil {
			return c.LocalErrorJSQ("Invalid integer in forum list", w, r, user, isJs)
		}
		updateMap[fid] = index
	}
	c.Forums.UpdateOrder(updateMap)

	return successRedirect("/panel/forums/", w, r, isJs)
}

func ForumsEdit(w http.ResponseWriter, r *http.Request, user c.User, sfid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "edit_forum", "forums")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return c.NoPermissions(w, r, user)
	}
	basePage.Header.AddScriptAsync("panel_forum_edit.js")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	forum, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if forum.Preset == "" {
		forum.Preset = "custom"
	}

	glist, err := c.Groups.GetAll()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var gplist []c.GroupForumPermPreset
	for gid, group := range glist {
		if gid == 0 {
			continue
		}
		forumPerms, err := c.FPStore.Get(fid, group.ID)
		if err == sql.ErrNoRows {
			forumPerms = c.BlankForumPerms()
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
		preset := c.ForumPermsToGroupForumPreset(forumPerms)
		gplist = append(gplist, c.GroupForumPermPreset{group, preset, preset == "default"})
	}

	if r.FormValue("updated") == "1" {
		basePage.AddNotice("panel_forum_updated")
	}

	pi := c.PanelEditForumPage{basePage, forum.ID, forum.Name, forum.Desc, forum.Active, forum.Preset, gplist}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage,"","","panel_forum_edit",&pi})
}

func ForumsEditSubmit(w http.ResponseWriter, r *http.Request, user c.User, sfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return c.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, user, isJs)
	}

	forum, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("The forum you're trying to edit doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}

	forumName := r.PostFormValue("forum_name")
	forumDesc := r.PostFormValue("forum_desc")
	forumPreset := c.StripInvalidPreset(r.PostFormValue("forum_preset"))
	forumActive := r.PostFormValue("forum_active")

	var active = false
	if forumActive == "" {
		active = forum.Active
	} else if forumActive == "1" || forumActive == "Show" {
		active = true
	}

	err = forum.Update(forumName, forumDesc, active, forumPreset)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}
	// ? Should we redirect to the forum editor instead?
	return successRedirect("/panel/forums/", w, r, isJs)
}

func ForumsEditPermsSubmit(w http.ResponseWriter, r *http.Request, user c.User, sfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return c.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, user, isJs)
	}

	gid, err := strconv.Atoi(r.PostFormValue("gid"))
	if err != nil {
		return c.LocalErrorJSQ("Invalid Group ID", w, r, user, isJs)
	}

	forum, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("This forum doesn't exist", w, r, user, isJs)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}

	permPreset := c.StripInvalidGroupForumPreset(r.PostFormValue("perm_preset"))
	err = forum.SetPreset(permPreset, gid)
	if err != nil {
		return c.LocalErrorJSQ(err.Error(), w, r, user, isJs)
	}

	return successRedirect("/panel/forums/edit/"+strconv.Itoa(fid)+"?updated=1", w, r, isJs)
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

func ForumsEditPermsAdvance(w http.ResponseWriter, r *http.Request, user c.User, paramList string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "edit_forum", "forums")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return c.NoPermissions(w, r, user)
	}

	fid, gid, err := forumPermsExtractDash(paramList)
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}

	forum, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	if forum.Preset == "" {
		forum.Preset = "custom"
	}

	forumPerms, err := c.FPStore.Get(fid, gid)
	if err == sql.ErrNoRows {
		forumPerms = c.BlankForumPerms()
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	var formattedPermList []c.NameLangToggle

	// TODO: Load the phrases in bulk for efficiency?
	// TODO: Reduce the amount of code duplication between this and the group editor. Also, can we grind this down into one line or use a code generator to stay current more easily?
	var addNameLangToggle = func(permStr string, perm bool) {
		formattedPermList = append(formattedPermList, c.NameLangToggle{permStr, phrases.GetLocalPermPhrase(permStr), perm})
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
		basePage.AddNotice("panel_forum_perms_updated")
	}

	pi := c.PanelEditForumGroupPage{basePage, forum.ID, gid, forum.Name, forum.Desc, forum.Active, forum.Preset, formattedPermList}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage,"","","panel_forum_edit_perms",&pi})
}

func ForumsEditPermsAdvanceSubmit(w http.ResponseWriter, r *http.Request, user c.User, paramList string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return c.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, gid, err := forumPermsExtractDash(paramList)
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}

	forum, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	forumPerms, err := c.FPStore.GetCopy(fid, gid)
	if err == sql.ErrNoRows {
		forumPerms = *c.BlankForumPerms()
	} else if err != nil {
		return c.InternalError(err, w, r)
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
		return c.LocalErrorJSQ(err.Error(), w, r, user, isJs)
	}

	return successRedirect("/panel/forums/edit/perms/"+strconv.Itoa(fid)+"-"+strconv.Itoa(gid)+"?updated=1", w, r, isJs)
}
