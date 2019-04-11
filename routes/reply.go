package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
)

type ReplyStmts struct {
	updateAttachs     *sql.Stmt
	createReplyPaging *sql.Stmt
}

var replyStmts ReplyStmts

// TODO: Move this statement somewhere else
func init() {
	common.DbInits.Add(func(acc *qgen.Accumulator) error {
		replyStmts = ReplyStmts{
			// TODO: Less race-y attachment count updates
			updateAttachs:     acc.Update("replies").Set("attachCount = ?").Where("rid = ?").Prepare(),
			createReplyPaging: acc.Select("replies").Cols("rid").Where("rid >= ? - 1 AND tid = ?").Orderby("rid ASC").Prepare(),
		}
		return acc.FirstError()
	})
}

type JsonReply struct {
	Content string
}

func CreateReplySubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	// TODO: Use this
	js := r.FormValue("js") == "1"
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		return common.PreErrorJSQ("Failed to convert the Topic ID", w, r, js)
	}

	topic, err := common.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("Couldn't find the parent topic", w, r, js)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Add hooks to make use of headerLite
	lite, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateReply {
		return common.NoPermissionsJSQ(w, r, user, js)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return common.NoPermissionsJSQ(w, r, user, js)
	}

	content := common.PreparseMessage(r.PostFormValue("reply-content"))
	// TODO: Fully parse the post and put that in the parsed column
	rid, err := common.Rstore.Create(topic, content, user.LastIP, user.ID)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	reply, err := common.Rstore.Get(rid)
	if err != nil {
		return common.LocalErrorJSQ("Unable to load the reply", w, r, user, js)
	}

	// Handle the file attachments
	// TODO: Stop duplicating this code
	if user.Perms.UploadFiles {
		_, rerr := uploadAttachment(w, r, user, topic.ParentID, "forums", rid, "replies")
		if rerr != nil {
			return rerr
		}
	}

	if r.PostFormValue("has_poll") == "1" {
		var maxPollOptions = 10
		var pollInputItems = make(map[int]string)
		for key, values := range r.Form {
			//common.DebugDetail("key: ", key)
			//common.DebugDetailf("values: %+v\n", values)
			for _, value := range values {
				if strings.HasPrefix(key, "pollinputitem[") {
					halves := strings.Split(key, "[")
					if len(halves) != 2 {
						return common.LocalErrorJSQ("Malformed pollinputitem", w, r, user, js)
					}
					halves[1] = strings.TrimSuffix(halves[1], "]")

					index, err := strconv.Atoi(halves[1])
					if err != nil {
						return common.LocalErrorJSQ("Malformed pollinputitem", w, r, user, js)
					}

					// If there are duplicates, then something has gone horribly wrong, so let's ignore them, this'll likely happen during an attack
					_, exists := pollInputItems[index]
					// TODO: Should we use SanitiseBody instead to keep the newlines?
					if !exists && len(common.SanitiseSingleLine(value)) != 0 {
						pollInputItems[index] = common.SanitiseSingleLine(value)
						if len(pollInputItems) >= maxPollOptions {
							break
						}
					}
				}
			}
		}

		// Make sure the indices are sequential to avoid out of bounds issues
		var seqPollInputItems = make(map[int]string)
		for i := 0; i < len(pollInputItems); i++ {
			seqPollInputItems[i] = pollInputItems[i]
		}

		pollType := 0 // Basic single choice
		_, err := common.Polls.Create(reply, pollType, seqPollInputItems)
		if err != nil {
			return common.LocalErrorJSQ("Failed to add poll to reply", w, r, user, js) // TODO: Might need to be an internal error as it could leave phantom polls?
		}
	}

	err = common.Forums.UpdateLastTopic(tid, user.ID, topic.ParentID)
	if err != nil && err != sql.ErrNoRows {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	common.AddActivityAndNotifyAll(user.ID, topic.CreatedBy, "reply", "topic", tid)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	wcount := common.WordCount(content)
	err = user.IncreasePostStats(wcount, false)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	nTopic, err := common.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("Couldn't find the parent topic", w, r, js)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	page := common.LastPage(nTopic.PostCount, common.Config.ItemsPerPage)

	rows, err := replyStmts.createReplyPaging.Query(reply.ID, topic.ID)
	if err != nil && err != sql.ErrNoRows {
		return common.InternalErrorJSQ(err, w, r, js)
	}
	defer rows.Close()

	var rids []int
	for rows.Next() {
		var rid int
		err := rows.Scan(&rid)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, js)
		}
		rids = append(rids, rid)
	}
	err = rows.Err()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}
	if len(rids) == 0 {
		return common.NotFoundJSQ(w, r, nil, js)
	}

	if page > 1 {
		var offset int
		if rids[0] == reply.ID {
			offset = 1
		} else if len(rids) == 2 && rids[1] == reply.ID {
			offset = 2
		}
		page = common.LastPage(nTopic.PostCount-(len(rids)+offset), common.Config.ItemsPerPage)
	}

	counters.PostCounter.Bump()
	skip, rerr := lite.Hooks.VhookSkippable("action_end_create_reply", reply.ID)
	if skip || rerr != nil {
		return rerr
	}

	prid, _ := strconv.Atoi(r.FormValue("prid"))
	if js && (prid == 0 || rids[0] == prid) {
		outBytes, err := json.Marshal(JsonReply{common.ParseMessage(reply.Content, topic.ParentID, "forums")})
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, js)
		}
		w.Write(outBytes)
	} else {
		var spage string
		if page > 1 {
			spage = "?page=" + strconv.Itoa(page)
		}
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tid)+spage+"#post-"+strconv.Itoa(reply.ID), http.StatusSeeOther)
	}
	return nil
}

