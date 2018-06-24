package common

import (
	"strconv"
	"sync"
	"sync/atomic"

	"../query_gen/lib"
)

var TopicList TopicListInt

type TopicListHolder struct {
	List      []*TopicsRow
	ForumList []Forum
	Paginator Paginator
}

type TopicListInt interface {
	GetListByCanSee(canSee []int, page int) (topicList []*TopicsRow, forumList []Forum, paginator Paginator, err error)
	GetListByGroup(group *Group, page int) (topicList []*TopicsRow, forumList []Forum, paginator Paginator, err error)
	GetList(page int) (topicList []*TopicsRow, forumList []Forum, paginator Paginator, err error)

	RebuildPermTree() error
}

type DefaultTopicList struct {
	// TODO: Rewrite this to put permTree as the primary and put canSeeStr on each group?
	oddGroups  map[int]*TopicListHolder
	evenGroups map[int]*TopicListHolder
	oddLock    sync.RWMutex
	evenLock   sync.RWMutex

	permTree atomic.Value // [string(canSee)]canSee
	//permTree map[string][]int // [string(canSee)]canSee
}

// We've removed the topic list cache cap as admins really shouldn't be abusing groups like this with plugin_guilds around and it was extremely fiddly.
// If this becomes a problem later on, then we can revisit this with a fresh perspective, particularly with regards to what people expect a group to really be
// Also, keep in mind that as-long as the groups don't all have unique sets of forums they can see, then we can optimise a large portion of the work away.
func NewDefaultTopicList() (*DefaultTopicList, error) {
	tList := &DefaultTopicList{
		oddGroups:  make(map[int]*TopicListHolder),
		evenGroups: make(map[int]*TopicListHolder),
	}

	err := tList.RebuildPermTree()
	if err != nil {
		return nil, err
	}
	err = tList.Tick()
	if err != nil {
		return nil, err
	}

	AddScheduledHalfSecondTask(tList.Tick)
	//AddScheduledSecondTask(tList.GroupCountTick) // TODO: Dynamically change the groups in the short list to be optimised every second
	return tList, nil
}

func (tList *DefaultTopicList) Tick() error {
	var oddLists = make(map[int]*TopicListHolder)
	var evenLists = make(map[int]*TopicListHolder)

	var addList = func(gid int, holder *TopicListHolder) {
		if gid%2 == 0 {
			evenLists[gid] = holder
		} else {
			oddLists[gid] = holder
		}
	}

	var canSeeHolders = make(map[string]*TopicListHolder)
	for name, canSee := range tList.permTree.Load().(map[string][]int) {
		topicList, forumList, paginator, err := tList.GetListByCanSee(canSee, 1)
		if err != nil {
			return err
		}
		canSeeHolders[name] = &TopicListHolder{topicList, forumList, paginator}
	}

	allGroups, err := Groups.GetAll()
	if err != nil {
		return err
	}
	for _, group := range allGroups {
		// ? - Move the user count check to instance initialisation? Might require more book-keeping, particularly when a user moves into a zero user group
		if group.UserCount == 0 && group.ID != GuestUser.Group {
			continue
		}
		var canSee = make([]byte, len(group.CanSee))
		for i, item := range group.CanSee {
			canSee[i] = byte(item)
		}
		addList(group.ID, canSeeHolders[string(canSee)])
	}

	tList.oddLock.Lock()
	tList.oddGroups = oddLists
	tList.oddLock.Unlock()

	tList.evenLock.Lock()
	tList.evenGroups = evenLists
	tList.evenLock.Unlock()

	return nil
}

