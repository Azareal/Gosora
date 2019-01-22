package common

import "github.com/Azareal/Gosora/common/phrases"

func wolContextRender(widget *Widget, hvars interface{}) (string, error) {
	header := hvars.(*Header)
	if header.Zone != "view_topic" {
		return "", nil
	}
	var ucount int
	var users []*User
	topicMutex.RLock()
	topic, ok := topicWatchers[header.ZoneID]
	if ok {
		ucount = len(topic)
		if ucount < 30 {
			for wsUser, _ := range topic {
				users = append(users, wsUser.User)
			}
		}
	}
	topicMutex.RUnlock()
	wol := &wolUsers{header, phrases.GetTmplPhrase("widget.online_view_topic_name"), users, ucount}
	err := header.Theme.RunTmpl("widget_online", wol, header.Writer)
	return "", err
}
