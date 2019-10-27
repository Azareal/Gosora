package routes

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"

	//"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
	"github.com/Azareal/Gosora/common/phrases"
	qgen "github.com/Azareal/Gosora/query_gen"
)

type TopicStmts struct {
	getLikedTopic *sql.Stmt
	updateAttachs *sql.Stmt
}

var topicStmts TopicStmts

// TODO: Move these DbInits into a TopicList abstraction
func init() {
	c.DbInits.Add(func(acc *qgen.Accumulator) error {
		topicStmts = TopicStmts{
			getLikedTopic: acc.Select("likes").Columns("targetItem").Where("sentBy = ? && targetItem = ? && targetType = 'topics'").Prepare(),
			// TODO: Less race-y attachment count updates
			updateAttachs: acc.Update("topics").Set("attachCount = ?").Where("tid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

func ViewTopic(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header, urlBit string) c.RouteError {
	page, _ := strconv.Atoi(r.FormValue("page"))
	_, tid, err := ParseSEOURL(urlBit)
	if err != nil {
		return c.SimpleError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, header)
	}

	// Get the topic...
	topic, err := c.GetTopicUser(&user, tid)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, nil) // TODO: Can we add a simplified invocation of header here? This is likely to be an extremely common NotFound
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	ferr := c.ForumUserCheck(header, w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic {
		return c.NoPermissions(w, r, user)
	}
	header.Title = topic.Title
	header.Path = c.BuildTopicURL(c.NameToSlug(topic.Title), topic.ID)

	// TODO: Cache ContentHTML when possible?
	topic.ContentHTML = c.ParseMessage(topic.Content, topic.ParentID, "forums")
	// TODO: Do this more efficiently by avoiding the allocations entirely in ParseMessage, if there's nothing to do.
	if topic.ContentHTML == topic.Content {
		topic.ContentHTML = topic.Content
	}
	topic.ContentLines = strings.Count(topic.Content, "\n")

	header.OGDesc = topic.Content
	if len(header.OGDesc) > 200 {
		header.OGDesc = header.OGDesc[:197] + "..."
	}

	postGroup, err := c.Groups.Get(topic.Group)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	topic.Tag = postGroup.Tag
	if postGroup.IsMod {
		topic.ClassName = c.Config.StaffCSS
	}
	topic.Deletable = user.Perms.DeleteTopic || topic.CreatedBy == user.ID

	forum, err := c.Forums.Get(topic.ParentID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var poll c.Poll
	if topic.Poll != 0 {
		pPoll, err := c.Polls.Get(topic.Poll)
		if err != nil {
			log.Print("Couldn't find the attached poll for topic " + strconv.Itoa(topic.ID))
			return c.InternalError(err, w, r)
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
			return c.InternalError(err, w, r)
		}
	}

	if topic.AttachCount > 0 {
		attachs, err := c.Attachments.MiniGetList("topics", topic.ID)
		if err != nil {
			// TODO: We might want to be a little permissive here in-case of a desync?
			return c.InternalError(err, w, r)
		}
		topic.Attachments = attachs
	}

	// Calculate the offset
	offset, page, lastPage := c.PageOffset(topic.PostCount, page, c.Config.ItemsPerPage)
	pageList := c.Paginate(page, lastPage, 5)
	tpage := c.TopicPage{header, nil, topic, forum, poll, c.Paginator{pageList, page, lastPage}}

	// Get the replies if we have any...
	if topic.PostCount > 0 {
		var pFrag int
		if strings.HasPrefix(r.URL.Fragment, "post-") {
			pFrag, _ = strconv.Atoi(strings.TrimPrefix(r.URL.Fragment, "post-"))
		}

		rlist, ogdesc, err := topic.Replies(offset, pFrag, &user)
		if err == sql.ErrNoRows {
			return c.LocalError("Bad Page. Some of the posts may have been deleted or you got here by directly typing in the page number.", w, r, user)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
		header.OGDesc = ogdesc
		tpage.ItemList = rlist
	}

	header.Zone = "view_topic"
	header.ZoneID = topic.ID
	header.ZoneData = topic

	var rerr c.RouteError
	tmpl := forum.Tmpl
	if tmpl == "" {
		rerr = renderTemplate("topic", w, r, header, tpage)
	} else {
		tmpl = "topic_" + tmpl
		err = renderTemplate3(tmpl, tmpl, w, r, header, tpage)
		if err != nil {
			rerr = renderTemplate("topic", w, r, header, tpage)
		}
	}
	counters.TopicViewCounter.Bump(topic.ID) // TODO: Move this into the router?
	counters.ForumViewCounter.Bump(topic.ParentID)
	return rerr
}

// TODO: Avoid uploading this again if the attachment already exists? They'll resolve to the same hash either way, but we could save on some IO / bandwidth here
// TODO: Enforce the max request limit on all of this topic's attachments
// TODO: Test this route
func AddAttachToTopicSubmit(w http.ResponseWriter, r *http.Request, user c.User, stid string) c.RouteError {
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return c.LocalErrorJS(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
	}
	topic, err := c.Topics.Get(tid)
	if err != nil {
		return c.NotFoundJS(w, r)
	}

	_, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic || !user.Perms.UploadFiles {
		return c.NoPermissionsJS(w, r, user)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return c.NoPermissionsJS(w, r, user)
	}

	// Handle the file attachments
	pathMap, rerr := uploadAttachment(w, r, user, topic.ParentID, "forums", tid, "topics", "")
	if rerr != nil {
		// TODO: This needs to be a JS error...
		return rerr
	}
	if len(pathMap) == 0 {
		return c.InternalErrorJS(errors.New("no paths for attachment add"), w, r)
	}

	var elemStr string
	for path, aids := range pathMap {
		elemStr += "\"" + path + "\":\"" + aids + "\","
	}
	if len(elemStr) > 1 {
		elemStr = elemStr[:len(elemStr)-1]
	}

	w.Write([]byte(`{"success":1,"elems":[{` + elemStr + `}]}`))
	return nil
}

func RemoveAttachFromTopicSubmit(w http.ResponseWriter, r *http.Request, user c.User, stid string) c.RouteError {
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return c.LocalErrorJS(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
	}
	topic, err := c.Topics.Get(tid)
	if err != nil {
		return c.NotFoundJS(w, r)
	}

	_, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic {
		return c.NoPermissionsJS(w, r, user)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return c.NoPermissionsJS(w, r, user)
	}

	for _, said := range strings.Split(r.PostFormValue("aids"), ",") {
		aid, err := strconv.Atoi(said)
		if err != nil {
			return c.LocalErrorJS(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
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
func CreateTopic(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header, sfid string) c.RouteError {
	var fid int
	var err error
	if sfid != "" {
		fid, err = strconv.Atoi(sfid)
		if err != nil {
			return c.LocalError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
		}
	}
	if fid == 0 {
		fid = c.Config.DefaultForum
	}

	ferr := c.ForumUserCheck(header, w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateTopic {
		return c.NoPermissions(w, r, user)
	}
	// TODO: Add a phrase for this
	header.Title = phrases.GetTitlePhrase("create_topic")
	header.Zone = "create_topic"

	// Lock this to the forum being linked?
	// Should we always put it in strictmode when it's linked from another forum? Well, the user might end up changing their mind on what forum they want to post in and it would be a hassle, if they had to switch pages, even if it is a single click for many (exc. mobile)
	var strict bool
	header.Hooks.VhookNoRet("topic_create_pre_loop", w, r, fid, &header, &user, &strict)

	// TODO: Re-add support for plugin_guilds
	var forumList []c.Forum
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = c.Forums.GetAllVisibleIDs()
		if err != nil {
			return c.InternalError(err, w, r)
		}
	} else {
		group, err := c.Groups.Get(user.Group)
		if err != nil {
			// TODO: Refactor this
			c.LocalError("Something weird happened behind the scenes", w, r, user)
			log.Printf("Group #%d doesn't exist, but it's set on c.User #%d", user.Group, user.ID)
			return nil
		}
		canSee = group.CanSee
	}

	// TODO: plugin_superadmin needs to be able to override this loop. Skip flag on topic_create_pre_loop?
	for _, ffid := range canSee {
		// TODO: Surely, there's a better way of doing this. I've added it in for now to support plugin_guilds, but we really need to clean this up
		if strict && ffid != fid {
			continue
		}

		// Do a bulk forum fetch, just in case it's the SqlForumStore?
		forum := c.Forums.DirtyGet(ffid)
		if forum.Name != "" && forum.Active {
			fcopy := forum.Copy()
			// TODO: Abstract this
			if header.Hooks.HookSkippable("topic_create_frow_assign", &fcopy) {
				continue
			}
			forumList = append(forumList, fcopy)
		}
	}

	return renderTemplate("create_topic", w, r, header, c.CreateTopicPage{header, forumList, fid})
}

func CreateTopicSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	fid, err := strconv.Atoi(r.PostFormValue("topic-board"))
	if err != nil {
		return c.LocalError(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}
	// TODO: Add hooks to make use of headerLite
	lite, ferr := c.SimpleForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateTopic {
		return c.NoPermissions(w, r, user)
	}

	tname := c.SanitiseSingleLine(r.PostFormValue("topic-name"))
	content := c.PreparseMessage(r.PostFormValue("topic-content"))
	// TODO: Fully parse the post and store it in the parsed column
	tid, err := c.Topics.Create(fid, tname, content, user.ID, user.LastIP)
	if err != nil {
		switch err {
		case c.ErrNoRows:
			return c.LocalError("Something went wrong, perhaps the forum got deleted?", w, r, user)
		case c.ErrNoTitle:
			return c.LocalError("This topic doesn't have a title", w, r, user)
		case c.ErrLongTitle:
			return c.LocalError("The length of the title is too long, max: "+strconv.Itoa(c.Config.MaxTopicTitleLength), w, r, user)
		case c.ErrNoBody:
			return c.LocalError("This topic doesn't have a body", w, r, user)
		}
		return c.InternalError(err, w, r)
	}

	topic, err := c.Topics.Get(tid)
	if err != nil {
		return c.LocalError("Unable to load the topic", w, r, user)
	}
	if r.PostFormValue("has_poll") == "1" {
		maxPollOptions := 10
		pollInputItems := make(map[int]string)
		for key, values := range r.Form {
			for _, value := range values {
				if !strings.HasPrefix(key, "pollinputitem[") {
					continue
				}
				halves := strings.Split(key, "[")
				if len(halves) != 2 {
					return c.LocalError("Malformed pollinputitem", w, r, user)
				}
				halves[1] = strings.TrimSuffix(halves[1], "]")

				index, err := strconv.Atoi(halves[1])
				if err != nil {
					return c.LocalError("Malformed pollinputitem", w, r, user)
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

		if len(pollInputItems) > 0 {
			// Make sure the indices are sequential to avoid out of bounds issues
			seqPollInputItems := make(map[int]string)
			for i := 0; i < len(pollInputItems); i++ {
				seqPollInputItems[i] = pollInputItems[i]
			}

			pollType := 0 // Basic single choice
			_, err := c.Polls.Create(topic, pollType, seqPollInputItems)
			if err != nil {
				return c.LocalError("Failed to add poll to topic", w, r, user) // TODO: Might need to be an internal error as it could leave phantom polls?
			}
		}
	}

	err = c.Subscriptions.Add(user.ID, tid, "topic")
	if err != nil {
		return c.InternalError(err, w, r)
	}

	err = user.IncreasePostStats(c.WordCount(content), true)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// Handle the file attachments
	if user.Perms.UploadFiles {
		_, rerr := uploadAttachment(w, r, user, fid, "forums", tid, "topics", "")
		if rerr != nil {
			return rerr
		}
	}

	counters.PostCounter.Bump()
	counters.TopicCounter.Bump()
	// TODO: Pass more data to this hook?
	skip, rerr := lite.Hooks.VhookSkippable("action_end_create_topic", tid, &user)
	if skip || rerr != nil {
		return rerr
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func uploadFilesWithHash(w http.ResponseWriter, r *http.Request, user c.User, dir string) (filenames []string, rerr c.RouteError) {
	files, ok := r.MultipartForm.File["upload_files"]
	if !ok {
		return nil, nil
	}
	if len(files) > 5 {
		return nil, c.LocalError("You can't attach more than five files", w, r, user)
	}

	for _, file := range files {
		if file.Filename == "" {
			continue
		}
		//c.DebugLog("file.Filename ", file.Filename)

		extarr := strings.Split(file.Filename, ".")
		if len(extarr) < 2 {
			return nil, c.LocalError("Bad file", w, r, user)
		}
		ext := extarr[len(extarr)-1]

		// TODO: Can we do this without a regex?
		reg, err := regexp.Compile("[^A-Za-z0-9]+")
		if err != nil {
			return nil, c.LocalError("Bad file extension", w, r, user)
		}
		ext = strings.ToLower(reg.ReplaceAllString(ext, ""))
		if !c.AllowedFileExts.Contains(ext) {
			return nil, c.LocalError("You're not allowed to upload files with this extension", w, r, user)
		}

		inFile, err := file.Open()
		if err != nil {
			return nil, c.LocalError("Upload failed", w, r, user)
		}
		defer inFile.Close()

		hasher := sha256.New()
		_, err = io.Copy(hasher, inFile)
		if err != nil {
			return nil, c.LocalError("Upload failed [Hashing Failed]", w, r, user)
		}
		inFile.Close()

		checksum := hex.EncodeToString(hasher.Sum(nil))
		filename := checksum + "." + ext

		inFile, err = file.Open()
		if err != nil {
			return nil, c.LocalError("Upload failed", w, r, user)
		}
		defer inFile.Close()

		if ext != "jpg" && ext != "png" && ext != "gif" {
			outFile, err := os.Create(dir + filename)
			if err != nil {
				return nil, c.LocalError("Upload failed [File Creation Failed]", w, r, user)
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return nil, c.LocalError("Upload failed [Copy Failed]", w, r, user)
			}
		} else {
			img, _, err := image.Decode(inFile)
			if err != nil {
				return nil, c.LocalError("Upload failed [Image Decoding Failed]",w,r,user)
			}

			outFile, err := os.Create(dir + filename)
			if err != nil {
				return nil, c.LocalError("Upload failed [File Creation Failed]", w, r, user)
			}
			defer outFile.Close()
			
			switch ext {
			case "gif":
				err = gif.Encode(outFile, img, nil)
			case "png":
				err = png.Encode(outFile, img)
			default:
				err = jpeg.Encode(outFile, img, nil)
			}
			if err != nil {
				return nil, c.LocalError("Upload failed [Image Encoding Failed]", w,r,user)
			}
		}

		filenames = append(filenames, filename)
	}

	return filenames, nil
}

// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
// TODO: Disable stat updates in posts handled by plugin_guilds
func EditTopicSubmit(w http.ResponseWriter, r *http.Request, user c.User, stid string) c.RouteError {
	js := (r.PostFormValue("js") == "1")
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return c.PreErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, js)
	}

	topic, err := c.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The topic you tried to edit doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Add hooks to make use of headerLite
	lite, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic {
		return c.NoPermissionsJSQ(w, r, user, js)
	}
	if topic.IsClosed && !user.Perms.CloseTopic {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	err = topic.Update(r.PostFormValue("topic_name"), r.PostFormValue("topic_content"))
	// TODO: Avoid duplicating this across this route and the topic creation route
	if err != nil {
		switch err {
		case c.ErrNoTitle:
			return c.LocalErrorJSQ("This topic doesn't have a title", w, r, user, js)
		case c.ErrLongTitle:
			return c.LocalErrorJSQ("The length of the title is too long, max: "+strconv.Itoa(c.Config.MaxTopicTitleLength), w, r, user, js)
		case c.ErrNoBody:
			return c.LocalErrorJSQ("This topic doesn't have a body", w, r, user, js)
		}
		return c.InternalErrorJSQ(err, w, r, js)
	}

	err = c.Forums.UpdateLastTopic(topic.ID, user.ID, topic.ParentID)
	if err != nil && err != sql.ErrNoRows {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// TODO: Avoid the load to get this faster?
	topic, err = c.Topics.Get(topic.ID)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The updated topic doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	skip, rerr := lite.Hooks.VhookSkippable("action_end_edit_topic", topic.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	if !js {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		outBytes, err := json.Marshal(JsonReply{c.ParseMessage(topic.Content, topic.ParentID, "forums")})
		if err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
		}
		w.Write(outBytes)
	}
	return nil
}

// TODO: Add support for soft-deletion and add a permission for hard delete in addition to the usual
// TODO: Disable stat updates in posts handled by plugin_guilds
func DeleteTopicSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	// TODO: Move this to some sort of middleware
	var tids []int
	js := false
	if c.ReqIsJson(r) {
		if r.Body == nil {
			return c.PreErrorJS("No request body", w, r)
		}
		err := json.NewDecoder(r.Body).Decode(&tids)
		if err != nil {
			return c.PreErrorJS("We weren't able to parse your data", w, r)
		}
		js = true
	} else {
		tid, err := strconv.Atoi(r.URL.Path[len("/topic/delete/submit/"):])
		if err != nil {
			return c.PreError("The provided TopicID is not a valid number.", w, r)
		}
		tids = append(tids, tid)
	}
	if len(tids) == 0 {
		return c.LocalErrorJSQ("You haven't provided any IDs", w, r, user, js)
	}

	for _, tid := range tids {
		topic, err := c.Topics.Get(tid)
		if err == sql.ErrNoRows {
			return c.PreErrorJSQ("The topic you tried to delete doesn't exist.", w, r, js)
		} else if err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
		}

		// TODO: Add hooks to make use of headerLite
		lite, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
		if ferr != nil {
			return ferr
		}
		if topic.CreatedBy != user.ID {
			if !user.Perms.ViewTopic || !user.Perms.DeleteTopic {
				return c.NoPermissionsJSQ(w, r, user, js)
			}
		}

		// We might be able to handle this err better
		err = topic.Delete()
		if err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
		}

		err = c.ModLogs.Create("delete", tid, "topic", user.LastIP, user.ID)
		if err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
		}

		// ? - We might need to add soft-delete before we can do an action reply for this
		/*_, err = stmts.createActionReply.Exec(tid,"delete",ipaddress,user.ID)
		if err != nil {
			return c.InternalErrorJSQ(err,w,r,js)
		}*/

		// TODO: Do a bulk delete action hook?
		skip, rerr := lite.Hooks.VhookSkippable("action_end_delete_topic", topic.ID, &user)
		if skip || rerr != nil {
			return rerr
		}

		log.Printf("Topic #%d was deleted by UserID #%d", tid, user.ID)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func StickTopicSubmit(w http.ResponseWriter, r *http.Request, user c.User, stid string) c.RouteError {
	topic, lite, rerr := topicActionPre(stid, "pin", w, r, user)
	if rerr != nil {
		return rerr
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		return c.NoPermissions(w, r, user)
	}
	return topicActionPost(topic.Stick(), "stick", w, r, lite, topic, user)
}

func topicActionPre(stid string, action string, w http.ResponseWriter, r *http.Request, user c.User) (*c.Topic, *c.HeaderLite, c.RouteError) {
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return nil, nil, c.PreError(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
	}

	t, err := c.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return nil, nil, c.PreError("The topic you tried to "+action+" doesn't exist.", w, r)
	} else if err != nil {
		return nil, nil, c.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	lite, ferr := c.SimpleForumUserCheck(w, r, &user, t.ParentID)
	if ferr != nil {
		return nil, nil, ferr
	}
	return t, lite, nil
}

func topicActionPost(err error, action string, w http.ResponseWriter, r *http.Request, lite *c.HeaderLite, topic *c.Topic, user c.User) c.RouteError {
	if err != nil {
		return c.InternalError(err, w, r)
	}
	err = addTopicAction(action, topic, user)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	skip, rerr := lite.Hooks.VhookSkippable("action_end_"+action+"_topic", topic.ID, &user)
	if skip || rerr != nil {
		return rerr
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(topic.ID), http.StatusSeeOther)
	return nil
}

func UnstickTopicSubmit(w http.ResponseWriter, r *http.Request, u c.User, stid string) c.RouteError {
	t, lite, rerr := topicActionPre(stid, "unpin", w, r, u)
	if rerr != nil {
		return rerr
	}
	if !u.Perms.ViewTopic || !u.Perms.PinTopic {
		return c.NoPermissions(w, r, u)
	}
	return topicActionPost(t.Unstick(), "unstick", w, r, lite, t, u)
}

func LockTopicSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	// TODO: Move this to some sort of middleware
	var tids []int
	js := false
	if c.ReqIsJson(r) {
		if r.Body == nil {
			return c.PreErrorJS("No request body", w, r)
		}
		err := json.NewDecoder(r.Body).Decode(&tids)
		if err != nil {
			return c.PreErrorJS("We weren't able to parse your data", w, r)
		}
		js = true
	} else {
		tid, err := strconv.Atoi(r.URL.Path[len("/topic/lock/submit/"):])
		if err != nil {
			return c.PreError("The provided TopicID is not a valid number.", w, r)
		}
		tids = append(tids, tid)
	}
	if len(tids) == 0 {
		return c.LocalErrorJSQ("You haven't provided any IDs", w, r, user, js)
	}

	for _, tid := range tids {
		topic, err := c.Topics.Get(tid)
		if err == sql.ErrNoRows {
			return c.PreErrorJSQ("The topic you tried to lock doesn't exist.", w, r, js)
		} else if err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
		}

		// TODO: Add hooks to make use of headerLite
		lite, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.CloseTopic {
			return c.NoPermissionsJSQ(w, r, user, js)
		}

		err = topic.Lock()
		if err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
		}

		err = addTopicAction("lock", topic, user)
		if err != nil {
			return c.InternalErrorJSQ(err, w, r, js)
		}

		// TODO: Do a bulk lock action hook?
		skip, rerr := lite.Hooks.VhookSkippable("action_end_lock_topic", topic.ID, &user)
		if skip || rerr != nil {
			return rerr
		}
	}

	if len(tids) == 1 {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tids[0]), http.StatusSeeOther)
	}
	return nil
}

func UnlockTopicSubmit(w http.ResponseWriter, r *http.Request, u c.User, stid string) c.RouteError {
	t, lite, rerr := topicActionPre(stid, "unlock", w, r, u)
	if rerr != nil {
		return rerr
	}
	if !u.Perms.ViewTopic || !u.Perms.CloseTopic {
		return c.NoPermissions(w, r, u)
	}
	return topicActionPost(t.Unlock(), "unlock", w, r, lite, t, u)
}

// ! JS only route
// TODO: Figure a way to get this route to work without JS
func MoveTopicSubmit(w http.ResponseWriter, r *http.Request, user c.User, sfid string) c.RouteError {
	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.PreErrorJS(phrases.GetErrorPhrase("id_must_be_integer"), w, r)
	}

	// TODO: Move this to some sort of middleware
	var tids []int
	if r.Body == nil {
		return c.PreErrorJS("No request body", w, r)
	}
	err = json.NewDecoder(r.Body).Decode(&tids)
	if err != nil {
		return c.PreErrorJS("We weren't able to parse your data", w, r)
	}
	if len(tids) == 0 {
		return c.LocalErrorJS("You haven't provided any IDs", w, r)
	}

	for _, tid := range tids {
		topic, err := c.Topics.Get(tid)
		if err == sql.ErrNoRows {
			return c.PreErrorJS("The topic you tried to move doesn't exist.", w, r)
		} else if err != nil {
			return c.InternalErrorJS(err, w, r)
		}

		// TODO: Add hooks to make use of headerLite
		_, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.MoveTopic {
			return c.NoPermissionsJS(w, r, user)
		}
		lite, ferr := c.SimpleForumUserCheck(w, r, &user, fid)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.MoveTopic {
			return c.NoPermissionsJS(w, r, user)
		}

		err = topic.MoveTo(fid)
		if err != nil {
			return c.InternalErrorJS(err, w, r)
		}

		// ? - Is there a better way of doing this?
		err = addTopicAction("move-"+strconv.Itoa(fid), topic, user)
		if err != nil {
			return c.InternalErrorJS(err, w, r)
		}

		// TODO: Do a bulk move action hook?
		skip, rerr := lite.Hooks.VhookSkippable("action_end_move_topic", topic.ID, &user)
		if skip || rerr != nil {
			return rerr
		}
	}

	if len(tids) == 1 {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tids[0]), http.StatusSeeOther)
	}
	return nil
}

func addTopicAction(action string, t *c.Topic, u c.User) error {
	err := c.ModLogs.Create(action, t.ID, "topic", u.LastIP, u.ID)
	if err != nil {
		return err
	}
	return t.CreateActionReply(action, u.LastIP, u.ID)
}

// TODO: Refactor this
func LikeTopicSubmit(w http.ResponseWriter, r *http.Request, user c.User, stid string) c.RouteError {
	js := r.PostFormValue("js") == "1"
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return c.PreErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, js)
	}

	topic, err := c.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The requested topic doesn't exist.", w, r, js)
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
	if topic.CreatedBy == user.ID {
		return c.LocalErrorJSQ("You can't like your own topics", w, r, user, js)
	}

	_, err = c.Users.Get(topic.CreatedBy)
	if err != nil && err == sql.ErrNoRows {
		return c.LocalErrorJSQ("The target user doesn't exist", w, r, user, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	score := 1
	err = topic.Like(score, user.ID)
	if err == c.ErrAlreadyLiked {
		return c.LocalErrorJSQ("You already liked this", w, r, user, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	// ! Be careful about leaking per-route permission state with &user
	alert := c.Alert{ActorID: user.ID, TargetUserID: topic.CreatedBy, Event: "like", ElementType: "topic", ElementID: tid, Actor: &user}
	err = c.AddActivityAndNotifyTarget(alert)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	skip, rerr := lite.Hooks.VhookSkippable("action_end_like_topic", topic.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	if !js {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}
