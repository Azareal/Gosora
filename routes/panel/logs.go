package panel

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"../../common"
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
	return renderTemplate("panel_reglogs", w, r, user, &pi)
}

// TODO: Log errors when something really screwy is going on?
func handleUnknownUser(user *common.User, err error) *common.User {
	if err != nil {
		return &common.User{Name: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
	}
	return user
}
func handleUnknownTopic(topic *common.Topic, err error) *common.Topic {
	if err != nil {
		return &common.Topic{Title: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
	}
	return topic
}

// TODO: Move the log building logic into /common/ and it's own abstraction
func topicElementTypeAction(action string, elementType string, elementID int, actor *common.User, topic *common.Topic) (out string) {
	if action == "delete" {
		return fmt.Sprintf("Topic #%d was deleted by <a href='%s'>%s</a>", elementID, actor.Link, actor.Name)
	}
	switch action {
	case "lock":
		out = "<a href='%s'>%s</a> was locked by <a href='%s'>%s</a>"
	case "unlock":
		out = "<a href='%s'>%s</a> was reopened by <a href='%s'>%s</a>"
	case "stick":
		out = "<a href='%s'>%s</a> was pinned by <a href='%s'>%s</a>"
	case "unstick":
		out = "<a href='%s'>%s</a> was unpinned by <a href='%s'>%s</a>"
	case "move":
		out = "<a href='%s'>%s</a> was moved by <a href='%s'>%s</a>" // TODO: Add where it was moved to, we'll have to change the source data for that, most likely? Investigate that and try to work this in
	default:
		return fmt.Sprintf("Unknown action '%s' on elementType '%s' by <a href='%s'>%s</a>", action, elementType, actor.Link, actor.Name)
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
		switch action {
		case "ban":
			out = "<a href='%s'>%s</a> was banned by <a href='%s'>%s</a>"
		case "unban":
			out = "<a href='%s'>%s</a> was unbanned by <a href='%s'>%s</a>"
		case "activate":
			out = "<a href='%s'>%s</a> was activated by <a href='%s'>%s</a>"
		}
		out = fmt.Sprintf(out, targetUser.Link, targetUser.Name, actor.Link, actor.Name)
	case "reply":
		if action == "delete" {
			topic := handleUnknownTopic(common.TopicByReplyID(elementID))
			out = fmt.Sprintf("A reply in <a href='%s'>%s</a> was deleted by <a href='%s'>%s</a>", topic.Link, topic.Title, actor.Link, actor.Name)
		}
	}

	if out == "" {
		out = fmt.Sprintf("Unknown action '%s' on elementType '%s' by <a href='%s'>%s</a>", action, elementType, actor.Link, actor.Name)
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
	return renderTemplate("panel_modlogs", w, r, user, &pi)
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
	return renderTemplate("panel_adminlogs", w, r, user, &pi)
}
