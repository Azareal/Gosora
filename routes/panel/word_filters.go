package panel

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
)

func WordFilters(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, u, "word_filters", "word-filters")
	if ferr != nil {
		return ferr
	}
	if !u.Perms.EditSettings {
		return c.NoPermissions(w, r, u)
	}

	// TODO: What if this list gets too long?
	filters, e := c.WordFilters.GetAll()
	if e != nil {
		return c.InternalError(e, w, r)
	}

	pi := c.PanelPage{basePage, tList, filters}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_word_filters", &pi})
}

func WordFiltersCreateSubmit(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.EditSettings {
		return c.NoPermissions(w, r, u)
	}
	js := r.PostFormValue("js") == "1"

	// ? - We're not doing a full sanitise here, as it would be useful if admins were able to put down rules for replacing things with HTML, etc.
	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return c.LocalErrorJSQ("You need to specify what word you want to match", w, r, u, js)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replace := strings.TrimSpace(r.PostFormValue("replace"))

	wfid, e := c.WordFilters.Create(find, replace)
	if e != nil {
		return c.InternalErrorJSQ(e, w, r, js)
	}
	e = c.AdminLogs.Create("create", wfid, "word_filter", u.GetIP(), u.ID)
	if e != nil {
		return c.InternalError(e, w, r)
	}

	return successRedirect("/panel/settings/word-filters/", w, r, js)
}

// TODO: Implement this as a non-JS fallback
func WordFiltersEdit(w http.ResponseWriter, r *http.Request, u *c.User, wfid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, u, "edit_word_filter", "word-filters")
	if ferr != nil {
		return ferr
	}
	if !u.Perms.EditSettings {
		return c.NoPermissions(w, r, u)
	}
	_ = wfid

	pi := c.PanelPage{basePage, tList, nil}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_word_filters_edit", &pi})
}

func WordFiltersEditSubmit(w http.ResponseWriter, r *http.Request, u *c.User, swfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	js := r.PostFormValue("js") == "1"
	if !u.Perms.EditSettings {
		return c.NoPermissionsJSQ(w, r, u, js)
	}

	wfid, err := strconv.Atoi(swfid)
	if err != nil {
		return c.LocalErrorJSQ("The word filter ID must be an integer.", w, r, u, js)
	}
	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return c.LocalErrorJSQ("You need to specify what word you want to match", w, r, u, js)
	}
	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replace := strings.TrimSpace(r.PostFormValue("replace"))

	wf, err := c.WordFilters.Get(wfid)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("This word filter doesn't exist.", w, r, u, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	err = c.WordFilters.Update(wfid, find, replace)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	lBytes, err := json.Marshal(c.WordFilterDiff{wf.Find, wf.Replace, find, replace})
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	err = c.AdminLogs.CreateExtra("edit", wfid, "word_filter", u.GetIP(), u.ID, string(lBytes))
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}

func WordFiltersDeleteSubmit(w http.ResponseWriter, r *http.Request, u *c.User, swfid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	js := r.PostFormValue("js") == "1"
	if !u.Perms.EditSettings {
		return c.NoPermissionsJSQ(w, r, u, js)
	}

	wfid, err := strconv.Atoi(swfid)
	if err != nil {
		return c.LocalErrorJSQ("The word filter ID must be an integer.", w, r, u, js)
	}
	err = c.WordFilters.Delete(wfid)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("This word filter doesn't exist", w, r, u, js)
	}
	err = c.AdminLogs.Create("delete", wfid, "word_filter", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}
