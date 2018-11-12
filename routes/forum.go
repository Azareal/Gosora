package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
)

type ForumStmts struct {
	getTopics *sql.Stmt
}

var forumStmts ForumStmts

// TODO: Move these DbInits into *Forum as Topics()
func init() {
	common.DbInits.Add(func(acc *qgen.Accumulator) error {
		forumStmts = ForumStmts{
			getTopics: acc.Select("topics").Columns("tid, title, content, createdBy, is_closed, sticky, createdAt, lastReplyAt, lastReplyBy, parentID, views, postCount, likeCount").Where("parentID = ?").Orderby("sticky DESC, lastReplyAt DESC, createdBy DESC").Limit("?,?").Prepare(),
		}
		return acc.FirstError()
	})
}

func ViewForum(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header, sfid string) common.RouteError {
	page, _ := strconv.Atoi(r.FormValue("page"))
	_, fid, err := ParseSEOURL(sfid)
	if err != nil {
		return common.PreError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r)
	}

	ferr := common.ForumUserCheck(header, w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic {
		return common.NoPermissions(w, r, user)
	}
	header.Zone = "view_forum"

	// TODO: Fix this double-check
	forum, err := common.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return common.NotFound(w, r, header)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}
	header.Title = forum.Name

	// TODO: Does forum.TopicCount take the deleted items into consideration for guests? We don't have soft-delete yet, only hard-delete
	offset, page, lastPage := common.PageOffset(forum.TopicCount, page, common.Config.ItemsPerPage)

	// TODO: Move this to *Forum
	rows, err := forumStmts.getTopics.Query(fid, offset, common.Config.ItemsPerPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	// TODO: Use something other than TopicsRow as we don't need to store the forum name and link on each and every topic item?
	var topicList []*common.TopicsRow
	var reqUserList = make(map[int]bool)
	for rows.Next() {
		var topicItem = common.TopicsRow{ID: 0}
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.IsClosed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.LastReplyBy, &topicItem.ParentID, &topicItem.ViewCount, &topicItem.PostCount, &topicItem.LikeCount)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		topicItem.Link = common.BuildTopicURL(common.NameToSlug(topicItem.Title), topicItem.ID)
		topicItem.RelativeLastReplyAt = common.RelativeTime(topicItem.LastReplyAt)

		header.Hooks.VhookNoRet("forum_trow_assign", &topicItem, &forum)
		topicList = append(topicList, &topicItem)
		reqUserList[topicItem.CreatedBy] = true
		reqUserList[topicItem.LastReplyBy] = true
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TODO: What if a user is deleted via the Control Panel?
	userList, err := common.Users.BulkGetMap(idSlice)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Second pass to the add the user data
	// TODO: Use a pointer to TopicsRow instead of TopicsRow itself?
	for _, topicItem := range topicList {
		topicItem.Creator = userList[topicItem.CreatedBy]
		topicItem.LastUser = userList[topicItem.LastReplyBy]
	}

	pageList := common.Paginate(forum.TopicCount, common.Config.ItemsPerPage, 5)
	pi := common.ForumPage{header, topicList, forum, common.Paginator{pageList, page, lastPage}}
	ferr = renderTemplate("forum", w, r, header, pi)
	counters.ForumViewCounter.Bump(forum.ID)
	return ferr
}
