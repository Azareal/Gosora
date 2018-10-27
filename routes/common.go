package routes

import (
	"net/http"

	"github.com/Azareal/Gosora/common"
)

var successJSONBytes = []byte(`{"success":"1"}`)

func renderTemplate(tmplName string, w http.ResponseWriter, r *http.Request, header *common.Header, pi interface{}) common.RouteError {
	if common.RunPreRenderHook("pre_render_"+tmplName, w, r, &header.CurrentUser, pi) {
		return nil
	}
	err := common.RunThemeTemplate(header.Theme.Name, tmplName, pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}
