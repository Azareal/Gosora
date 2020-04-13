package common

import (
	//"log"
	"database/sql"
	"strconv"
	"sync"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var TopicList TopicListInt

const (
	TopicListDefault = iota
	TopicListMostViewed
)

type TopicListHolder struct {
	List      []*TopicsRow
	ForumList []Forum
	Paginator Paginator
}

type ForumTopicListHolder struct {
	List      []*TopicsRow
	Paginator Paginator
}

type TopicListInt interface {
	GetListByCanSee(canSee []int, page, orderby int, filterIDs []int) (topicList []*TopicsRow, forumList []Forum, pagi Paginator, err error)
	GetListByGroup(g *Group, page, orderby int, filterIDs []int) (topicList []*TopicsRow, forumList []Forum, pagi Paginator, err error)
	GetListByForum(f *Forum, page, orderby int) (topicList []*TopicsRow, pagi Paginator, err error)
	GetList(page, orderby int, filterIDs []int) (topicList []*TopicsRow, forumList []Forum, pagi Paginator, err error)
}

type DefaultTopicList struct {
	// TODO: Rewrite this to put permTree as the primary and put canSeeStr on each group?
	oddGroups  map[int][2]*TopicListHolder
	evenGroups map[int][2]*TopicListHolder
	oddLock    sync.RWMutex
	evenLock   sync.RWMutex

	forums    map[int]*ForumTopicListHolder
	forumLock sync.RWMutex

	qcounts  map[int]*sql.Stmt
	qcounts2 map[int]*sql.Stmt
	qLock    sync.RWMutex
	qLock2   sync.RWMutex

	//permTree atomic.Value // [string(canSee)]canSee
	//permTree map[string][]int // [string(canSee)]canSee

	getTopicsByForum *sql.Stmt
	//getTidsByForum *sql.Stmt
}

// We've removed the topic list cache cap as admins really shouldn't be abusing groups like this with plugin_guilds around and it was extremely fiddly.
// If this becomes a problem later on, then we can revisit this with a fresh perspective, particularly with regards to what people expect a group to really be
// Also, keep in mind that as-long as the groups don't all have unique sets of forums they can see, then we can optimise a large portion of the work away.
func NewDefaultTopicList(acc *qgen.Accumulator) (*DefaultTopicList, error) {
	tList := &DefaultTopicList{
		oddGroups:        make(map[int][2]*TopicListHolder),
		evenGroups:       make(map[int][2]*TopicListHolder),
		forums:           make(map[int]*ForumTopicListHolder),
		qcounts:          make(map[int]*sql.Stmt),
		qcounts2:         make(map[int]*sql.Stmt),
		getTopicsByForum: acc.Select("topics").Columns("tid,title,content,createdBy,is_closed,sticky,createdAt,lastReplyAt,lastReplyBy,lastReplyID,views,postCount,likeCount").Where("parentID=?").Orderby("sticky DESC,lastReplyAt DESC,createdBy DESC").Limit("?,?").Prepare(),
		//getTidsByForum: acc.Select("topics").Columns("tid").Where("parentID=?").Orderby("sticky DESC,lastReplyAt DESC,createdBy DESC").Limit("?,?").Prepare(),
	}
	if err := acc.FirstError(); err != nil {
		return nil, err
	}

	err := tList.Tick()
	if err != nil {
		return nil, err
	}

	AddScheduledHalfSecondTask(tList.Tick)
	//AddScheduledSecondTask(tList.GroupCountTick) // TODO: Dynamically change the groups in the short list to be optimised every second
	return tList, nil
}

func (tList *DefaultTopicList) Tick() error {
	//fmt.Println("TopicList.Tick")
	if !TopicListThaw.Thawed() {
		return nil
	}
	//fmt.Println("building topic list")

	oddLists := make(map[int][2]*TopicListHolder)
	evenLists := make(map[int][2]*TopicListHolder)
	addList := func(gid int, h [2]*TopicListHolder) {
		if gid%2 == 0 {
			evenLists[gid] = h
		} else {
			oddLists[gid] = h
		}
	}

	allGroups, err := Groups.GetAll()
	if err != nil {
		return err
	}

	gidToCanSee := make(map[int]string)
	permTree := make(map[string][]int) // [string(canSee)]canSee
	for _, g := range allGroups {
		// ? - Move the user count check to instance initialisation? Might require more book-keeping, particularly when a user moves into a zero user group
		if g.UserCount == 0 && g.ID != GuestUser.Group {
			continue
		}

		canSee := make([]byte, len(g.CanSee))
		for i, item := range g.CanSee {
			canSee[i] = byte(item)
		}

		canSeeInt := make([]int, len(canSee))
		copy(canSeeInt, g.CanSee)
		sCanSee := string(canSee)
		permTree[sCanSee] = canSeeInt
		gidToCanSee[g.ID] = sCanSee
	}

	canSeeHolders := make(map[string][2]*TopicListHolder)
	forumCounts := make(map[int]int)
	for name, canSee := range permTree {
		topicList, forumList, pagi, err := tList.GetListByCanSee(canSee, 1, 0, nil)
		if err != nil {
			return err
		}
		topicList2, forumList2, pagi2, err := tList.GetListByCanSee(canSee, 2, 0, nil)
		if err != nil {
			return err
		}
		canSeeHolders[name] = [2]*TopicListHolder{
			&TopicListHolder{topicList, forumList, pagi},
			&TopicListHolder{topicList2, forumList2, pagi2},
		}
		if len(canSee) > 1 {
			forumCounts[len(canSee)] += 1
		}
	}
	for gid, canSee := range gidToCanSee {
		addList(gid, canSeeHolders[canSee])
	}

	tList.oddLock.Lock()
	tList.oddGroups = oddLists
	tList.oddLock.Unlock()

	tList.evenLock.Lock()
	tList.evenGroups = evenLists
	tList.evenLock.Unlock()

	topc := []int{0, 0, 0, 0, 0, 0}
	addC := func(c int) {
		lowI, low := 0, topc[0]
		for i, top := range topc {
			if top < low {
				lowI = i
				low = top
			}
		}
		if c > low {
			topc[lowI] = c
		}
	}
	for forumCount := range forumCounts {
		addC(forumCount)
	}

	qcounts := make(map[int]*sql.Stmt)
	qcounts2 := make(map[int]*sql.Stmt)
	for _, top := range topc {
		if top == 0 {
			continue
		}

		var qlist string
		for i := 0; i < top; i++ {
			if i != 0 {
				qlist += ","
			}
			qlist += "?"
		}
		cols := "tid,title,content,createdBy,is_closed,sticky,createdAt,lastReplyAt,lastReplyBy,lastReplyID,parentID,views,postCount,likeCount,attachCount,poll,data"

		stmt, err := qgen.Builder.SimpleSelect("topics", cols, "parentID IN("+qlist+")", "views DESC,lastReplyAt DESC,createdBy DESC", "?,?")
		if err != nil {
			return err
		}
		qcounts[top] = stmt

		stmt, err = qgen.Builder.SimpleSelect("topics", cols, "parentID IN("+qlist+")", "sticky DESC,lastReplyAt DESC,createdBy DESC", "?,?")
		if err != nil {
			return err
		}
		qcounts2[top] = stmt
	}

	tList.qLock.Lock()
	tList.qcounts = qcounts
	tList.qLock.Unlock()

	tList.qLock2.Lock()
	tList.qcounts2 = qcounts2
	tList.qLock2.Unlock()

	forums, err := Forums.GetAll()
	if err != nil {
		return err
	}

	top8 := []*Forum{nil, nil, nil, nil, nil, nil, nil, nil}
	z := true
	addScore2 := func(f *Forum) {
		for i, top := range top8 {
			if top.TopicCount < f.TopicCount {
				top8[i] = f
				return
			}
		}
	}
	addScore := func(f *Forum) {
		if z {
			for i, top := range top8 {
				if top == nil {
					top8[i] = f
					return
				}
			}
			z = false
			addScore2(f)
		}
		addScore2(f)
	}

	var fshort []*Forum
	for _, f := range forums {
		if f.Name == "" || !f.Active || (f.ParentType != "" && f.ParentType != "forum") {
			continue
		}
		if f.TopicCount == 0 {
			fshort = append(fshort, f)
			continue
		}
		addScore(f)
	}
	for _, f := range top8 {
		if f != nil {
			fshort = append(fshort, f)
		}
	}

	fList := make(map[int]*ForumTopicListHolder)
	for _, f := range fshort {
		topicList, pagi, err := tList.GetListByForum(f, 1, 0)
		if err != nil {
			return err
		}
		fList[f.ID] = &ForumTopicListHolder{topicList, pagi}
	}

	tList.forumLock.Lock()
	tList.forums = fList
	tList.forumLock.Unlock()

	hTbl := GetHookTable()
	_, _ = hTbl.VhookSkippable("tasks_tick_topic_list", tList)

	return nil
}

// TODO: Add Topics() method to *Forum?
// TODO: Implement orderby
func (tList *DefaultTopicList) GetListByForum(f *Forum, page, orderby int) (topicList []*TopicsRow, pagi Paginator, err error) {
	if page == 0 {
		page = 1
	}
	if f.TopicCount == 0 {
		_, page, lastPage := PageOffset(f.TopicCount, page, Config.ItemsPerPage)
		pageList := Paginate(page, lastPage, 5)
		return topicList, Paginator{pageList, page, lastPage}, nil
	}
	if page == 1 && orderby == 0 {
		var h *ForumTopicListHolder
		var ok bool
		tList.forumLock.RLock()
		h, ok = tList.forums[f.ID]
		tList.forumLock.RUnlock()
		if ok {
			return h.List, h.Paginator, nil
		}
	}

	// TODO: Does forum.TopicCount take the deleted items into consideration for guests? We don't have soft-delete yet, only hard-delete
	offset, page, lastPage := PageOffset(f.TopicCount, page, Config.ItemsPerPage)

	rows, err := tList.getTopicsByForum.Query(f.ID, offset, Config.ItemsPerPage)
	if err != nil {
		return nil, Paginator{nil, 1, 1}, err
	}
	defer rows.Close()

	// TODO: Use something other than TopicsRow as we don't need to store the forum name and link on each and every topic item?
	reqUserList := make(map[int]bool)
	for rows.Next() {
		t := TopicsRow{ID: 0}
		err := rows.Scan(&t.ID, &t.Title, &t.Content, &t.CreatedBy, &t.IsClosed, &t.Sticky, &t.CreatedAt, &t.LastReplyAt, &t.LastReplyBy, &t.LastReplyID, &t.ViewCount, &t.PostCount, &t.LikeCount)
		if err != nil {
			return nil, Paginator{nil, 1, 1}, err
		}

		t.ParentID = f.ID
		t.Link = BuildTopicURL(NameToSlug(t.Title), t.ID)
		// TODO: Create a specialised function with a bit less overhead for getting the last page for a post count
		_, _, lastPage := PageOffset(t.PostCount, 1, Config.ItemsPerPage)
		t.LastPage = lastPage

		//header.Hooks.VhookNoRet("forum_trow_assign", &t, &forum)
		topicList = append(topicList, &t)
		reqUserList[t.CreatedBy] = true
		reqUserList[t.LastReplyBy] = true
	}
	if err = rows.Err(); err != nil {
		return nil, Paginator{nil, 1, 1}, err
	}

	// Convert the user ID map to a slice, then bulk load the users
	idSlice := make([]int, len(reqUserList))
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
	for _, t := range topicList {
		t.Creator = userList[t.CreatedBy]
		t.LastUser = userList[t.LastReplyBy]
	}

	pageList := Paginate(page, lastPage, 5)
	return topicList, Paginator{pageList, page, lastPage}, nil
}

func (tList *DefaultTopicList) GetListByGroup(g *Group, page, orderby int, filterIDs []int) (topicList []*TopicsRow, forumList []Forum, pagi Paginator, err error) {
	if page == 0 {
		page = 1
	}
	// TODO: Cache the first three pages not just the first along with all the topics on this beaten track
	// TODO: Move this into CanSee to reduce redundancy
	if (page == 1 || page == 2) && orderby == 0 && len(filterIDs) == 0 {
		var h [2]*TopicListHolder
		var ok bool
		if g.ID%2 == 0 {
			tList.evenLock.RLock()
			h, ok = tList.evenGroups[g.ID]
			tList.evenLock.RUnlock()
		} else {
			tList.oddLock.RLock()
			h, ok = tList.oddGroups[g.ID]
			tList.oddLock.RUnlock()
		}
		if ok {
			return h[page-1].List, h[page-1].ForumList, h[page-1].Paginator, nil
		}
	}

	// TODO: Make CanSee a method on *Group with a canSee field? Have a CanSee method on *User to cover the case of superadmins?
	//log.Printf("deoptimising for %d on page %d\n", g.ID, page)
	return tList.GetListByCanSee(g.CanSee, page, orderby, filterIDs)
}

func (tList *DefaultTopicList) GetListByCanSee(canSee []int, page, orderby int, filterIDs []int) (topicList []*TopicsRow, forumList []Forum, pagi Paginator, err error) {
	// TODO: Optimise this by filtering canSee and then fetching the forums?
	// We need a list of the visible forums for Quick Topic
	// ? - Would it be useful, if we could post in social groups from /topics/?
	for _, fid := range canSee {
		f := Forums.DirtyGet(fid)
		if f.Name != "" && f.Active && (f.ParentType == "" || f.ParentType == "forum") && f.TopicCount != 0 {
			fcopy := f.Copy()
			// TODO: Add a hook here for plugin_guilds
			forumList = append(forumList, fcopy)
		}
	}

	inSlice := func(haystack []int, needle int) bool {
		for _, item := range haystack {
			if needle == item {
				return true
			}
		}
		return false
	}

	var filteredForums []Forum
	if len(filterIDs) > 0 {
		for _, f := range forumList {
			if inSlice(filterIDs, f.ID) {
				filteredForums = append(filteredForums, f)
			}
		}
	} else {
		filteredForums = forumList
	}
	if len(filteredForums) == 1 && orderby == 0 {
		topicList, pagi, err = tList.GetListByForum(&filteredForums[0], page, orderby)
		return topicList, forumList, pagi, err
	}

	var topicCount int
	for _, f := range filteredForums {
		topicCount += f.TopicCount
	}

	// ? - Should we be showing plugin_guilds posts on /topics/?
	argList, qlist := ForumListToArgQ(filteredForums)
	if qlist == "" {
		// We don't want to kill the page, so pass an empty slice and nil error
		return topicList, filteredForums, Paginator{[]int{}, 1, 1}, nil
	}

	topicList, pagi, err = tList.getList(page, orderby, topicCount, argList, qlist)
	return topicList, filteredForums, pagi, err
}

// TODO: Reduce the number of returns
func (tList *DefaultTopicList) GetList(page, orderby int, filterIDs []int) (topicList []*TopicsRow, forumList []Forum, pagi Paginator, err error) {
	// TODO: Make CanSee a method on *Group with a canSee field? Have a CanSee method on *User to cover the case of superadmins?
	cCanSee, err := Forums.GetAllVisibleIDs()
	if err != nil {
		return nil, nil, Paginator{nil, 1, 1}, err
	}

	inSlice := func(haystack []int, needle int) bool {
		for _, item := range haystack {
			if needle == item {
				return true
			}
		}
		return false
	}

	var canSee []int
	if len(filterIDs) > 0 {
		for _, fid := range cCanSee {
			if inSlice(filterIDs, fid) {
				canSee = append(canSee, fid)
			}
		}
	} else {
		canSee = cCanSee
	}

	// We need a list of the visible forums for Quick Topic
	// ? - Would it be useful, if we could post in social groups from /topics/?
	var topicCount int
	for _, fid := range canSee {
		f := Forums.DirtyGet(fid)
		if f.Name != "" && f.Active && (f.ParentType == "" || f.ParentType == "forum") && f.TopicCount != 0 {
			fcopy := f.Copy()
			// TODO: Add a hook here for plugin_guilds
			forumList = append(forumList, fcopy)
			topicCount += fcopy.TopicCount
		}
	}
	if len(forumList) == 1 && orderby == 0 {
		topicList, pagi, err = tList.GetListByForum(&forumList[0], page, orderby)
		return topicList, forumList, pagi, err
	}

	// ? - Should we be showing plugin_guilds posts on /topics/?
	argList, qlist := ForumListToArgQ(forumList)
	if qlist == "" {
		// If the super admin can't see anything, then things have gone terribly wrong
		return topicList, forumList, Paginator{[]int{}, 1, 1}, err
	}

	topicList, pagi, err = tList.getList(page, orderby, topicCount, argList, qlist)
	return topicList, forumList, pagi, err
}

// TODO: Rename this to TopicListStore and pass back a TopicList instance holding the pagination data and topic list rather than passing them back one argument at a time
// TODO: Make orderby an enum of sorts
func (tList *DefaultTopicList) getList(page, orderby, topicCount int, argList []interface{}, qlist string) (topicList []*TopicsRow, paginator Paginator, err error) {
	//log.Printf("argList: %+v\n",argList)
	//log.Printf("qlist: %+v\n",qlist)
	var orderq string
	var stmt *sql.Stmt
	if orderby == TopicListMostViewed {
		tList.qLock.RLock()
		stmt = tList.qcounts[len(argList)-2]
		tList.qLock.RUnlock()
		if stmt == nil {
			orderq = "views DESC,lastReplyAt DESC,createdBy DESC"
		}
	} else {
		tList.qLock2.RLock()
		stmt = tList.qcounts2[len(argList)-2]
		tList.qLock2.RUnlock()
		if stmt == nil {
			orderq = "sticky DESC,lastReplyAt DESC,createdBy DESC"
		}
	}
	offset, page, lastPage := PageOffset(topicCount, page, Config.ItemsPerPage)

	// TODO: Prepare common qlist lengths to speed this up in common cases, prepared statements are prepared lazily anyway, so it probably doesn't matter if we do ten or so
	if stmt == nil {
		stmt, err = qgen.Builder.SimpleSelect("topics", "tid,title,content,createdBy,is_closed,sticky,createdAt,lastReplyAt,lastReplyBy,lastReplyID,parentID,views,postCount,likeCount,attachCount,poll,data", "parentID IN("+qlist+")", orderq, "?,?")
		if err != nil {
			return nil, Paginator{nil, 1, 1}, err
		}
		defer stmt.Close()
	}

	argList = append(argList, offset)
	argList = append(argList, Config.ItemsPerPage)

	rows, err := stmt.Query(argList...)
	if err != nil {
		return nil, Paginator{nil, 1, 1}, err
	}
	defer rows.Close()

	rcache := Rstore.GetCache()
	rcap := rcache.GetCapacity()
	rlen := rcache.Length()
	tc := Topics.GetCache()
	reqUserList := make(map[int]bool)
	for rows.Next() {
		// TODO: Embed Topic structs in TopicsRow to make it easier for us to reuse this work in the topic cache
		t := TopicsRow{}
		err := rows.Scan(&t.ID, &t.Title, &t.Content, &t.CreatedBy, &t.IsClosed, &t.Sticky, &t.CreatedAt, &t.LastReplyAt, &t.LastReplyBy, &t.LastReplyID, &t.ParentID, &t.ViewCount, &t.PostCount, &t.LikeCount, &t.AttachCount, &t.Poll, &t.Data)
		if err != nil {
			return nil, Paginator{nil, 1, 1}, err
		}

		t.Link = BuildTopicURL(NameToSlug(t.Title), t.ID)
		// TODO: Pass forum to something like topicItem.Forum and use that instead of these two properties? Could be more flexible.
		forum := Forums.DirtyGet(t.ParentID)
		t.ForumName = forum.Name
		t.ForumLink = forum.Link

		// TODO: Create a specialised function with a bit less overhead for getting the last page for a post count
		_, _, lastPage := PageOffset(t.PostCount, 1, Config.ItemsPerPage)
		t.LastPage = lastPage

		// TODO: Rename this Vhook to better reflect moving the topic list from /routes/ to /common/
		GetHookTable().Vhook("topics_topic_row_assign", &t, &forum)
		topicList = append(topicList, &t)
		reqUserList[t.CreatedBy] = true
		reqUserList[t.LastReplyBy] = true

		//log.Print("rlen: ", rlen)
		//log.Print("rcap: ", rcap)
		//log.Print("t.PostCount: ", t.PostCount)
		//log.Print("t.PostCount == 2 && rlen < rcap: ", t.PostCount == 2 && rlen < rcap)

		// Avoid the extra queries on topic list pages, if we already have what we want...
		hRids := false
		if tc != nil {
			if t, err := tc.Get(t.ID); err == nil {
				hRids = len(t.Rids) != 0
			}
		}

		if t.PostCount == 2 && rlen < rcap && !hRids && page < 5 {
			rids, err := GetRidsForTopic(t.ID, 0)
			if err != nil {
				return nil, Paginator{nil, 1, 1}, err
			}

			//log.Print("rids: ", rids)
			if len(rids) == 0 {
				continue
			}
			_, _ = Rstore.Get(rids[0])
			rlen++
			t.Rids = []int{rids[0]}
		}

		if tc != nil {
			if _, err := tc.Get(t.ID); err == sql.ErrNoRows {
				_ = tc.Set(t.Topic())
			}
		}
	}
	if err = rows.Err(); err != nil {
		return nil, Paginator{nil, 1, 1}, err
	}

	// Convert the user ID map to a slice, then bulk load the users
	idSlice := make([]int, len(reqUserList))
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
	for _, t := range topicList {
		t.Creator = userList[t.CreatedBy]
		t.LastUser = userList[t.LastReplyBy]
	}

	pageList := Paginate(page, lastPage, 5)
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
// TODO: Check the TopicCount field on the forums instead? Make sure it's in sync first.
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
	for _, f := range forums {
		topicCount += f.TopicCount
	}
	return topicCount, nil
}
