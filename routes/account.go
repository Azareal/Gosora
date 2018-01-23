package routes

import (
	"net/http"

	"../common"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

func AccountEditCritical(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pi := common.Page{"Edit Password", user, headerVars, tList, nil}
	if common.PreRenderHooks["pre_render_account_own_edit_critical"] != nil {
		if common.RunPreRenderHook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return nil
		}
	}
	err := common.Templates.ExecuteTemplate(w, "account_own_edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}
