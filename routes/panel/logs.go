package panel

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

// TODO: Link the usernames for successful registrations to the profiles
func LogsRegs(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "registration_logs", "logs")
	if ferr != nil {
		return ferr
	}

	logCount := common.RegLogs.GlobalCount()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := common.PageOffset(logCount, page, perPage)

	logs, err := common.RegLogs.GetOffset(offset, perPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	var llist = make([]common.PageRegLogItem, len(logs))
	for index, log := range logs {
		llist[index] = common.PageRegLogItem{log, strings.Replace(strings.TrimSuffix(log.FailureReason, "|"), "|", " | ", -1)}
	}

	pageList := common.Paginate(logCount, perPage, 5)
	pi := common.PanelRegLogsPage{basePage, llist, common.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel_reglogs", w, r, basePage.Header, &pi)
}

// TODO: Log errors when something really screwy is going on?
// TODO: Base the slugs on the localised usernames?
func handleUnknownUser(user *common.User, err error) *common.User {
	if err != nil {
		return &common.User{Name: phrases.GetTmplPhrase("user_unknown"), Link: common.BuildProfileURL("unknown", 0)}
	}
	return user
}
func handleUnknownTopic(topic *common.Topic, err error) *common.Topic {
	if err != nil {
		return &common.Topic{Title: phrases.GetTmplPhrase("topic_unknown"), Link: common.BuildTopicURL("unknown", 0)}
	}
	return topic
}

// TODO: Move the log building logic into /common/ and it's own abstraction
func topicElementTypeAction(action string, elementType string, elementID int, actor *common.User, topic *common.Topic) (out string) {
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
			forum, err := common.Forums.Get(fid)
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

func modlogsElementType(action string, elementType string, elementID int, actor *common.User) (out string) {
	switch elementType {
	case "topic":
		topic := handleUnknownTopic(common.Topics.Get(elementID))
		out = topicElementTypeAction(action, elementType, elementID, actor, topic)
	case "user":
		targetUser := handleUnknownUser(common.Users.Get(elementID))
		out = phrases.GetTmplPhrasef("panel_logs_moderation_action_user_"+action, targetUser.Link, targetUser.Name, actor.Link, actor.Name)
	case "reply":
		if action == "delete" {
			topic := handleUnknownTopic(common.TopicByReplyID(elementID))
			out = phrases.GetTmplPhrasef("panel_logs_moderation_action_reply_delete", topic.Link, topic.Title, actor.Link, actor.Name)
		}
	}

	if out == "" {
		out = phrases.GetTmplPhrasef("panel_logs_moderation_action_unknown", action, elementType, actor.Link, actor.Name)
	}
	return out
}

func LogsMod(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "mod_logs", "logs")
	if ferr != nil {
		return ferr
	}

	logCount := common.ModLogs.GlobalCount()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := common.PageOffset(logCount, page, perPage)

	logs, err := common.ModLogs.GetOffset(offset, perPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	var llist = make([]common.PageLogItem, len(logs))
	for index, log := range logs {
		actor := handleUnknownUser(common.Users.Get(log.ActorID))
		action := modlogsElementType(log.Action, log.ElementType, log.ElementID, actor)
		llist[index] = common.PageLogItem{Action: template.HTML(action), IPAddress: log.IPAddress, DoneAt: log.DoneAt}
	}

	pageList := common.Paginate(logCount, perPage, 5)
	pi := common.PanelLogsPage{basePage, llist, common.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel_modlogs", w, r, basePage.Header, &pi)
}

func LogsAdmin(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "admin_logs", "logs")
	if ferr != nil {
		return ferr
	}

	logCount := common.ModLogs.GlobalCount()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := common.PageOffset(logCount, page, perPage)

	logs, err := common.AdminLogs.GetOffset(offset, perPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	var llist = make([]common.PageLogItem, len(logs))
	for index, log := range logs {
		actor := handleUnknownUser(common.Users.Get(log.ActorID))
		action := modlogsElementType(log.Action, log.ElementType, log.ElementID, actor)
		llist[index] = common.PageLogItem{Action: template.HTML(action), IPAddress: log.IPAddress, DoneAt: log.DoneAt}
	}

	pageList := common.Paginate(logCount, perPage, 5)
	pi := common.PanelLogsPage{basePage, llist, common.Paginator{pageList, page, lastPage}}
	return renderTemplate("panel_adminlogs", w, r, basePage.Header, &pi)
}
