package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/common"
)

var successJSONBytes = []byte(`{"success":"1"}`)

func ParseSEOURL(urlBit string) (slug string, id int, err error) {
	halves := strings.Split(urlBit, ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	tid, err := strconv.Atoi(halves[1])
	return halves[0], tid, err
}

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
