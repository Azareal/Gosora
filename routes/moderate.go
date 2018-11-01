package routes

import (
	"net/http"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func IPSearch(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = phrases.GetTitlePhrase("ip_search")

	// TODO: How should we handle the permissions if we extend this into an alt detector of sorts?
	if !user.Perms.ViewIPs {
		return common.NoPermissions(w, r, user)
	}

	// TODO: Reject IP Addresses with illegal characters
	var ip = common.SanitiseSingleLine(r.FormValue("ip"))
	uids, err := common.IPSearch.Lookup(ip)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: What if a user is deleted via the Control Panel? We'll cross that bridge when we come to it, although we might lean towards blanking the account and removing the related data rather than purging it
	userList, err := common.Users.BulkGetMap(uids)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pi := common.IPSearchPage{header, userList, ip}
	if common.RunPreRenderHook("pre_render_ip_search", w, r, &user, &pi) {
		return nil
	}
	err = common.RunThemeTemplate(header.Theme.Name, "ip_search", pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}
