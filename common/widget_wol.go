package common

import (
	"bytes"
	"net/http/httptest"

	"github.com/Azareal/Gosora/common/phrases"
)

type wolUsers struct {
	*Header
	Name      string
	Users     []*User
	UserCount int
}

func wolInit(widget *Widget, schedule *WidgetScheduler) error {
	schedule.Add(widget)
	return nil
}

func wolBuild(widget *Widget, hvars interface{}) (string, error) {
	ucount := WsHub.UserCount()
	// We don't want a ridiculously long list, so we'll show the number if it's too high and only show staff individually
	var users []*User
	if ucount < 30 {
		users = WsHub.AllUsers()
		if len(users) >= 30 {
			users = nil
		}
	}
	wol := &wolUsers{hvars.(*Header), phrases.GetTmplPhrase("widget.online_name"), users, ucount}
	err := wol.Header.Theme.RunTmpl("widget_online", wol, wol.Header.Writer)
	return "", err
}

func wolRender(widget *Widget, hvars interface{}) (string, error) {
	iTickMask := widget.TickMask.Load()
	if iTickMask != nil {
		tickMask := iTickMask.(*Widget)
		if tickMask != nil {
			return tickMask.Body, nil
		}
	}
	return wolBuild(widget, hvars)
}

func wolTick(widget *Widget) error {
	w := httptest.NewRecorder()
	_, err := wolBuild(widget, SimpleDefaultHeader(w))
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(w.Result().Body)
	widget.TickMask.Store(buf.String())
	return nil
}
