package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	"../common"
)

var successJSONBytes = []byte(`{"success":"1"}`)

// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
// TODO: Disable stat updates in posts handled by plugin_guilds
func EditTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	isJs := (r.PostFormValue("js") == "1")

	tid, err := strconv.Atoi(stid)
	if err != nil {
		return common.PreErrorJSQ("The provided TopicID is not a valid number.", w, r, isJs)
	}

	topic, err := common.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The topic you tried to edit doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = topic.Update(r.PostFormValue("topic_name"), r.PostFormValue("topic_content"))
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	err = common.Forums.UpdateLastTopic(topic.ID, user.ID, topic.ParentID)
	if err != nil && err != sql.ErrNoRows {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}
