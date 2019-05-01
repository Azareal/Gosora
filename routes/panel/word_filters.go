package panel

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
)

func WordFilters(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "word_filters", "word-filters")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return c.NoPermissions(w, r, user)
	}

	// TODO: What if this list gets too long?
	filterList, err := c.WordFilters.GetAll()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.PanelPage{basePage, tList, filterList}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage,"","","panel_word_filters",&pi})
}

func WordFiltersCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return c.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	// ? - We're not doing a full sanitise here, as it would be useful if admins were able to put down rules for replacing things with HTML, etc.
	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return c.LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	err := c.WordFilters.Create(find, replacement)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}

	return successRedirect("/panel/settings/word-filters/", w, r, isJs)
}

// TODO: Implement this as a non-JS fallback
func WordFiltersEdit(w http.ResponseWriter, r *http.Request, user c.User, wfid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "edit_word_filter", "word-filters")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return c.NoPermissions(w, r, user)
	}
	_ = wfid

	pi := c.PanelPage{basePage, tList, nil}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage,"","","panel_word_filters_edit",&pi})
}

func WordFiltersEditSubmit(w http.ResponseWriter, r *http.Request, user c.User, wfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	// TODO: Either call it isJs or js rather than flip-flopping back and forth across the routes x.x
	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		return c.NoPermissionsJSQ(w, r, user, isJs)
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		return c.LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
	}

	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return c.LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	err = c.WordFilters.Update(id, find, replacement)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}

func WordFiltersDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, wfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		return c.NoPermissionsJSQ(w, r, user, isJs)
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		return c.LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
	}

	err = c.WordFilters.Delete(id)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("This word filter doesn't exist", w, r, user, isJs)
	}

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}
