package panel

import (
	"net/http"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}
var successJSONBytes = []byte(`{"success":"1"}`)

// We're trying to reduce the amount of boilerplate in here, so I added these two functions, they might wind up circulating outside this file in the future
func successRedirect(dest string, w http.ResponseWriter, r *http.Request, isJs bool) c.RouteError {
	if !isJs {
		http.Redirect(w, r, dest, http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

// TODO: Prerender needs to handle dyntmpl templates better...
func renderTemplate(tmplName string, w http.ResponseWriter, r *http.Request, header *c.Header, pi interface{}) c.RouteError {
	header.AddScript("global.js")
	if c.RunPreRenderHook("pre_render_"+tmplName, w, r, &header.CurrentUser, pi) {
		return nil
	}
	// TODO: Prepend this with panel_?
	err := header.Theme.RunTmpl(tmplName, pi, w)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	return nil
}

func buildBasePage(w http.ResponseWriter, r *http.Request, user *c.User, titlePhrase string, zone string) (*c.BasePanelPage, c.RouteError) {
	header, stats, ferr := c.PanelUserCheck(w, r, user)
	if ferr != nil {
		return nil, ferr
	}
	header.Title = phrases.GetTitlePhrase("panel_" + titlePhrase)

	return &c.BasePanelPage{header, stats, zone, c.ReportForumID}, nil
}
