package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
	p "github.com/Azareal/Gosora/common/phrases"
	qgen "github.com/Azareal/Gosora/query_gen"
)

type ReplyStmts struct {
	updateAttachs     *sql.Stmt
	createReplyPaging *sql.Stmt
}

var replyStmts ReplyStmts

// TODO: Move this statement somewhere else
func init() {
	c.DbInits.Add(func(acc *qgen.Accumulator) error {
		replyStmts = ReplyStmts{
			// TODO: Less race-y attachment count updates
			updateAttachs:     acc.Update("replies").Set("attachCount=?").Where("rid=?").Prepare(),
			createReplyPaging: acc.Select("replies").Cols("rid").Where("rid >= ? - 1 AND tid=?").Orderby("rid ASC").Prepare(),
		}
		return acc.FirstError()
	})
}

type JsonReply struct {
	Content string
}

func CreateReplySubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	// TODO: Use this
	js := r.FormValue("js") == "1"
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		return c.PreErrorJSQ("Failed to convert the Topic ID", w, r, js)
	}

	topic, err := c.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("Couldn't find the parent topic", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Add hooks to make use of headerLite
	lite, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateReply {
		return c.NoPermissionsJSQ(w, r, user, js)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	content := c.PreparseMessage(r.PostFormValue("content"))
	// TODO: Fully parse the post and put that in the parsed column
	rid, err := c.Rstore.Create(topic, content, user.GetIP(), user.ID)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	reply, err := c.Rstore.Get(rid)
	if err != nil {
		return c.LocalErrorJSQ("Unable to load the reply", w, r, user, js)
	}

	// Handle the file attachments
	// TODO: Stop duplicating this code
	if user.Perms.UploadFiles {
		_, rerr := uploadAttachment(w, r, user, topic.ParentID, "forums", rid, "replies", strconv.Itoa(topic.ID))
		if rerr != nil {
			return rerr
		}
	}

	if r.PostFormValue("has_poll") == "1" {
		maxPollOptions := 10
		pollInputItems := make(map[int]string)
		for key, values := range r.Form {
			//c.DebugDetail("key: ", key)
			//c.DebugDetailf("values: %+v\n", values)
			for _, value := range values {
				if !strings.HasPrefix(key, "pollinputitem[") {
					continue
				}
				halves := strings.Split(key, "[")
				if len(halves) != 2 {
					return c.LocalErrorJSQ("Malformed pollinputitem", w, r, user, js)
				}
				halves[1] = strings.TrimSuffix(halves[1], "]")

				index, err := strconv.Atoi(halves[1])
				if err != nil {
					return c.LocalErrorJSQ("Malformed pollinputitem", w, r, user, js)
				}

				// If there are duplicates, then something has gone horribly wrong, so let's ignore them, this'll likely happen during an attack
				_, exists := pollInputItems[index]
				// TODO: Should we use SanitiseBody instead to keep the newlines?
				if !exists && len(c.SanitiseSingleLine(value)) != 0 {
					pollInputItems[index] = c.SanitiseSingleLine(value)
					if len(pollInputItems) >= maxPollOptions {
						break
					}
				}
			}
		}

		// Make sure the indices are sequential to avoid out of bounds issues
		seqPollInputItems := make(map[int]string)
		for i := 0; i < len(pollInputItems); i++ {
			seqPollInputItems[i] = pollInputItems[i]
		}

		pollType := 0 // Basic single choice
		_, err := c.Polls.Create(reply, pollType, seqPollInputItems)
		if err != nil {
			return c.LocalErrorJSQ("Failed to add poll to reply", w, r, user, js) // TODO: Might need to be an internal error as it could leave phantom polls?
		}
	}

	err = c.Forums.UpdateLastTopic(tid, user.ID, topic.ParentID)
	if err != nil && err != sql.ErrNoRows {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	c.AddActivityAndNotifyAll(c.Alert{ActorID: user.ID, TargetUserID: topic.CreatedBy, Event: "reply", ElementType: "topic", ElementID: tid, Extra: strconv.Itoa(rid)})
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	err = user.IncreasePostStats(c.WordCount(content), false)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	nTopic, err := c.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("Couldn't find the parent topic", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	page := c.LastPage(nTopic.PostCount, c.Config.ItemsPerPage)

	rows, err := replyStmts.createReplyPaging.Query(reply.ID, topic.ID)
	if err != nil && err != sql.ErrNoRows {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	defer rows.Close()

	var rids []int
	for rows.Next() {
		var rid int
		if err := rows.Scan(&rid); err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
		}
		rids = append(rids, rid)
	}
	if err := rows.Err(); err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	if len(rids) == 0 {
		return c.NotFoundJSQ(w, r, nil, js)
	}

	if page > 1 {
		var offset int
		if rids[0] == reply.ID {
			offset = 1
		} else if len(rids) == 2 && rids[1] == reply.ID {
			offset = 2
		}
		page = c.LastPage(nTopic.PostCount-(len(rids)+offset), c.Config.ItemsPerPage)
	}

	counters.PostCounter.Bump()
	skip, rerr := lite.Hooks.VhookSkippable("action_end_create_reply", reply.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	prid, _ := strconv.Atoi(r.FormValue("prid"))
	if js && (prid == 0 || rids[0] == prid) {
		outBytes, err := json.Marshal(JsonReply{c.ParseMessage(reply.Content, topic.ParentID, "forums", user.ParseSettings, &user)})
		if err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
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
func ReplyEditSubmit(w http.ResponseWriter, r *http.Request, user c.User, srid string) c.RouteError {
	js := r.PostFormValue("js") == "1"
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return c.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, js)
	}

	reply, err := c.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The target reply doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	topic, err := reply.Topic()
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The parent topic doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Add hooks to make use of headerLite
	lite, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		return c.NoPermissionsJSQ(w, r, user, js)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	err = reply.SetPost(r.PostFormValue("edit_item"))
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The parent topic doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Avoid the load to get this faster?
	reply, err = c.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The updated reply doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	skip, rerr := lite.Hooks.VhookSkippable("action_end_edit_reply", reply.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	if !js {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(topic.ID)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		outBytes, err := json.Marshal(JsonReply{c.ParseMessage(reply.Content, topic.ParentID, "forums", user.ParseSettings, &user)})
		if err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
		}
		w.Write(outBytes)
	}

	return nil
}

// TODO: Refactor this
// TODO: Disable stat updates in posts handled by plugin_guilds
func ReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, srid string) c.RouteError {
	js := r.PostFormValue("js") == "1"
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return c.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, js)
	}

	reply, err := c.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The reply you tried to delete doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	topic, err := c.Topics.Get(reply.ParentID)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The parent topic doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Add hooks to make use of headerLite
	lite, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if reply.CreatedBy != user.ID {
		if !user.Perms.ViewTopic || !user.Perms.DeleteReply {
			return c.NoPermissionsJSQ(w, r, user, js)
		}
	}
	if err := reply.Delete(); err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	skip, rerr := lite.Hooks.VhookSkippable("action_end_delete_reply", reply.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	//log.Printf("Reply #%d was deleted by c.User #%d", rid, user.ID)
	if !js {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}

	// ? - What happens if an error fires after a redirect...?
	/*creator, err := c.Users.Get(reply.CreatedBy)
	if err == nil {
		err = creator.DecreasePostStats(c.WordCount(reply.Content), false)
		if err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
		}
	} else if err != sql.ErrNoRows {
		return c.InternalErrorJSQ(err, w, r, js)
	}*/

	err = c.ModLogs.Create("delete", reply.ParentID, "reply", user.GetIP(), user.ID)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	return nil
}

// TODO: Avoid uploading this again if the attachment already exists? They'll resolve to the same hash either way, but we could save on some IO / bandwidth here
// TODO: Enforce the max request limit on all of this topic's attachments
// TODO: Test this route
func AddAttachToReplySubmit(w http.ResponseWriter, r *http.Request, user c.User, srid string) c.RouteError {
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return c.LocalErrorJS(p.GetErrorPhrase("id_must_be_integer"), w, r)
	}

	reply, err := c.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return c.PreErrorJS("You can't attach to something which doesn't exist!", w, r)
	} else if err != nil {
		return c.InternalErrorJS(err, w, r)
	}

	topic, err := c.Topics.Get(reply.ParentID)
	if err != nil {
		return c.NotFoundJS(w, r)
	}

	lite, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply || !user.Perms.UploadFiles {
		return c.NoPermissionsJS(w, r, user)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return c.NoPermissionsJS(w, r, user)
	}

	// Handle the file attachments
	pathMap, rerr := uploadAttachment(w, r, user, topic.ParentID, "forums", rid, "replies", strconv.Itoa(topic.ID))
	if rerr != nil {
		// TODO: This needs to be a JS error...
		return rerr
	}
	if len(pathMap) == 0 {
		return c.InternalErrorJS(errors.New("no paths for attachment add"), w, r)
	}
	_ = c.Rstore.GetCache().Remove(reply.ID)

	skip, rerr := lite.Hooks.VhookSkippable("action_end_add_attach_to_reply", reply.ID, &user)
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

	w.Write([]byte(`{"success":1,"elems":{` + elemStr + `}}`))
	return nil
}