func (tList *DefaultTopicList) RebuildPermTree() error {
	// TODO: Do something more efficient than this
	allGroups, err := Groups.GetAll()
	if err != nil {
		return err
	}

	var permTree = make(map[string][]int) // [string(canSee)]canSee
	for _, group := range allGroups {
		var canSee = make([]byte, len(group.CanSee))
		for i, item := range group.CanSee {
			canSee[i] = byte(item)
		}
		var canSeeInt = make([]int, len(canSee))
		copy(canSeeInt, group.CanSee)
		permTree[string(canSee)] = canSeeInt
	}
	tList.permTree.Store(permTree)

	return nil
}

func (tList *DefaultTopicList) GetListByGroup(group *Group, page int) (topicList []*TopicsRow, forumList []Forum, paginator Paginator, err error) {
	// TODO: Cache the first three pages not just the first along with all the topics on this beaten track
	if page == 1 {
		var holder *TopicListHolder
		var ok bool
		if group.ID%2 == 0 {
			tList.evenLock.RLock()
			holder, ok = tList.evenGroups[group.ID]
			tList.evenLock.RUnlock()
		} else {
			tList.oddLock.RLock()
			holder, ok = tList.oddGroups[group.ID]
			tList.oddLock.RUnlock()
		}
		if ok {
			return holder.List, holder.ForumList, holder.Paginator, nil
		}
	}

	// TODO: Make CanSee a method on *Group with a canSee field? Have a CanSee method on *User to cover the case of superadmins?
	return tList.GetListByCanSee(group.CanSee, page)
}

func (tList *DefaultTopicList) GetListByCanSee(canSee []int, page int) (topicList []*TopicsRow, forumList []Forum, paginator Paginator, err error) {
	// We need a list of the visible forums for Quick Topic
	// ? - Would it be useful, if we could post in social groups from /topics/?
	for _, fid := range canSee {
		forum := Forums.DirtyGet(fid)
		if forum.Name != "" && forum.Active && (forum.ParentType == "" || forum.ParentType == "forum") {
			fcopy := forum.Copy()
			// TODO: Add a hook here for plugin_guilds
			forumList = append(forumList, fcopy)
		}
	}

	// ? - Should we be showing plugin_guilds posts on /topics/?
	argList, qlist := ForumListToArgQ(forumList)
	if qlist == "" {
		// We don't want to kill the page, so pass an empty slice and nil error
		return topicList, forumList, Paginator{[]int{}, 1, 1}, nil
	}

	topicList, paginator, err = tList.getList(page, argList, qlist)
	return topicList, forumList, paginator, err
}

// TODO: Reduce the number of returns
func (tList *DefaultTopicList) GetList(page int) (topicList []*TopicsRow, forumList []Forum, paginator Paginator, err error) {
	// TODO: Make CanSee a method on *Group with a canSee field? Have a CanSee method on *User to cover the case of superadmins?
	canSee, err := Forums.GetAllVisibleIDs()
	if err != nil {
		return nil, nil, Paginator{nil, 1, 1}, err
	}

	// We need a list of the visible forums for Quick Topic
	// ? - Would it be useful, if we could post in social groups from /topics/?
	for _, fid := range canSee {
		forum := Forums.DirtyGet(fid)
		if forum.Name != "" && forum.Active && (forum.ParentType == "" || forum.ParentType == "forum") {
			fcopy := forum.Copy()
			// TODO: Add a hook here for plugin_guilds
			forumList = append(forumList, fcopy)
		}
	}

	// ? - Should we be showing plugin_guilds posts on /topics/?
	argList, qlist := ForumListToArgQ(forumList)
	if qlist == "" {
		// If the super admin can't see anything, then things have gone terribly wrong
		return topicList, forumList, Paginator{[]int{}, 1, 1}, err
	}

	topicList, paginator, err = tList.getList(page, argList, qlist)
	return topicList, forumList, paginator, err
}

