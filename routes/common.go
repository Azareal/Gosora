package routes

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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
	if header.MetaDesc != "" && header.OGDesc == "" {
		header.OGDesc = header.MetaDesc
	}
	// TODO: Expand this to non-HTTPS requests too
	if !header.LooseCSP && common.Site.EnableSsl {
		w.Header().Set("Content-Security-Policy", "default-src https: 'unsafe-eval'; style-src https: 'unsafe-eval' 'unsafe-inline'; img-src https: 'unsafe-eval' 'unsafe-inline'; connect-src * 'unsafe-eval' 'unsafe-inline'; upgrade-insecure-requests")
	}
	if header.CurrentUser.IsAdmin {
		header.Elapsed1 = time.Since(header.StartedAt).String()
	}
	if common.RunPreRenderHook("pre_render_"+tmplName, w, r, &header.CurrentUser, pi) {
		return nil
	}
	err := header.Theme.RunTmpl(tmplName, pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}