// TODO: Disable stat updates in posts handled by plugin_guilds
// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
func ReplyEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	js := (r.PostFormValue("js") == "1")
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, js)
	}

	reply, err := common.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The target reply doesn't exist.", w, r, js)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	topic, err := reply.Topic()
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, js)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Add hooks to make use of headerLite
	lite, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		return common.NoPermissionsJSQ(w, r, user, js)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return common.NoPermissionsJSQ(w, r, user, js)
	}

	err = reply.SetPost(r.PostFormValue("edit_item"))
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, js)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Avoid the load to get this faster?
	reply, err = common.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The updated reply doesn't exist.", w, r, js)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	skip, rerr := lite.Hooks.VhookSkippable("action_end_edit_reply", reply.ID)
	if skip || rerr != nil {
		return rerr
	}

	if !js {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(topic.ID)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		outBytes, err := json.Marshal(JsonReply{common.ParseMessage(reply.Content, topic.ParentID, "forums")})
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, js)
		}
		w.Write(outBytes)
	}

	return nil
}

// TODO: Refactor this
// TODO: Disable stat updates in posts handled by plugin_guilds
func ReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("isJs") == "1")
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	reply, err := common.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The reply you tried to delete doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	topic, err := common.Topics.Get(reply.ParentID)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	lite, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.DeleteReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.Delete()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	skip, rerr := lite.Hooks.VhookSkippable("action_end_delete_reply", reply.ID)
	if skip || rerr != nil {
		return rerr
	}

	//log.Printf("Reply #%d was deleted by common.User #%d", rid, user.ID)
	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}

	// ? - What happens if an error fires after a redirect...?
	replyCreator, err := common.Users.Get(reply.CreatedBy)
	if err == nil {
		wcount := common.WordCount(reply.Content)
		err = replyCreator.DecreasePostStats(wcount, false)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}
	} else if err != sql.ErrNoRows {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	err = common.ModLogs.Create("delete", reply.ParentID, "reply", user.LastIP, user.ID)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return nil
}

// TODO: Avoid uploading this again if the attachment already exists? They'll resolve to the same hash either way, but we could save on some IO / bandwidth here
// TODO: Enforce the max request limit on all of this topic's attachments
// TODO: Test this route
func AddAttachToReplySubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.LocalErrorJS(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
	}

	reply, err := common.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJS("You can't attach to something which doesn't exist!", w, r)
	} else if err != nil {
		return common.InternalErrorJS(err, w, r)
	}

	topic, err := common.Topics.Get(reply.ParentID)
	if err != nil {
		return common.NotFoundJS(w, r)
	}

	lite, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply || !user.Perms.UploadFiles {
		return common.NoPermissionsJS(w, r, user)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return common.NoPermissionsJS(w, r, user)
	}

	// Handle the file attachments
	pathMap, rerr := uploadAttachment(w, r, user, topic.ParentID, "forums", rid, "replies")
	if rerr != nil {
		// TODO: This needs to be a JS error...
		return rerr
	}
	if len(pathMap) == 0 {
		return common.InternalErrorJS(errors.New("no paths for attachment add"), w, r)
	}

	skip, rerr := lite.Hooks.VhookSkippable("action_end_add_attach_to_reply", reply.ID)
	if skip || rerr != nil {
		return rerr
	}

	var elemStr string
	for path, aids := range pathMap {
		elemStr += "\"" + path + "\":\"" + aids + "\","
	}
	if len(elemStr) > 1 {
		elemStr = elemStr[:len(elemStr)-1]
	}

	w.Write([]byte(`{"success":"1","elems":[{` + elemStr + `}]}`))
	return nil
}

