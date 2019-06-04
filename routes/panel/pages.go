package panel

import (
	"database/sql"
	"net/http"
	"strconv"

	c "github.com/Azareal/Gosora/common"
)

func Pages(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "pages", "pages")
	if ferr != nil {
		return ferr
	}

	if r.FormValue("created") == "1" {
		basePage.AddNotice("panel_page_created")
	} else if r.FormValue("deleted") == "1" {
		basePage.AddNotice("panel_page_deleted")
	}

	// TODO: Test the pagination here
	pageCount := c.Pages.Count()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 15
	offset, page, lastPage := c.PageOffset(pageCount, page, perPage)

	cPages, err := c.Pages.GetOffset(offset, perPage)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelCustomPagesPage{basePage, cPages, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_page_list", "", "panel_pages", &pi})
}

func PagesCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pname := r.PostFormValue("name")
	if pname == "" {
		return c.LocalError("No name was provided for this page", w, r, user)
	}
	ptitle := r.PostFormValue("title")
	if ptitle == "" {
		return c.LocalError("No title was provided for this page", w, r, user)
	}
	pbody := r.PostFormValue("body")
	if pbody == "" {
		return c.LocalError("No body was provided for this page", w, r, user)
	}

	page := c.BlankCustomPage()
	page.Name = pname
	page.Title = ptitle
	page.Body = pbody
	_, err := page.Create()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/pages/?created=1", http.StatusSeeOther)
	return nil
}

func PagesEdit(w http.ResponseWriter, r *http.Request, user c.User, spid string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "pages_edit", "pages")
	if ferr != nil {
		return ferr
	}
	if r.FormValue("updated") == "1" {
		basePage.AddNotice("panel_page_updated")
	}

	pid, err := strconv.Atoi(spid)
	if err != nil {
		return c.LocalError("Page ID needs to be an integer", w, r, user)
	}

	page, err := c.Pages.Get(pid)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.PanelCustomPageEditPage{basePage, page}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_page_edit", "", "panel_pages_edit", &pi})
}

func PagesEditSubmit(w http.ResponseWriter, r *http.Request, user c.User, spid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pid, err := strconv.Atoi(spid)
	if err != nil {
		return c.LocalError("Page ID needs to be an integer", w, r, user)
	}

	pname := r.PostFormValue("name")
	if pname == "" {
		return c.LocalError("No name was provided for this page", w, r, user)
	}
	ptitle := r.PostFormValue("title")
	if ptitle == "" {
		return c.LocalError("No title was provided for this page", w, r, user)
	}
	pbody := r.PostFormValue("body")
	if pbody == "" {
		return c.LocalError("No body was provided for this page", w, r, user)
	}

	page, err := c.Pages.Get(pid)
	if err != nil {
		return c.NotFound(w, r, nil)
	}
	page.Name = pname
	page.Title = ptitle
	page.Body = pbody
	err = page.Commit()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/pages/?updated=1", http.StatusSeeOther)
	return nil
}

func PagesDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, spid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pid, err := strconv.Atoi(spid)
	if err != nil {
		return c.LocalError("Page ID needs to be an integer", w, r, user)
	}

	err = c.Pages.Delete(pid)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/pages/?deleted=1", http.StatusSeeOther)
	return nil
}
