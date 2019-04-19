package routes

import (
	"net/http"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func IPSearch(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header) c.RouteError {
	header.Title = phrases.GetTitlePhrase("ip_search")
	// TODO: How should we handle the permissions if we extend this into an alt detector of sorts?
	if !user.Perms.ViewIPs {
		return c.NoPermissions(w, r, user)
	}

	// TODO: Reject IP Addresses with illegal characters
	var ip = c.SanitiseSingleLine(r.FormValue("ip"))
	uids, err := c.IPSearch.Lookup(ip)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: What if a user is deleted via the Control Panel? We'll cross that bridge when we come to it, although we might lean towards blanking the account and removing the related data rather than purging it
	userList, err := c.Users.BulkGetMap(uids)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.IPSearchPage{header, userList, ip}
	if c.RunPreRenderHook("pre_render_ip_search", w, r, &user, &pi) {
		return nil
	}
	err = header.Theme.RunTmpl("ip_search", pi, w)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	return nil
}
