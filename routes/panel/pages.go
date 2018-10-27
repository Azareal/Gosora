package panel

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Azareal/Gosora/common"
)

func Pages(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "pages", "pages")
	if ferr != nil {
		return ferr
	}

	if r.FormValue("created") == "1" {
		basePage.AddNotice("panel_page_created")
	} else if r.FormValue("deleted") == "1" {
		basePage.AddNotice("panel_page_deleted")
	}

	pageCount := common.Pages.GlobalCount()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := common.PageOffset(pageCount, page, perPage)

	cPages, err := common.Pages.GetOffset(offset, perPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pageList := common.Paginate(pageCount, perPage, 5)
	pi := common.PanelCustomPagesPage{basePage, cPages, common.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel_pages", w, r, user, &pi)
}

func PagesCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pname := r.PostFormValue("name")
	if pname == "" {
		return common.LocalError("No name was provided for this page", w, r, user)
	}
	ptitle := r.PostFormValue("title")
	if ptitle == "" {
		return common.LocalError("No title was provided for this page", w, r, user)
	}
	pbody := r.PostFormValue("body")
	if pbody == "" {
		return common.LocalError("No body was provided for this page", w, r, user)
	}

	page := common.BlankCustomPage()
	page.Name = pname
	page.Title = ptitle
	page.Body = pbody
	_, err := page.Create()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/pages/?created=1", http.StatusSeeOther)
	return nil
}

func PagesEdit(w http.ResponseWriter, r *http.Request, user common.User, spid string) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "pages_edit", "pages")
	if ferr != nil {
		return ferr
	}
	if r.FormValue("updated") == "1" {
		basePage.AddNotice("panel_page_updated")
	}

	pid, err := strconv.Atoi(spid)
	if err != nil {
		return common.LocalError("Page ID needs to be an integer", w, r, user)
	}

	page, err := common.Pages.Get(pid)
	if err == sql.ErrNoRows {
		return common.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	pi := common.PanelCustomPageEditPage{basePage, page}
	return renderTemplate("panel_pages_edit", w, r, user, &pi)
}

func PagesEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, spid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pid, err := strconv.Atoi(spid)
	if err != nil {
		return common.LocalError("Page ID needs to be an integer", w, r, user)
	}

	pname := r.PostFormValue("name")
	if pname == "" {
		return common.LocalError("No name was provided for this page", w, r, user)
	}
	ptitle := r.PostFormValue("title")
	if ptitle == "" {
		return common.LocalError("No title was provided for this page", w, r, user)
	}
	pbody := r.PostFormValue("body")
	if pbody == "" {
		return common.LocalError("No body was provided for this page", w, r, user)
	}

	page, err := common.Pages.Get(pid)
	if err != nil {
		return common.NotFound(w, r, nil)
	}
	page.Name = pname
	page.Title = ptitle
	page.Body = pbody
	err = page.Commit()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/pages/?updated=1", http.StatusSeeOther)
	return nil
}

func PagesDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, spid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pid, err := strconv.Atoi(spid)
	if err != nil {
		return common.LocalError("Page ID needs to be an integer", w, r, user)
	}

	err = common.Pages.Delete(pid)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/pages/?deleted=1", http.StatusSeeOther)
	return nil
}