// TODO: Reduce the amount of duplication between this and RemoveAttachFromTopicSubmit
func RemoveAttachFromReplySubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.LocalErrorJS(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
	}

	reply, err := common.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJS("You can't attach from something which doesn't exist!", w, r)
	} else if err != nil {
		return common.InternalErrorJS(err, w, r)
	}

	topic, err := common.Topics.Get(reply.ParentID)
	if err != nil {
		return common.NotFoundJS(w, r)
	}

	lite, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		return common.NoPermissionsJS(w, r, user)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return common.NoPermissionsJS(w, r, user)
	}

	for _, said := range strings.Split(r.PostFormValue("aids"), ",") {
		aid, err := strconv.Atoi(said)
		if err != nil {
			return common.LocalErrorJS(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
		}
		rerr := deleteAttachment(w, r, user, aid, true)
		if rerr != nil {
			// TODO: This needs to be a JS error...
			return rerr
		}
	}

	skip, rerr := lite.Hooks.VhookSkippable("action_end_remove_attach_from_reply", reply.ID)
	if skip || rerr != nil {
		return rerr
	}

	w.Write(successJSONBytes)
	return nil
}

// TODO: Move the profile reply routes to their own file?
func ProfileReplyCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	if !user.Perms.ViewTopic || !user.Perms.CreateReply {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		return common.LocalError("Invalid UID", w, r, user)
	}

	profileOwner, err := common.Users.Get(uid)
	if err == sql.ErrNoRows {
		return common.LocalError("The profile you're trying to post on doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	content := common.PreparseMessage(r.PostFormValue("reply-content"))
	// TODO: Fully parse the post and store it in the parsed column
	_, err = common.Prstore.Create(profileOwner.ID, content, user.ID, user.LastIP)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// ! Be careful about leaking per-route permission state with &user
	alert := common.Alert{0, user.ID, profileOwner.ID, "reply", "user", profileOwner.ID, &user}
	err = common.AddActivityAndNotifyTarget(alert)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	counters.PostCounter.Bump()
	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func ProfileReplyEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
	}

	reply, err := common.Prstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The target reply doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	creator, err := common.Users.Get(reply.CreatedBy)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// ? Does the admin understand that this group perm affects this?
	if user.ID != creator.ID && !user.Perms.EditReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.SetBody(r.PostFormValue("edit_item"))
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/user/"+strconv.Itoa(creator.ID)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func ProfileReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
	}

	reply, err := common.Prstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The target reply doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	creator, err := common.Users.Get(reply.CreatedBy)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if user.ID != creator.ID && !user.Perms.DeleteReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.Delete()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	//log.Printf("The profile post '%d' was deleted by common.User #%d", reply.ID, user.ID)

	if !isJs {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(creator.ID), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func ReplyLikeSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	reply, err := common.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("You can't like something which doesn't exist!", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	topic, err := common.Topics.Get(reply.ParentID)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}
	if reply.CreatedBy == user.ID {
		return common.LocalErrorJSQ("You can't like your own replies", w, r, user, isJs)
	}

	_, err = common.Users.Get(reply.CreatedBy)
	if err != nil && err != sql.ErrNoRows {
		return common.LocalErrorJSQ("The target user doesn't exist", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	err = reply.Like(user.ID)
	if err == common.ErrAlreadyLiked {
		return common.LocalErrorJSQ("You've already liked this!", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// ! Be careful about leaking per-route permission state with &user
	alert := common.Alert{0, user.ID, reply.CreatedBy, "like", "post", rid, &user}
	err = common.AddActivityAndNotifyTarget(alert)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}