// TODO: Rename this to TopicListStore and pass back a TopicList instance holding the pagination data and topic list rather than passing them back one argument at a time
func (tList *DefaultTopicList) getList(page int, argList []interface{}, qlist string) (topicList []*TopicsRow, paginator Paginator, err error) {
	topicCount, err := ArgQToTopicCount(argList, qlist)
	if err != nil {
		return nil, Paginator{nil, 1, 1}, err
	}
	offset, page, lastPage := PageOffset(topicCount, page, Config.ItemsPerPage)

	stmt, err := qgen.Builder.SimpleSelect("topics", "tid, title, content, createdBy, is_closed, sticky, createdAt, lastReplyAt, lastReplyBy, parentID, postCount, likeCount", "parentID IN("+qlist+")", "sticky DESC, lastReplyAt DESC, createdBy DESC", "?,?")
	if err != nil {
		return nil, Paginator{nil, 1, 1}, err
	}
	defer stmt.Close()

	argList = append(argList, offset)
	argList = append(argList, Config.ItemsPerPage)

	rows, err := stmt.Query(argList...)
	if err != nil {
		return nil, Paginator{nil, 1, 1}, err
	}
	defer rows.Close()

	var reqUserList = make(map[int]bool)
	for rows.Next() {
		topicItem := TopicsRow{ID: 0}
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.IsClosed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.LastReplyBy, &topicItem.ParentID, &topicItem.PostCount, &topicItem.LikeCount)
		if err != nil {
			return nil, Paginator{nil, 1, 1}, err
		}

		topicItem.Link = BuildTopicURL(NameToSlug(topicItem.Title), topicItem.ID)
		// TODO: Pass forum to something like topicItem.Forum and use that instead of these two properties? Could be more flexible.
		forum := Forums.DirtyGet(topicItem.ParentID)
		topicItem.ForumName = forum.Name
		topicItem.ForumLink = forum.Link

		//topicItem.RelativeCreatedAt = RelativeTime(topicItem.CreatedAt)
		topicItem.RelativeLastReplyAt = RelativeTime(topicItem.LastReplyAt)

		// TODO: Rename this Vhook to better reflect moving the topic list from /routes/ to /common/
		RunVhook("topics_topic_row_assign", &topicItem, &forum)
		topicList = append(topicList, &topicItem)
		reqUserList[topicItem.CreatedBy] = true
		reqUserList[topicItem.LastReplyBy] = true
	}
	err = rows.Err()
	if err != nil {
		return nil, Paginator{nil, 1, 1}, err
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TODO: What if a user is deleted via the Control Panel?
	userList, err := Users.BulkGetMap(idSlice)
	if err != nil {
		return nil, Paginator{nil, 1, 1}, err
	}

	// Second pass to the add the user data
	// TODO: Use a pointer to TopicsRow instead of TopicsRow itself?
	for _, topicItem := range topicList {
		topicItem.Creator = userList[topicItem.CreatedBy]
		topicItem.LastUser = userList[topicItem.LastReplyBy]
	}

	pageList := Paginate(topicCount, Config.ItemsPerPage, 5)
	return topicList, Paginator{pageList, page, lastPage}, nil
}

// Internal. Don't rely on it.
func ForumListToArgQ(forums []Forum) (argList []interface{}, qlist string) {
	for _, forum := range forums {
		argList = append(argList, strconv.Itoa(forum.ID))
		qlist += "?,"
	}
	if qlist != "" {
		qlist = qlist[0 : len(qlist)-1]
	}
	return argList, qlist
}

// Internal. Don't rely on it.
func ArgQToTopicCount(argList []interface{}, qlist string) (topicCount int, err error) {
	topicCountStmt, err := qgen.Builder.SimpleCount("topics", "parentID IN("+qlist+")", "")
	if err != nil {
		return 0, err
	}
	defer topicCountStmt.Close()

	err = topicCountStmt.QueryRow(argList...).Scan(&topicCount)
	if err != nil && err != ErrNoRows {
		return 0, err
	}
	return topicCount, err
}

func TopicCountInForums(forums []Forum) (topicCount int, err error) {
	return ArgQToTopicCount(ForumListToArgQ(forums))
}
