package panel

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"../../common"
)

//routePanelWordFilter
func WordFilters(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "word_filters", "word-filters")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}

	filterList, err := common.WordFilters.GetAll()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pi := common.PanelPage{basePage, tList, filterList}
	return panelRenderTemplate("panel_word_filters", w, r, user, &pi)
}

//routePanelWordFiltersCreateSubmit
func WordFiltersCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	// ? - We're not doing a full sanitise here, as it would be useful if admins were able to put down rules for replacing things with HTML, etc.
	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return common.LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	err := common.WordFilters.Create(find, replacement)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	return panelSuccessRedirect("/panel/settings/word-filters/", w, r, isJs)
}

// TODO: Implement this as a non-JS fallback
//routePanelWordFiltersEdit
func WordFiltersEdit(w http.ResponseWriter, r *http.Request, user common.User, wfid string) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "edit_word_filter", "word-filters")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}
	_ = wfid

	pi := common.PanelPage{basePage, tList, nil}
	return panelRenderTemplate("panel_word_filters_edit", w, r, user, &pi)
}

//routePanelWordFiltersEditSubmit
func WordFiltersEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, wfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	// TODO: Either call it isJs or js rather than flip-flopping back and forth across the routes x.x
	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		return common.LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
	}

	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return common.LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	err = common.WordFilters.Update(id, find, replacement)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}

//routePanelWordFiltersDeleteSubmit
func WordFiltersDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, wfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		return common.LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
	}

	err = common.WordFilters.Delete(id)
	if err == sql.ErrNoRows {
		return common.LocalErrorJSQ("This word filter doesn't exist", w, r, user, isJs)
	}

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}
