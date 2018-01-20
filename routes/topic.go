package routes

import (
	"database/sql"
	"encoding/json"
	"log"
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

// TODO: Add support for soft-deletion and add a permission for hard delete in addition to the usual
// TODO: Disable stat updates in posts handled by plugin_guilds
func DeleteTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	// TODO: Move this to some sort of middleware
	var tids []int
	var isJs = false
	if common.ReqIsJson(r) {
		if r.Body == nil {
			return common.PreErrorJS("No request body", w, r)
		}
		err := json.NewDecoder(r.Body).Decode(&tids)
		if err != nil {
			return common.PreErrorJS("We weren't able to parse your data", w, r)
		}
		isJs = true
	} else {
		tid, err := strconv.Atoi(r.URL.Path[len("/topic/delete/submit/"):])
		if err != nil {
			return common.PreError("The provided TopicID is not a valid number.", w, r)
		}
		tids = append(tids, tid)
	}
	if len(tids) == 0 {
		return common.LocalErrorJSQ("You haven't provided any IDs", w, r, user, isJs)
	}

	for _, tid := range tids {
		topic, err := common.Topics.Get(tid)
		if err == sql.ErrNoRows {
			return common.PreErrorJSQ("The topic you tried to delete doesn't exist.", w, r, isJs)
		} else if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}

		// TODO: Add hooks to make use of headerLite
		_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.DeleteTopic {
			return common.NoPermissionsJSQ(w, r, user, isJs)
		}

		// We might be able to handle this err better
		err = topic.Delete()
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}

		err = common.ModLogs.Create("delete", tid, "topic", user.LastIP, user.ID)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}

		// ? - We might need to add soft-delete before we can do an action reply for this
		/*_, err = stmts.createActionReply.Exec(tid,"delete",ipaddress,user.ID)
		if err != nil {
			return common.InternalErrorJSQ(err,w,r,isJs)
		}*/

		log.Printf("Topic #%d was deleted by common.User #%d", tid, user.ID)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func StickTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return common.PreError("The provided TopicID is not a valid number.", w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return common.PreError("The topic you tried to pin doesn't exist.", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		return common.NoPermissions(w, r, user)
	}

	err = topic.Stick()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("stick", tid, "topic", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	err = topic.CreateActionReply("stick", user.LastIP, user)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func UnstickTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return common.PreError("The provided TopicID is not a valid number.", w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return common.PreError("The topic you tried to unpin doesn't exist.", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		return common.NoPermissions(w, r, user)
	}

	err = topic.Unstick()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("unstick", tid, "topic", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	err = topic.CreateActionReply("unstick", user.LastIP, user)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func LockTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	// TODO: Move this to some sort of middleware
	var tids []int
	var isJs = false
	if common.ReqIsJson(r) {
		if r.Body == nil {
			return common.PreErrorJS("No request body", w, r)
		}
		err := json.NewDecoder(r.Body).Decode(&tids)
		if err != nil {
			return common.PreErrorJS("We weren't able to parse your data", w, r)
		}
		isJs = true
	} else {
		tid, err := strconv.Atoi(r.URL.Path[len("/topic/lock/submit/"):])
		if err != nil {
			return common.PreError("The provided TopicID is not a valid number.", w, r)
		}
		tids = append(tids, tid)
	}
	if len(tids) == 0 {
		return common.LocalErrorJSQ("You haven't provided any IDs", w, r, user, isJs)
	}

	for _, tid := range tids {
		topic, err := common.Topics.Get(tid)
		if err == sql.ErrNoRows {
			return common.PreErrorJSQ("The topic you tried to lock doesn't exist.", w, r, isJs)
		} else if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}

		// TODO: Add hooks to make use of headerLite
		_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.CloseTopic {
			return common.NoPermissionsJSQ(w, r, user, isJs)
		}

		err = topic.Lock()
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}

		err = common.ModLogs.Create("lock", tid, "topic", user.LastIP, user.ID)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}
		err = topic.CreateActionReply("lock", user.LastIP, user)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}
	}

	if len(tids) == 1 {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tids[0]), http.StatusSeeOther)
	}
	return nil
}

func UnlockTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return common.PreError("The provided TopicID is not a valid number.", w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return common.PreError("The topic you tried to unlock doesn't exist.", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CloseTopic {
		return common.NoPermissions(w, r, user)
	}

	err = topic.Unlock()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("unlock", tid, "topic", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	err = topic.CreateActionReply("unlock", user.LastIP, user)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

// ! JS only route
// TODO: Figure a way to get this route to work without JS
func MoveTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.PreErrorJS("The provided Forum ID is not a valid number.", w, r)
	}

	// TODO: Move this to some sort of middleware
	var tids []int
	if r.Body == nil {
		return common.PreErrorJS("No request body", w, r)
	}
	err = json.NewDecoder(r.Body).Decode(&tids)
	if err != nil {
		return common.PreErrorJS("We weren't able to parse your data", w, r)
	}
	if len(tids) == 0 {
		return common.LocalErrorJS("You haven't provided any IDs", w, r)
	}

	for _, tid := range tids {
		topic, err := common.Topics.Get(tid)
		if err == sql.ErrNoRows {
			return common.PreErrorJS("The topic you tried to move doesn't exist.", w, r)
		} else if err != nil {
			return common.InternalErrorJS(err, w, r)
		}

		// TODO: Add hooks to make use of headerLite
		_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.MoveTopic {
			return common.NoPermissionsJS(w, r, user)
		}
		_, ferr = common.SimpleForumUserCheck(w, r, &user, fid)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.MoveTopic {
			return common.NoPermissionsJS(w, r, user)
		}

		err = topic.MoveTo(fid)
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}

		// TODO: Log more data so we can list the destination forum in the action post?
		err = common.ModLogs.Create("move", tid, "topic", user.LastIP, user.ID)
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}
		err = topic.CreateActionReply("move", user.LastIP, user)
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}
	}

	if len(tids) == 1 {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tids[0]), http.StatusSeeOther)
	}
	return nil
}
