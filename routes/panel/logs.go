package panel

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

// TODO: Link the usernames for successful registrations to the profiles
func LogsRegs(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "registration_logs", "logs")
	if ferr != nil {
		return ferr
	}

	logCount := c.RegLogs.Count()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 12
	offset, page, lastPage := c.PageOffset(logCount, page, perPage)

	logs, err := c.RegLogs.GetOffset(offset, perPage)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	var llist = make([]c.PageRegLogItem, len(logs))
	for index, log := range logs {
		llist[index] = c.PageRegLogItem{log, strings.Replace(strings.TrimSuffix(log.FailureReason, "|"), "|", " | ", -1)}
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelRegLogsPage{basePage, llist, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage,"","","panel_reglogs", pi})
}

// TODO: Log errors when something really screwy is going on?
// TODO: Base the slugs on the localised usernames?
func handleUnknownUser(user *c.User, err error) *c.User {
	if err != nil {
		return &c.User{Name: phrases.GetTmplPhrase("user_unknown"), Link: c.BuildProfileURL("unknown", 0)}
	}
	return user
}
func handleUnknownTopic(topic *c.Topic, err error) *c.Topic {
	if err != nil {
		return &c.Topic{Title: phrases.GetTmplPhrase("topic_unknown"), Link: c.BuildTopicURL("unknown", 0)}
	}
	return topic
}

// TODO: Move the log building logic into /common/ and it's own abstraction
func topicElementTypeAction(action string, elementType string, elementID int, actor *c.User, topic *c.Topic) (out string) {
	if action == "delete" {
		return phrases.GetTmplPhrasef("panel_logs_moderation_action_topic_delete", elementID, actor.Link, actor.Name)
	}
	var tbit string
	aarr := strings.Split(action, "-")
	switch aarr[0] {
	case "lock","unlock","stick","unstick":
		tbit = aarr[0]
	case "move":
		if len(aarr) == 2 {
			fid, _ := strconv.Atoi(aarr[1])
			forum, err := c.Forums.Get(fid)
			if err == nil {
				return phrases.GetTmplPhrasef("panel_logs_moderation_action_topic_move_dest", topic.Link, topic.Title, forum.Link, forum.Name, actor.Link, actor.Name)
			}
		}
		tbit = "move"
	default:
		return phrases.GetTmplPhrasef("panel_logs_moderation_action_topic_unknown", action, elementType, actor.Link, actor.Name)
	}
	if tbit != "" {
		return phrases.GetTmplPhrasef("panel_logs_moderation_action_topic_"+tbit, topic.Link, topic.Title, actor.Link, actor.Name)
	}
	return fmt.Sprintf(out, topic.Link, topic.Title, actor.Link, actor.Name)
}

func modlogsElementType(action string, elementType string, elementID int, actor *c.User) (out string) {
	switch elementType {
	case "topic":
		topic := handleUnknownTopic(c.Topics.Get(elementID))
		out = topicElementTypeAction(action, elementType, elementID, actor, topic)
	case "user":
		targetUser := handleUnknownUser(c.Users.Get(elementID))
		out = phrases.GetTmplPhrasef("panel_logs_moderation_action_user_"+action, targetUser.Link, targetUser.Name, actor.Link, actor.Name)
	case "reply":
		if action == "delete" {
			topic := handleUnknownTopic(c.TopicByReplyID(elementID))
			out = phrases.GetTmplPhrasef("panel_logs_moderation_action_reply_delete", topic.Link, topic.Title, actor.Link, actor.Name)
		}
	}

	if out == "" {
		out = phrases.GetTmplPhrasef("panel_logs_moderation_action_unknown", action, elementType, actor.Link, actor.Name)
	}
	return out
}

func LogsMod(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "mod_logs", "logs")
	if ferr != nil {
		return ferr
	}

	logCount := c.ModLogs.Count()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 12
	offset, page, lastPage := c.PageOffset(logCount, page, perPage)

	logs, err := c.ModLogs.GetOffset(offset, perPage)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	var llist = make([]c.PageLogItem, len(logs))
	for index, log := range logs {
		actor := handleUnknownUser(c.Users.Get(log.ActorID))
		action := modlogsElementType(log.Action, log.ElementType, log.ElementID, actor)
		llist[index] = c.PageLogItem{Action: template.HTML(action), IPAddress: log.IPAddress, DoneAt: log.DoneAt}
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelLogsPage{basePage, llist, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage,"","","panel_modlogs", pi})
}

func LogsAdmin(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "admin_logs", "logs")
	if ferr != nil {
		return ferr
	}

	logCount := c.ModLogs.Count()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 12
	offset, page, lastPage := c.PageOffset(logCount, page, perPage)

	logs, err := c.AdminLogs.GetOffset(offset, perPage)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	var llist = make([]c.PageLogItem, len(logs))
	for index, log := range logs {
		actor := handleUnknownUser(c.Users.Get(log.ActorID))
		action := modlogsElementType(log.Action, log.ElementType, log.ElementID, actor)
		llist[index] = c.PageLogItem{Action: template.HTML(action), IPAddress: log.IPAddress, DoneAt: log.DoneAt}
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.PanelLogsPage{basePage, llist, c.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage,"","","panel_adminlogs", pi})
}
