package common

import (
	"bytes"
	"net/http/httptest"

	"github.com/Azareal/Gosora/common/phrases"
	min "github.com/Azareal/Gosora/common/templates"
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

func wolGetUsers() ([]*User,int) {
	ucount := WsHub.UserCount()
	// We don't want a ridiculously long list, so we'll show the number if it's too high and only show staff individually
	var users []*User
	if ucount < 30 {
		users = WsHub.AllUsers()
		if len(users) >= 30 {
			users = nil
		}
	}
	return users, ucount
}

func wolBuild(widget *Widget, hvars interface{}) (string, error) {
	users, ucount := wolGetUsers()
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

var wolLastUsers []*User

func wolTick(widget *Widget) error {
	w := httptest.NewRecorder()
	users, ucount := wolGetUsers()
	//log.Printf("users: %+v\n",users)
	//log.Printf("wolLastUsers: %+v\n",wolLastUsers)

	// Avoid rebuilding the widget, if the users are exactly the same as on the last tick
	if len(users) == len(wolLastUsers) {
		diff := false
		for i, user := range users {
			if user.ID != wolLastUsers[i].ID {
				diff = true
			}
		}
		if !diff {
			iTickMask := widget.TickMask.Load()
			if iTickMask != nil {
				tickMask := iTickMask.(*Widget)
				if tickMask != nil {
					return nil
				}
			}
		}
	}

	wol := &wolUsers{SimpleDefaultHeader(w), phrases.GetTmplPhrase("widget.online_name"), users, ucount}
	err := wol.Header.Theme.RunTmpl("widget_online", wol, wol.Header.Writer)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(w.Result().Body)
	bs := buf.String()
	if Config.MinifyTemplates {
		bs = min.Minify(bs)
	}

	twidget := &Widget{}
	*twidget = *widget
	twidget.Body = bs
	widget.TickMask.Store(twidget)
	wolLastUsers = users

	hTbl := GetHookTable()
	_, _ = hTbl.VhookSkippable("tasks_tick_widget_wol", widget, bs)

	return nil
}
