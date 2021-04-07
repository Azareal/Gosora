package panel

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	p "github.com/Azareal/Gosora/common/phrases"
)

func Forums(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, u, "forums", "forums")
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
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
	for _, f := range forums {
		if f.Name != "" && f.ParentID == 0 {
			fadmin := c.ForumAdmin{f.ID, f.Name, f.Desc, f.Active, f.Preset, f.TopicCount, c.PresetToLang(f.Preset)}
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
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_forums", &pi})
}

func ForumsCreateSubmit(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
	}

	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	preset := c.StripInvalidPreset(r.PostFormValue("preset"))
	factive := r.PostFormValue("active")
	active := (factive == "on" || factive == "1")

	fid, err := c.Forums.Create(name, desc, active, preset)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	err = c.AdminLogs.Create("create", fid, "forum", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/forums/?created=1", http.StatusSeeOther)
	return nil
}

// TODO: Revamp this
func ForumsDelete(w http.ResponseWriter, r *http.Request, u *c.User, sfid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, u, "delete_forum", "forums")
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.LocalError("The provided Forum ID is not a valid number.", w, r, u)
	}
	forum, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalError("The forum you're trying to delete doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	confirmMsg := p.GetTmplPhrasef("panel_forum_delete_are_you_sure", forum.Name)
	youSure := c.AreYouSure{"/panel/forums/delete/submit/" + strconv.Itoa(fid), confirmMsg}

	pi := c.PanelPage{basePage, tList, youSure}
	if c.RunPreRenderHook("pre_render_panel_delete_forum", w, r, u, &pi) {
		return nil
	}
	return renderTemplate("panel_are_you_sure", w, r, basePage.Header, &pi)
}

func ForumsDeleteSubmit(w http.ResponseWriter, r *http.Request, u *c.User, sfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.LocalError("The provided Forum ID is not a valid number.", w, r, u)
	}
	err = c.Forums.Delete(fid)
	if err == sql.ErrNoRows {
		return c.LocalError("The forum you're trying to delete doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	err = c.AdminLogs.Create("delete", fid, "forum", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/forums/?deleted=1", http.StatusSeeOther)
	return nil
}

func ForumsOrderSubmit(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	// TODO: Move this even earlier?
	js := r.PostFormValue("js") == "1"
	if !u.Perms.ManageForums {
		return c.NoPermissionsJSQ(w, r, u, js)
	}
	sitems := strings.TrimSuffix(strings.TrimPrefix(r.PostFormValue("items"), "{"), "}")
	//fmt.Printf("sitems: %+v\n", sitems)

	updateMap := make(map[int]int)
	for index, sfid := range strings.Split(sitems, ",") {
		fid, err := strconv.Atoi(sfid)
		if err != nil {
			return c.LocalErrorJSQ("Invalid integer in forum list", w, r, u, js)
		}
		updateMap[fid] = index
	}
	err := c.Forums.UpdateOrder(updateMap)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	err = c.AdminLogs.Create("reorder", 0, "forum", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	return successRedirect("/panel/forums/", w, r, js)
}

func ForumsEdit(w http.ResponseWriter, r *http.Request, u *c.User, sfid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, u, "edit_forum", "forums")
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.SimpleError(p.GetErrorPhrase("url_id_must_be_integer"), w, r, basePage.Header)
	}
	basePage.Header.AddScriptAsync("panel_forum_edit.js")

	f, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalError("The forum you're trying to edit doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if f.Preset == "" {
		f.Preset = "custom"
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

	falist, e := c.ForumActionStore.GetInForum(f.ID)
	if err != sql.ErrNoRows && e != nil {
		return c.InternalError(e, w, r)
	}
	afalist := make([]*c.ForumActionAction, len(falist))
	for i, faitem := range falist {
		afalist[i] = &c.ForumActionAction{faitem, c.ConvActToString(faitem.Action)}
	}

	pi := c.PanelEditForumPage{basePage, f.ID, f.Name, f.Desc, f.Active, f.Preset, gplist, afalist}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_forum_edit", &pi})
}

