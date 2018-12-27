package routes

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
)

type TopicStmts struct {
	getReplies    *sql.Stmt
	getLikedTopic *sql.Stmt
	updateAttachs *sql.Stmt
}

var topicStmts TopicStmts

// TODO: Move these DbInits into a TopicList abstraction
func init() {
	common.DbInits.Add(func(acc *qgen.Accumulator) error {
		topicStmts = TopicStmts{
			getReplies:    acc.SimpleLeftJoin("replies", "users", "replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress, replies.likeCount, replies.actionType", "replies.createdBy = users.uid", "replies.tid = ?", "replies.rid ASC", "?,?"),
			getLikedTopic: acc.Select("likes").Columns("targetItem").Where("sentBy = ? && targetItem = ? && targetType = 'topics'").Prepare(),
			// TODO: Less race-y attachment count updates
			updateAttachs: acc.Update("topics").Set("attachCount = ?").Where("tid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

func ViewTopic(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header, urlBit string) common.RouteError {
	page, _ := strconv.Atoi(r.FormValue("page"))
	_, tid, err := ParseSEOURL(urlBit)
	if err != nil {
		return common.PreError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r)
	}

	// Get the topic...
	topic, err := common.GetTopicUser(&user, tid)
	if err == sql.ErrNoRows {
		return common.NotFound(w, r, nil) // TODO: Can we add a simplified invocation of headerVars here? This is likely to be an extremely common NotFound
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	ferr := common.ForumUserCheck(header, w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic {
		return common.NoPermissions(w, r, user)
	}
	header.Title = topic.Title
	header.Zone = "view_topic"
	header.Path = common.BuildTopicURL(common.NameToSlug(topic.Title), topic.ID)

	// TODO: Cache ContentHTML when possible?
	topic.ContentHTML = common.ParseMessage(topic.Content, topic.ParentID, "forums")
	topic.ContentLines = strings.Count(topic.Content, "\n")

	postGroup, err := common.Groups.Get(topic.Group)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	topic.Tag = postGroup.Tag
	if postGroup.IsMod {
		topic.ClassName = common.Config.StaffCSS
	}

	forum, err := common.Forums.Get(topic.ParentID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var poll common.Poll
	if topic.Poll != 0 {
		pPoll, err := common.Polls.Get(topic.Poll)
		if err != nil {
			log.Print("Couldn't find the attached poll for topic " + strconv.Itoa(topic.ID))
			return common.InternalError(err, w, r)
		}
		poll = pPoll.Copy()
		header.AddSheet("chartist/chartist.min.css")
		header.AddScript("chartist/chartist.min.js")
	}

	if topic.LikeCount > 0 && user.Liked > 0 {
		var disp int // Discard this value
		err = topicStmts.getLikedTopic.QueryRow(user.ID, topic.ID).Scan(&disp)
		if err == nil {
			topic.Liked = true
		} else if err != nil && err != sql.ErrNoRows {
			return common.InternalError(err, w, r)
		}
	}

	if topic.AttachCount > 0 {
		attachs, err := common.Attachments.MiniTopicGet(topic.ID)
		if err != nil {
			// TODO: We might want to be a little permissive here in-case of a desync?
			return common.InternalError(err, w, r)
		}
		topic.Attachments = attachs
	}

	// Calculate the offset
	offset, page, lastPage := common.PageOffset(topic.PostCount, page, common.Config.ItemsPerPage)
	pageList := common.Paginate(topic.PostCount, common.Config.ItemsPerPage, 5)
	tpage := common.TopicPage{header, []common.ReplyUser{}, topic, forum, poll, common.Paginator{pageList, page, lastPage}}

	// Get the replies if we have any...
	if topic.PostCount > 0 {
		var likedMap = make(map[int]int)
		var likedQueryList = []int{user.ID}

		rows, err := topicStmts.getReplies.Query(topic.ID, offset, common.Config.ItemsPerPage)
		if err == sql.ErrNoRows {
			return common.LocalError("Bad Page. Some of the posts may have been deleted or you got here by directly typing in the page number.", w, r, user)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}
		defer rows.Close()

		// TODO: Factor the user fields out and embed a user struct instead
		replyItem := common.ReplyUser{ClassName: ""}
		for rows.Next() {
			err := rows.Scan(&replyItem.ID, &replyItem.Content, &replyItem.CreatedBy, &replyItem.CreatedAt, &replyItem.LastEdit, &replyItem.LastEditBy, &replyItem.Avatar, &replyItem.CreatedByName, &replyItem.Group, &replyItem.URLPrefix, &replyItem.URLName, &replyItem.Level, &replyItem.IPAddress, &replyItem.LikeCount, &replyItem.ActionType)
			if err != nil {
				return common.InternalError(err, w, r)
			}

			replyItem.UserLink = common.BuildProfileURL(common.NameToSlug(replyItem.CreatedByName), replyItem.CreatedBy)
			replyItem.ParentID = topic.ID
			replyItem.ContentHtml = common.ParseMessage(replyItem.Content, topic.ParentID, "forums")
			replyItem.ContentLines = strings.Count(replyItem.Content, "\n")

			postGroup, err = common.Groups.Get(replyItem.Group)
			if err != nil {
				return common.InternalError(err, w, r)
			}

			if postGroup.IsMod {
				replyItem.ClassName = common.Config.StaffCSS
			} else {
				replyItem.ClassName = ""
			}

			// TODO: Make a function for this? Build a more sophisticated noavatar handling system? Do bulk user loads and let the common.UserStore initialise this?
			replyItem.Avatar, replyItem.MicroAvatar = common.BuildAvatar(replyItem.CreatedBy, replyItem.Avatar)
			replyItem.Tag = postGroup.Tag

			// We really shouldn't have inline HTML, we should do something about this...
			if replyItem.ActionType != "" {
				var action string
				switch replyItem.ActionType {
				case "lock":
					action = "lock"
					replyItem.ActionIcon = "&#x1F512;&#xFE0E"
				case "unlock":
					action = "unlock"
					replyItem.ActionIcon = "&#x1F513;&#xFE0E"
				case "stick":
					action = "stick"
					replyItem.ActionIcon = "&#x1F4CC;&#xFE0E"
				case "unstick":
					action = "unstick"
					replyItem.ActionIcon = "&#x1F4CC;&#xFE0E"
				case "move":
					action = "move"
					replyItem.ActionIcon = ""
				}
				if action != "" {
					replyItem.ActionType = phrases.GetTmplPhrasef("topic.action_topic_"+action, replyItem.UserLink, replyItem.CreatedByName)
				} else {
					// TODO: Only fire this off if a corresponding phrase for the ActionType doesn't exist? Or maybe have some sort of action registry?
					replyItem.ActionType = phrases.GetTmplPhrasef("topic.action_topic_default", replyItem.ActionType)
					replyItem.ActionIcon = ""
				}
			}

			if replyItem.LikeCount > 0 && user.Liked > 0 {
				likedMap[replyItem.ID] = len(tpage.ItemList)
				likedQueryList = append(likedQueryList, replyItem.ID)
			}

			header.Hooks.VhookNoRet("topic_reply_row_assign", &tpage, &replyItem)
			// TODO: Use a pointer instead to make it easier to abstract this loop? What impact would this have on escape analysis?
			tpage.ItemList = append(tpage.ItemList, replyItem)
		}
		err = rows.Err()
		if err != nil {
			return common.InternalError(err, w, r)
		}

		// TODO: Add a config setting to disable the liked query for a burst of extra speed
		if user.Liked > 0 && len(likedQueryList) > 1 /*&& user.LastLiked <= time.Now()*/ {
			// TODO: Abstract this
			rows, err := qgen.NewAcc().Select("likes").Columns("targetItem").Where("sentBy = ? AND targetType = 'replies'").In("targetItem", likedQueryList[1:]).Query(user.ID)
			if err != nil && err != sql.ErrNoRows {
				return common.InternalError(err, w, r)
			}
			defer rows.Close()

			for rows.Next() {
				var likeRid int
				err := rows.Scan(&likeRid)
				if err != nil {
					return common.InternalError(err, w, r)
				}
				tpage.ItemList[likedMap[likeRid]].Liked = true
			}
			err = rows.Err()
			if err != nil {
				return common.InternalError(err, w, r)
			}
		}
	}

	rerr := renderTemplate("topic", w, r, header, tpage)
	counters.TopicViewCounter.Bump(topic.ID) // TODO: Move this into the router?
	counters.ForumViewCounter.Bump(topic.ParentID)
	return rerr
}

// TODO: Avoid uploading this again if the attachment already exists? They'll resolve to the same hash either way, but we could save on some IO / bandwidth here
// TODO: Enforce the max request limit on all of this topic's attachments
// TODO: Test this route
func AddAttachToTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return common.LocalErrorJS(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
	}
	topic, err := common.Topics.Get(tid)
	if err != nil {
		return common.NotFoundJS(w, r)
	}

	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic || !user.Perms.UploadFiles {
		return common.NoPermissionsJS(w, r, user)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return common.NoPermissionsJS(w, r, user)
	}

	// Handle the file attachments
	pathMap, rerr := uploadAttachment(w, r, user, topic.ParentID, "forums", tid, "topics")
	if rerr != nil {
		// TODO: This needs to be a JS error...
		return rerr
	}
	if len(pathMap) == 0 {
		return common.InternalErrorJS(errors.New("no paths for attachment add"), w, r)
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

func RemoveAttachFromTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return common.LocalErrorJS(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
	}
	topic, err := common.Topics.Get(tid)
	if err != nil {
		return common.NotFoundJS(w, r)
	}

	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic {
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

	w.Write(successJSONBytes)
	return nil
}

// ? - Should we add a new permission or permission zone (like per-forum permissions) specifically for profile comment creation
// ? - Should we allow banned users to make reports? How should we handle report abuse?
// TODO: Add a permission to stop certain users from using custom avatars
// ? - Log username changes and put restrictions on this?
// TODO: Test this
// TODO: Revamp this route
func CreateTopic(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header, sfid string) common.RouteError {
	var fid int
	var err error
	if sfid != "" {
		fid, err = strconv.Atoi(sfid)
		if err != nil {
			return common.LocalError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
		}
	}
	if fid == 0 {
		fid = common.Config.DefaultForum
	}

	ferr := common.ForumUserCheck(header, w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateTopic {
		return common.NoPermissions(w, r, user)
	}
	// TODO: Add a phrase for this
	header.Title = phrases.GetTitlePhrase("create_topic")
	header.Zone = "create_topic"

	// Lock this to the forum being linked?
	// Should we always put it in strictmode when it's linked from another forum? Well, the user might end up changing their mind on what forum they want to post in and it would be a hassle, if they had to switch pages, even if it is a single click for many (exc. mobile)
	var strictmode bool
	header.Hooks.VhookNoRet("topic_create_pre_loop", w, r, fid, &header, &user, &strictmode)

	// TODO: Re-add support for plugin_guilds
	var forumList []common.Forum
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = common.Forums.GetAllVisibleIDs()
		if err != nil {
			return common.InternalError(err, w, r)
		}
	} else {
		group, err := common.Groups.Get(user.Group)
		if err != nil {
			// TODO: Refactor this
			common.LocalError("Something weird happened behind the scenes", w, r, user)
			log.Printf("Group #%d doesn't exist, but it's set on common.User #%d", user.Group, user.ID)
			return nil
		}
		canSee = group.CanSee
	}

	// TODO: plugin_superadmin needs to be able to override this loop. Skip flag on topic_create_pre_loop?
	for _, ffid := range canSee {
		// TODO: Surely, there's a better way of doing this. I've added it in for now to support plugin_guilds, but we really need to clean this up
		if strictmode && ffid != fid {
			continue
		}

		// Do a bulk forum fetch, just in case it's the SqlForumStore?
		forum := common.Forums.DirtyGet(ffid)
		if forum.Name != "" && forum.Active {
			fcopy := forum.Copy()
			// TODO: Abstract this
			if header.Hooks.HookSkippable("topic_create_frow_assign", &fcopy) {
				continue
			}
			forumList = append(forumList, fcopy)
		}
	}

	ctpage := common.CreateTopicPage{header, forumList, fid}
	return renderTemplate("create_topic", w, r, header, ctpage)
}

func CreateTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	fid, err := strconv.Atoi(r.PostFormValue("topic-board"))
	if err != nil {
		return common.LocalError("The provided ForumID is not a valid number.", w, r, user)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateTopic {
		return common.NoPermissions(w, r, user)
	}

	topicName := common.SanitiseSingleLine(r.PostFormValue("topic-name"))
	content := common.PreparseMessage(r.PostFormValue("topic-content"))
	// TODO: Fully parse the post and store it in the parsed column
	tid, err := common.Topics.Create(fid, topicName, content, user.ID, user.LastIP)
	if err != nil {
		switch err {
		case common.ErrNoRows:
			return common.LocalError("Something went wrong, perhaps the forum got deleted?", w, r, user)
		case common.ErrNoTitle:
			return common.LocalError("This topic doesn't have a title", w, r, user)
		case common.ErrLongTitle:
			return common.LocalError("The length of the title is too long, max: "+strconv.Itoa(common.Config.MaxTopicTitleLength), w, r, user)
		case common.ErrNoBody:
			return common.LocalError("This topic doesn't have a body", w, r, user)
		}
		return common.InternalError(err, w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err != nil {
		return common.LocalError("Unable to load the topic", w, r, user)
	}
	if r.PostFormValue("has_poll") == "1" {
		var maxPollOptions = 10
		var pollInputItems = make(map[int]string)
		for key, values := range r.Form {
			for _, value := range values {
				if strings.HasPrefix(key, "pollinputitem[") {
					halves := strings.Split(key, "[")
					if len(halves) != 2 {
						return common.LocalError("Malformed pollinputitem", w, r, user)
					}
					halves[1] = strings.TrimSuffix(halves[1], "]")

					index, err := strconv.Atoi(halves[1])
					if err != nil {
						return common.LocalError("Malformed pollinputitem", w, r, user)
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
		_, err := common.Polls.Create(topic, pollType, seqPollInputItems)
		if err != nil {
			return common.LocalError("Failed to add poll to topic", w, r, user) // TODO: Might need to be an internal error as it could leave phantom polls?
		}
	}

	err = common.Subscriptions.Add(user.ID, tid, "topic")
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = user.IncreasePostStats(common.WordCount(content), true)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Handle the file attachments
	if user.Perms.UploadFiles {
		_, rerr := uploadAttachment(w, r, user, fid, "forums", tid, "topics")
		if rerr != nil {
			return rerr
		}
	}

	counters.PostCounter.Bump()
	counters.TopicCounter.Bump()
	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func uploadFilesWithHash(w http.ResponseWriter, r *http.Request, user common.User, dir string) (filenames []string, rerr common.RouteError) {
	files, ok := r.MultipartForm.File["upload_files"]
	if !ok {
		return nil, nil
	}
	if len(files) > 5 {
		return nil, common.LocalError("You can't attach more than five files", w, r, user)
	}

	for _, file := range files {
		if file.Filename == "" {
			continue
		}
		//common.DebugLog("file.Filename ", file.Filename)

		extarr := strings.Split(file.Filename, ".")
		if len(extarr) < 2 {
			return nil, common.LocalError("Bad file", w, r, user)
		}
		ext := extarr[len(extarr)-1]

		// TODO: Can we do this without a regex?
		reg, err := regexp.Compile("[^A-Za-z0-9]+")
		if err != nil {
			return nil, common.LocalError("Bad file extension", w, r, user)
		}
		ext = strings.ToLower(reg.ReplaceAllString(ext, ""))
		if !common.AllowedFileExts.Contains(ext) {
			return nil, common.LocalError("You're not allowed to upload files with this extension", w, r, user)
		}

		infile, err := file.Open()
		if err != nil {
			return nil, common.LocalError("Upload failed", w, r, user)
		}
		defer infile.Close()

		hasher := sha256.New()
		_, err = io.Copy(hasher, infile)
		if err != nil {
			return nil, common.LocalError("Upload failed [Hashing Failed]", w, r, user)
		}
		infile.Close()

		checksum := hex.EncodeToString(hasher.Sum(nil))
		filename := checksum + "." + ext
		outfile, err := os.Create(dir + filename)
		if err != nil {
			return nil, common.LocalError("Upload failed [File Creation Failed]", w, r, user)
		}
		defer outfile.Close()

		infile, err = file.Open()
		if err != nil {
			return nil, common.LocalError("Upload failed", w, r, user)
		}
		defer infile.Close()

		_, err = io.Copy(outfile, infile)
		if err != nil {
			return nil, common.LocalError("Upload failed [Copy Failed]", w, r, user)
		}

		filenames = append(filenames, filename)
	}

	return filenames, nil
}

// TODO: Add a table for the files and lock the file row when performing tasks related to the file
func deleteAttachment(w http.ResponseWriter, r *http.Request, user common.User, aid int, js bool) common.RouteError {
	attach, err := common.Attachments.Get(aid)
	if err == sql.ErrNoRows {
		return common.NotFoundJSQ(w, r, nil, js)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	err = common.Attachments.Delete(aid)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}

	count := common.Attachments.CountInPath(attach.Path)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, js)
	}
	if count == 0 {
		err := os.Remove("./attachs/" + attach.Path)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, js)
		}
	}

	return nil
}

// TODO: Stop duplicating this code
// TODO: Use a transaction here
func uploadAttachment(w http.ResponseWriter, r *http.Request, user common.User, sid int, sectionTable string, oid int, originTable string) (pathMap map[string]string, rerr common.RouteError) {
	pathMap = make(map[string]string)
	files, rerr := uploadFilesWithHash(w, r, user, "./attachs/")
	if rerr != nil {
		return nil, rerr
	}

	for _, filename := range files {
		aid, err := common.Attachments.Add(sid, sectionTable, oid, originTable, user.ID, filename)
		if err != nil {
			return nil, common.InternalError(err, w, r)
		}

		_, ok := pathMap[filename]
		if ok {
			pathMap[filename] += "," + strconv.Itoa(aid)
		} else {
			pathMap[filename] = strconv.Itoa(aid)
		}

		switch sectionTable {
		case "topics":
			_, err = topicStmts.updateAttachs.Exec(common.Attachments.CountInTopic(oid), oid)
			if err != nil {
				return nil, common.InternalError(err, w, r)
			}
			err = common.Topics.Reload(oid)
			if err != nil {
				return nil, common.InternalError(err, w, r)
			}
		}
	}

	return pathMap, nil
}

// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
// TODO: Disable stat updates in posts handled by plugin_guilds
func EditTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	isJs := (r.PostFormValue("js") == "1")
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return common.PreErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, isJs)
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
	if topic.IsClosed && !user.Perms.CloseTopic {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = topic.Update(r.PostFormValue("topic_name"), r.PostFormValue("topic_content"))
	// TODO: Avoid duplicating this across this route and the topic creation route
	if err != nil {
		switch err {
		case common.ErrNoTitle:
			return common.LocalErrorJSQ("This topic doesn't have a title", w, r, user, isJs)
		case common.ErrLongTitle:
			return common.LocalErrorJSQ("The length of the title is too long, max: "+strconv.Itoa(common.Config.MaxTopicTitleLength), w, r, user, isJs)
		case common.ErrNoBody:
			return common.LocalErrorJSQ("This topic doesn't have a body", w, r, user, isJs)
		}
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

		log.Printf("Topic #%d was deleted by UserID #%d", tid, user.ID)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func StickTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	topic, rerr := topicActionPre(stid, "pin", w, r, user)
	if rerr != nil {
		return rerr
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		return common.NoPermissions(w, r, user)
	}
	return topicActionPost(topic.Stick(), "stick", w, r, topic, user)
}

func topicActionPre(stid string, action string, w http.ResponseWriter, r *http.Request, user common.User) (*common.Topic, common.RouteError) {
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return nil, common.PreError(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return nil, common.PreError("The topic you tried to "+action+" doesn't exist.", w, r)
	} else if err != nil {
		return nil, common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return nil, ferr
	}

	return topic, nil
}

func topicActionPost(err error, action string, w http.ResponseWriter, r *http.Request, topic *common.Topic, user common.User) common.RouteError {
	if err != nil {
		return common.InternalError(err, w, r)
	}
	err = addTopicAction(action, topic, user)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(topic.ID), http.StatusSeeOther)
	return nil
}

func UnstickTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	topic, rerr := topicActionPre(stid, "unpin", w, r, user)
	if rerr != nil {
		return rerr
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		return common.NoPermissions(w, r, user)
	}
	return topicActionPost(topic.Unstick(), "unstick", w, r, topic, user)
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

		err = addTopicAction("lock", topic, user)
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
	topic, rerr := topicActionPre(stid, "unlock", w, r, user)
	if rerr != nil {
		return rerr
	}
	if !user.Perms.ViewTopic || !user.Perms.CloseTopic {
		return common.NoPermissions(w, r, user)
	}
	return topicActionPost(topic.Unlock(), "unlock", w, r, topic, user)
}

// ! JS only route
// TODO: Figure a way to get this route to work without JS
func MoveTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.PreErrorJS(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
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
		err = addTopicAction("move", topic, user)
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}
	}

	if len(tids) == 1 {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tids[0]), http.StatusSeeOther)
	}
	return nil
}

func addTopicAction(action string, topic *common.Topic, user common.User) error {
	err := common.ModLogs.Create(action, topic.ID, "topic", user.LastIP, user.ID)
	if err != nil {
		return err
	}
	return topic.CreateActionReply(action, user.LastIP, user.ID)
}

// TODO: Refactor this
func LikeTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	isJs := (r.PostFormValue("isJs") == "1")
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return common.PreErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, isJs)
	}

	topic, err := common.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The requested topic doesn't exist.", w, r, isJs)
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
	if topic.CreatedBy == user.ID {
		return common.LocalErrorJSQ("You can't like your own topics", w, r, user, isJs)
	}

	_, err = common.Users.Get(topic.CreatedBy)
	if err != nil && err == sql.ErrNoRows {
		return common.LocalErrorJSQ("The target user doesn't exist", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	score := 1
	err = topic.Like(score, user.ID)
	if err == common.ErrAlreadyLiked {
		return common.LocalErrorJSQ("You already liked this", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// ! Be careful about leaking per-route permission state with &user
	alert := common.Alert{0, user.ID, topic.CreatedBy, "like", "topic", tid, &user}
	err = common.AddActivityAndNotifyTarget(alert)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}
