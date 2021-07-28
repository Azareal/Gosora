package panel

import (
	"database/sql"
	"net/http"
	"strconv"

	c "github.com/Azareal/Gosora/common"
)

func Pages(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := buildBasePage(w, r, u, "pages", "pages")
	if ferr != nil {
		return ferr
	}
	if r.FormValue("created") == "1" {
		bp.AddNotice("panel_page_created")
	} else if r.FormValue("deleted") == "1" {
		bp.AddNotice("panel_page_deleted")
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
	pi := c.PanelCustomPagesPage{bp, cPages, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_page_list", "", "panel_pages", &pi})
}

func PagesCreateSubmit(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}

	name := c.SanitiseSingleLine(r.PostFormValue("name"))
	if name == "" {
		return c.LocalError("No name was provided for this page", w, r, u)
	}
	title := c.SanitiseSingleLine(r.PostFormValue("title"))
	if title == "" {
		return c.LocalError("No title was provided for this page", w, r, u)
	}
	body := r.PostFormValue("body")
	if body == "" {
		return c.LocalError("No body was provided for this page", w, r, u)
	}

	page := c.BlankCustomPage()
	page.Name = name
	page.Title = title
	page.Body = body
	pid, err := page.Create()
	if err != nil {
		return c.InternalError(err, w, r)
	}
	err = c.AdminLogs.Create("create", pid, "page", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/pages/?created=1", http.StatusSeeOther)
	return nil
}

func PagesEdit(w http.ResponseWriter, r *http.Request, u *c.User, spid string) c.RouteError {
	bp, ferr := buildBasePage(w, r, u, "pages_edit", "pages")
	if ferr != nil {
		return ferr
	}
	if r.FormValue("updated") == "1" {
		bp.AddNotice("panel_page_updated")
	}

	pid, err := strconv.Atoi(spid)
	if err != nil {
		return c.LocalError("Page ID needs to be an integer", w, r, u)
	}
	page, err := c.Pages.Get(pid)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, bp.Header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.PanelCustomPageEditPage{bp, page}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_page_edit", "", "panel_pages_edit", &pi})
}

func PagesEditSubmit(w http.ResponseWriter, r *http.Request, u *c.User, spid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}

	pid, err := strconv.Atoi(spid)
	if err != nil {
		return c.LocalError("Page ID needs to be an integer", w, r, u)
	}
	name := c.SanitiseSingleLine(r.PostFormValue("name"))
	if name == "" {
		return c.LocalError("No name was provided for this page", w, r, u)
	}
	title := c.SanitiseSingleLine(r.PostFormValue("title"))
	if title == "" {
		return c.LocalError("No title was provided for this page", w, r, u)
	}
	body := r.PostFormValue("body")
	if body == "" {
		return c.LocalError("No body was provided for this page", w, r, u)
	}

	p, err := c.Pages.Get(pid)
	if err != nil {
		return c.NotFound(w, r, nil)
	}
	p.Name = name
	p.Title = title
	p.Body = body
	err = p.Commit()
	if err != nil {
		return c.InternalError(err, w, r)
	}
	err = c.AdminLogs.Create("edit", pid, "page", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/pages/?updated=1", http.StatusSeeOther)
	return nil
}

func PagesDeleteSubmit(w http.ResponseWriter, r *http.Request, u *c.User, spid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}

	pid, err := strconv.Atoi(spid)
	if err != nil {
		return c.LocalError("Page ID needs to be an integer", w, r, u)
	}
	err = c.Pages.Delete(pid)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	err = c.AdminLogs.Create("delete", pid, "page", u.GetIP(), u.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/pages/?deleted=1", http.StatusSeeOther)
	return nil
}