func ForumsEditSubmit(w http.ResponseWriter, r *http.Request, u *c.User, sfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
	}
	js := r.PostFormValue("js") == "1"

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, u, js)
	}
	forum, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("The forum you're trying to edit doesn't exist.", w, r, u, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	name := r.PostFormValue("forum_name")
	desc := r.PostFormValue("forum_desc")
	preset := c.StripInvalidPreset(r.PostFormValue("forum_preset"))
	factive := r.PostFormValue("forum_active")

	active := false
	if factive == "" {
		active = forum.Active
	} else if factive == "1" || factive == "Show" {
		active = true
	}

	err = forum.Update(name, desc, active, preset)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	err = c.AdminLogs.Create("edit", fid, "forum", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// ? Should we redirect to the forum editor instead?
	return successRedirect("/panel/forums/", w, r, js)
}

func ForumsEditPermsSubmit(w http.ResponseWriter, r *http.Request, u *c.User, sfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
	}
	js := r.PostFormValue("js") == "1"

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, u, js)
	}
	gid, err := strconv.Atoi(r.PostFormValue("gid"))
	if err != nil {
		return c.LocalErrorJSQ("Invalid Group ID", w, r, u, js)
	}

	f, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("This forum doesn't exist", w, r, u, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	permPreset := c.StripInvalidGroupForumPreset(r.PostFormValue("perm_preset"))
	err = f.SetPreset(permPreset, gid)
	if err != nil {
		return c.LocalErrorJSQ(err.Error(), w, r, u, js)
	}
	err = c.AdminLogs.Create("edit", fid, "forum", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	return successRedirect("/panel/forums/edit/"+strconv.Itoa(fid)+"?updated=1", w, r, js)
}

// A helper function for the Advanced portion of the Forum Perms Editor
func forumPermsExtractDash(paramList string) (fid, gid int, e error) {
	params := strings.Split(paramList, "-")
	if len(params) != 2 {
		return fid, gid, errors.New("Parameter count mismatch")
	}
	fid, e = strconv.Atoi(params[0])
	if e != nil {
		return fid, gid, errors.New("The provided Forum ID is not a valid number.")
	}
	gid, e = strconv.Atoi(params[1])
	if e != nil {
		e = errors.New("The provided Group ID is not a valid number.")
	}
	return fid, gid, e
}

func ForumsEditPermsAdvance(w http.ResponseWriter, r *http.Request, u *c.User, paramList string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, u, "edit_forum", "forums")
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
	}

	fid, gid, err := forumPermsExtractDash(paramList)
	if err != nil {
		return c.LocalError(err.Error(), w, r, u)
	}

	f, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalError("The forum you're trying to edit doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if f.Preset == "" {
		f.Preset = "custom"
	}

	fp, err := c.FPStore.Get(fid, gid)
	if err == sql.ErrNoRows {
		fp = c.BlankForumPerms()
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	var formattedPermList []c.NameLangToggle
	// TODO: Load the phrases in bulk for efficiency?
	// TODO: Reduce the amount of code duplication between this and the group editor. Also, can we grind this down into one line or use a code generator to stay current more easily?
	addToggle := func(permStr string, perm bool) {
		formattedPermList = append(formattedPermList, c.NameLangToggle{permStr, p.GetPermPhrase(permStr), perm})
	}
	addToggle("ViewTopic", fp.ViewTopic)
	addToggle("LikeItem", fp.LikeItem)
	addToggle("CreateTopic", fp.CreateTopic)
	//<--
	addToggle("EditTopic", fp.EditTopic)
	addToggle("DeleteTopic", fp.DeleteTopic)
	addToggle("CreateReply", fp.CreateReply)
	addToggle("EditReply", fp.EditReply)
	addToggle("DeleteReply", fp.DeleteReply)
	addToggle("PinTopic", fp.PinTopic)
	addToggle("CloseTopic", fp.CloseTopic)
	addToggle("MoveTopic", fp.MoveTopic)

	if r.FormValue("updated") == "1" {
		basePage.AddNotice("panel_forum_perms_updated")
	}

	pi := c.PanelEditForumGroupPage{basePage, f.ID, gid, f.Name, f.Desc, f.Active, f.Preset, formattedPermList}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_forum_edit_perms", &pi})
}