// TODO: Reduce the amount of duplication between this and RemoveAttachFromTopicSubmit
func RemoveAttachFromReplySubmit(w http.ResponseWriter, r *http.Request, user c.User, srid string) c.RouteError {
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return c.LocalErrorJS(p.GetErrorPhrase("id_must_be_integer"), w, r)
	}

	reply, err := c.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return c.PreErrorJS("You can't attach from something which doesn't exist!", w, r)
	} else if err != nil {
		return c.InternalErrorJS(err, w, r)
	}

	topic, err := c.Topics.Get(reply.ParentID)
	if err != nil {
		return c.NotFoundJS(w, r)
	}

	lite, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		return c.NoPermissionsJS(w, r, user)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return c.NoPermissionsJS(w, r, user)
	}

	saids := strings.Split(r.PostFormValue("aids"), ",")
	if len(saids) == 0 {
		return c.LocalErrorJS("No aids provided", w, r)
	}
	for _, said := range saids {
		aid, err := strconv.Atoi(said)
		if err != nil {
			return c.LocalErrorJS(p.GetErrorPhrase("id_must_be_integer"), w, r)
		}
		rerr := deleteAttachment(w, r, user, aid, true)
		if rerr != nil {
			// TODO: This needs to be a JS error...
			return rerr
		}
	}
	_ = c.Rstore.GetCache().Remove(reply.ID)

	skip, rerr := lite.Hooks.VhookSkippable("action_end_remove_attach_from_reply", reply.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	w.Write(successJSONBytes)
	return nil
}

