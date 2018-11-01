package panel

import (
	"net/http"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}
var successJSONBytes = []byte(`{"success":"1"}`)

// We're trying to reduce the amount of boilerplate in here, so I added these two functions, they might wind up circulating outside this file in the future
func successRedirect(dest string, w http.ResponseWriter, r *http.Request, isJs bool) common.RouteError {
	if !isJs {
		http.Redirect(w, r, dest, http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func renderTemplate(tmplName string, w http.ResponseWriter, r *http.Request, user common.User, pi interface{}) common.RouteError {
	if common.RunPreRenderHook("pre_render_"+tmplName, w, r, &user, pi) {
		return nil
	}
	// TODO: Prepend this with panel_?
	err := common.Templates.ExecuteTemplate(w, tmplName+".html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func buildBasePage(w http.ResponseWriter, r *http.Request, user *common.User, titlePhrase string, zone string) (*common.BasePanelPage, common.RouteError) {
	header, stats, ferr := common.PanelUserCheck(w, r, user)
	if ferr != nil {
		return nil, ferr
	}
	header.Title = phrases.GetTitlePhrase("panel_" + titlePhrase)

	return &common.BasePanelPage{header, stats, zone, common.ReportForumID}, nil
}