func ForumsEditPermsAdvanceSubmit(w http.ResponseWriter, r *http.Request, u *c.User, paramList string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
	}
	js := r.PostFormValue("js") == "1"

	fid, gid, err := forumPermsExtractDash(paramList)
	if err != nil {
		return c.LocalError(err.Error(), w, r, u)
	}

	f, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.LocalError("The forum you're trying to edit doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	fp, err := c.FPStore.GetCopy(fid, gid)
	if err == sql.ErrNoRows {
		fp = *c.BlankForumPerms()
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	ep := func(name string) bool {
		pvalue := r.PostFormValue("perm-" + name)
		return (pvalue == "1")
	}
	// TODO: Generate this code?
	fp.ViewTopic = ep("ViewTopic")
	fp.LikeItem = ep("LikeItem")
	fp.CreateTopic = ep("CreateTopic")
	fp.EditTopic = ep("EditTopic")
	fp.DeleteTopic = ep("DeleteTopic")
	fp.CreateReply = ep("CreateReply")
	fp.EditReply = ep("EditReply")
	fp.DeleteReply = ep("DeleteReply")
	fp.PinTopic = ep("PinTopic")
	fp.CloseTopic = ep("CloseTopic")
	fp.MoveTopic = ep("MoveTopic")

	err = f.SetPerms(&fp, "custom", gid)
	if err != nil {
		return c.LocalErrorJSQ(err.Error(), w, r, u, js)
	}
	err = c.AdminLogs.Create("edit", fid, "forum", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	return successRedirect("/panel/forums/edit/perms/"+strconv.Itoa(fid)+"-"+strconv.Itoa(gid)+"?updated=1", w, r, js)
}

func ForumsEditActionDeleteSubmit(w http.ResponseWriter, r *http.Request, u *c.User, sfaid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	// TODO: Should we split this permission?
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
	}
	js := r.PostFormValue("js") == "1"

	faid, e := strconv.Atoi(sfaid)
	if e != nil {
		return c.LocalError("The forum action ID is not a valid integer.", w, r, u)
	}
	e = c.ForumActionStore.Delete(faid)
	if e != nil {
		return c.InternalError(e, w, r)
	}

	fid, e := strconv.Atoi(r.FormValue("ret"))
	if e != nil {
		return c.LocalError("The forum action ID is not a valid integer.", w, r, u)
	}
	if !c.Forums.Exists(fid) {
		return c.LocalError("The target forum doesn't exist.", w, r, u)
	}

	return successRedirect("/panel/forums/edit/"+strconv.Itoa(fid)+"?updated=1", w, r, js)
}

func ForumsEditActionCreateSubmit(w http.ResponseWriter, r *http.Request, u *c.User, sfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	// TODO: Should we split this permission?
	if !u.Perms.ManageForums {
		return c.NoPermissions(w, r, u)
	}
	js := r.PostFormValue("js") == "1"

	fid, e := strconv.Atoi(sfid)
	if e != nil {
		return c.LocalError("The provided Forum ID is not a valid number.", w, r, u)
	}
	if !c.Forums.Exists(fid) {
		return c.LocalError("This forum does not exist", w, r, u)
	}

	runOnTopicCreation := r.PostFormValue("action_run_on_topic_creation") == "1"

	f := func(s string) (int, c.RouteError) {
		i, e := strconv.Atoi(r.PostFormValue(s))
		if e != nil {
			return i, c.LocalError(s+" is not a valid integer.", w, r, u)
		}
		if i < 0 {
			return i, c.LocalError(s+" cannot be less than 0", w, r, u)
		}
		return i, nil
	}
	runDaysAfterTopicCreation, re := f("action_run_days_after_topic_creation")
	if re != nil {
		return re
	}
	runDaysAfterTopicLastReply, re := f("action_run_days_after_topic_last_reply")
	if re != nil {
		return re
	}

	action := r.PostFormValue("action_action")
	aint := c.ConvStringToAct(action)
	if aint == -1 {
		return c.LocalError("invalid action", w, r, u)
	}

	extra := r.PostFormValue("action_extra")
	switch aint {
	case c.ForumActionMove:
		conv, e := strconv.Atoi(extra)
		if e != nil {
			return c.LocalError("action_extra is not a valid integer.", w, r, u)
		}
		extra = strconv.Itoa(conv)
	default:
		extra = ""
	}

	_, e = c.ForumActionStore.Add(&c.ForumAction{
		Forum:                      fid,
		RunOnTopicCreation:         runOnTopicCreation,
		RunDaysAfterTopicCreation:  runDaysAfterTopicCreation,
		RunDaysAfterTopicLastReply: runDaysAfterTopicLastReply,
		Action:                     aint,
		Extra:                      extra,
	})
	if e != nil {
		return c.InternalError(e, w, r)
	}

	return successRedirect("/panel/forums/edit/"+strconv.Itoa(fid)+"?updated=1", w, r, js)
}