func ReplyLikeSubmit(w http.ResponseWriter, r *http.Request, user c.User, srid string) c.RouteError {
	js := r.PostFormValue("js") == "1"
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return c.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, js)
	}

	reply, err := c.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("You can't like something which doesn't exist!", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	topic, err := c.Topics.Get(reply.ParentID)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The parent topic doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Add hooks to make use of headerLite
	lite, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		return c.NoPermissionsJSQ(w, r, user, js)
	}
	if reply.CreatedBy == user.ID {
		return c.LocalErrorJSQ("You can't like your own replies", w, r, user, js)
	}

	_, err = c.Users.Get(reply.CreatedBy)
	if err != nil && err != sql.ErrNoRows {
		return c.LocalErrorJSQ("The target user doesn't exist", w, r, user, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	err = reply.Like(user.ID)
	if err == c.ErrAlreadyLiked {
		return c.LocalErrorJSQ("You've already liked this!", w, r, user, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// ! Be careful about leaking per-route permission state with &user
	alert := c.Alert{ActorID: user.ID, TargetUserID: reply.CreatedBy, Event: "like", ElementType: "post", ElementID: rid, Actor: &user}
	err = c.AddActivityAndNotifyTarget(alert)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	skip, rerr := lite.Hooks.VhookSkippable("action_end_like_reply", reply.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	if !js {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}

func ReplyUnlikeSubmit(w http.ResponseWriter, r *http.Request, user c.User, srid string) c.RouteError {
	js := r.PostFormValue("js") == "1"
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return c.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, js)
	}

	reply, err := c.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("You can't unlike something which doesn't exist!", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	topic, err := c.Topics.Get(reply.ParentID)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The parent topic doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Add hooks to make use of headerLite
	lite, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	_, err = c.Users.Get(reply.CreatedBy)
	if err != nil && err != sql.ErrNoRows {
		return c.LocalErrorJSQ("The target user doesn't exist", w, r, user, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	err = reply.Unlike(user.ID)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Better coupling between the two params queries
	aids, err := c.Activity.AidsByParams("like", reply.ID, "post")
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	for _, aid := range aids {
		c.DismissAlert(reply.CreatedBy, aid)
	}
	err = c.Activity.DeleteByParams("like", reply.ID, "post")
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	skip, rerr := lite.Hooks.VhookSkippable("action_end_unlike_reply", reply.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	if !js {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}
