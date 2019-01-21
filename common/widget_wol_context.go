package common

import "github.com/Azareal/Gosora/common/phrases"

func wolContextRender(widget *Widget, hvars interface{}) (string, error) {
	ucount := WsHub.UserCount()
	// We don't want a ridiculously long list, so we'll show the number if it's too high and only show staff individually
	var users []*User
	if ucount < 30 {
		users = WsHub.AllUsers()
	}
	wol := &wolUsers{hvars.(*Header), phrases.GetTmplPhrase("widget.online_name"), users, ucount}
	err := wol.Header.Theme.RunTmpl("widget_online", wol, wol.Header.Writer)
	return "", err
}
