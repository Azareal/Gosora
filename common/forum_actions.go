package common

import (
	"database/sql"
	"fmt"
	"strconv"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var ForumActionStore ForumActionStoreInt

//var ForumActionRunnableStore ForumActionRunnableStoreInt

const (
	ForumActionDelete = iota
	ForumActionLock
	ForumActionUnlock
	ForumActionMove
)

func ConvStringToAct(s string) int {
	switch s {
	case "delete":
		return ForumActionDelete
	case "lock":
		return ForumActionLock
	case "unlock":
		return ForumActionUnlock
	case "move":
		return ForumActionMove
	}
	return -1
}
func ConvActToString(a int) string {
	switch a {
	case ForumActionDelete:
		return "delete"
	case ForumActionLock:
		return "lock"
	case ForumActionUnlock:
		return "unlock"
	case ForumActionMove:
		return "move"
	}
	return ""
}

var forumActionStmts ForumActionStmts

type ForumActionStmts struct {
	get1    *sql.Stmt
	get2    *sql.Stmt
	lock1   *sql.Stmt
	lock2   *sql.Stmt
	unlock1 *sql.Stmt
	unlock2 *sql.Stmt
}

type ForumAction struct {
	ID                         int
	Forum                      int
	RunOnTopicCreation         bool
	RunDaysAfterTopicCreation  int
	RunDaysAfterTopicLastReply int
	Action                     int
	Extra                      string
}

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		t := "topics"
		forumActionStmts = ForumActionStmts{
			get1: acc.Select(t).Cols("tid,createdBy,poll").Where("parentID=?").DateOlderThanQ("createdAt", "day").Stmt(),
			get2: acc.Select(t).Cols("tid,createdBy,poll").Where("parentID=?").DateOlderThanQ("lastReplyAt", "day").Stmt(),

			/*lock1:   acc.Update(t).Set("is_closed=1").Where("parentID=?").DateOlderThanQ("createdAt", "day").Stmt(),
			lock2:   acc.Update(t).Set("is_closed=1").Where("parentID=?").DateOlderThanQ("lastReplyAt", "day").Stmt(),
			unlock1: acc.Update(t).Set("is_closed=0").Where("parentID=?").DateOlderThanQ("createdAt", "day").Stmt(),
			unlock2: acc.Update(t).Set("is_closed=0").Where("parentID=?").DateOlderThanQ("lastReplyAt", "day").Stmt(),*/
		}
		return acc.FirstError()
	})
}

func (a *ForumAction) Run() error {
	if a.RunDaysAfterTopicCreation > 0 {
		if e := a.runDaysAfterTopicCreation(); e != nil {
			return e
		}
	}
	if a.RunDaysAfterTopicLastReply > 0 {
		if e := a.runDaysAfterTopicLastReply(); e != nil {
			return e
		}
	}
	return nil
}

func (a *ForumAction) runQ(stmt *sql.Stmt, days int, f func(t *Topic) error) error {
	rows, e := stmt.Query(days, a.Forum)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		// TODO: Decouple this
		t := &Topic{ParentID: a.Forum}
		if e := rows.Scan(&t.ID, &t.CreatedBy, &t.Poll); e != nil {
			return e
		}
		if e = f(t); e != nil {
			return e
		}
	}
	return rows.Err()
}

func (a *ForumAction) runDaysAfterTopicCreation() (e error) {
	switch a.Action {
	case ForumActionDelete:
		// TODO: Bulk delete?
		e = a.runQ(forumActionStmts.get1, a.RunDaysAfterTopicCreation, func(t *Topic) error {
			return t.Delete()
		})
	case ForumActionLock:
		/*_, e := forumActionStmts.lock1.Exec(a.Forum)
		if e != nil {
			return e
		}*/
		// TODO: Bulk lock? Lock and get resultset of changed topics somehow?
		fmt.Println("ForumActionLock")
		e = a.runQ(forumActionStmts.get1, a.RunDaysAfterTopicCreation, func(t *Topic) error {
			fmt.Printf("t: %+v\n", t)
			return t.Lock()
		})
	case ForumActionUnlock:
		// TODO: Bulk unlock? Unlock and get resultset of changed topics somehow?
		e = a.runQ(forumActionStmts.get1, a.RunDaysAfterTopicCreation, func(t *Topic) error {
			return t.Unlock()
		})
	case ForumActionMove:
		destForum, e := strconv.Atoi(a.Extra)
		if e != nil {
			return e
		}
		e = a.runQ(forumActionStmts.get1, a.RunDaysAfterTopicCreation, func(t *Topic) error {
			return t.MoveTo(destForum)
		})
	}
	return e
}

func (a *ForumAction) runDaysAfterTopicLastReply() (e error) {
	switch a.Action {
	case ForumActionDelete:
		e = a.runQ(forumActionStmts.get2, a.RunDaysAfterTopicLastReply, func(t *Topic) error {
			return t.Delete()
		})
	case ForumActionLock:
		// TODO: Bulk lock? Lock and get resultset of changed topics somehow?
		e = a.runQ(forumActionStmts.get2, a.RunDaysAfterTopicLastReply, func(t *Topic) error {
			return t.Lock()
		})
	case ForumActionUnlock:
		// TODO: Bulk unlock? Unlock and get resultset of changed topics somehow?
		e = a.runQ(forumActionStmts.get2, a.RunDaysAfterTopicLastReply, func(t *Topic) error {
			return t.Unlock()
		})
	case ForumActionMove:
		destForum, e := strconv.Atoi(a.Extra)
		if e != nil {
			return e
		}
		e = a.runQ(forumActionStmts.get2, a.RunDaysAfterTopicLastReply, func(t *Topic) error {
			return t.MoveTo(destForum)
		})
	}
	return nil
}

func (a *ForumAction) TopicCreation(tid int) error {
	if !a.RunOnTopicCreation {
		return nil
	}
	return nil
}

type ForumActionStoreInt interface {
	Get(faid int) (*ForumAction, error)
	GetInForum(fid int) ([]*ForumAction, error)
	GetAll() ([]*ForumAction, error)
	GetNewTopicActions(fid int) ([]*ForumAction, error)

	Add(fa *ForumAction) (int, error)
	Delete(faid int) error
	Exists(faid int) bool
	Count() int
	CountInForum(fid int) int

	DailyTick() error
}

type DefaultForumActionStore struct {
	get                *sql.Stmt
	getInForum         *sql.Stmt
	getAll             *sql.Stmt
	getNewTopicActions *sql.Stmt

	add          *sql.Stmt
	delete       *sql.Stmt
	exists       *sql.Stmt
	count        *sql.Stmt
	countInForum *sql.Stmt
}

func NewDefaultForumActionStore(acc *qgen.Accumulator) (*DefaultForumActionStore, error) {
	fa := "forums_actions"
	allCols := "faid,fid,runOnTopicCreation,runDaysAfterTopicCreation,runDaysAfterTopicLastReply,action,extra"
	return &DefaultForumActionStore{
		get:                acc.Select(fa).Columns("fid,runOnTopicCreation,runDaysAfterTopicCreation,runDaysAfterTopicLastReply,action,extra").Where("faid=?").Prepare(),
		getInForum:         acc.Select(fa).Columns("faid,runOnTopicCreation,runDaysAfterTopicCreation,runDaysAfterTopicLastReply,action,extra").Where("fid=?").Prepare(),
		getAll:             acc.Select(fa).Columns(allCols).Prepare(),
		getNewTopicActions: acc.Select(fa).Columns(allCols).Where("fid=? AND runOnTopicCreation=1").Prepare(),

		add:          acc.Insert(fa).Columns("fid,runOnTopicCreation,runDaysAfterTopicCreation,runDaysAfterTopicLastReply,action,extra").Fields("?,?,?,?,?,?").Prepare(),
		delete:       acc.Delete(fa).Where("faid=?").Prepare(),
		exists:       acc.Exists(fa, "faid").Prepare(),
		count:        acc.Count(fa).Prepare(),
		countInForum: acc.Count(fa).Where("fid=?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultForumActionStore) DailyTick() error {
	fas, e := s.GetAll()
	if e != nil {
		return e
	}
	for _, fa := range fas {
		if e := fa.Run(); e != nil {
			return e
		}
	}
	return nil
}

func (s *DefaultForumActionStore) Get(id int) (*ForumAction, error) {
	fa := ForumAction{ID: id}
	var str string
	e := s.get.QueryRow(id).Scan(&fa.Forum, &fa.RunOnTopicCreation, &fa.RunDaysAfterTopicCreation, &fa.RunDaysAfterTopicLastReply, &str, &fa.Extra)
	fa.Action = ConvStringToAct(str)
	return &fa, e
}

func (s *DefaultForumActionStore) GetInForum(fid int) (fas []*ForumAction, e error) {
	rows, e := s.getInForum.Query(fid)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var str string
	for rows.Next() {
		fa := ForumAction{Forum: fid}
		if e := rows.Scan(&fa.ID, &fa.RunOnTopicCreation, &fa.RunDaysAfterTopicCreation, &fa.RunDaysAfterTopicLastReply, &str, &fa.Extra); e != nil {
			return nil, e
		}
		fa.Action = ConvStringToAct(str)
		fas = append(fas, &fa)
	}
	return fas, rows.Err()
}

func (s *DefaultForumActionStore) GetAll() (fas []*ForumAction, e error) {
	rows, e := s.getAll.Query()
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var str string
	for rows.Next() {
		fa := ForumAction{}
		if e := rows.Scan(&fa.ID, &fa.Forum, &fa.RunOnTopicCreation, &fa.RunDaysAfterTopicCreation, &fa.RunDaysAfterTopicLastReply, &str, &fa.Extra); e != nil {
			return nil, e
		}
		fa.Action = ConvStringToAct(str)
		fas = append(fas, &fa)
	}
	return fas, rows.Err()
}

func (s *DefaultForumActionStore) GetNewTopicActions(fid int) (fas []*ForumAction, e error) {
	rows, e := s.getNewTopicActions.Query(fid)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var str string
	for rows.Next() {
		fa := ForumAction{RunOnTopicCreation: true}
		if e := rows.Scan(&fa.ID, &fa.Forum, &fa.RunDaysAfterTopicCreation, &fa.RunDaysAfterTopicLastReply, &str, &fa.Extra); e != nil {
			return nil, e
		}
		fa.Action = ConvStringToAct(str)
		fas = append(fas, &fa)
	}
	return fas, rows.Err()
}

func (s *DefaultForumActionStore) Add(fa *ForumAction) (int, error) {
	res, e := s.add.Exec(fa.Forum, fa.RunOnTopicCreation, fa.RunDaysAfterTopicCreation, fa.RunDaysAfterTopicLastReply, ConvActToString(fa.Action), fa.Extra)
	if e != nil {
		return 0, e
	}
	lastID, e := res.LastInsertId()
	return int(lastID), e
}

func (s *DefaultForumActionStore) Delete(id int) error {
	_, e := s.delete.Exec(id)
	return e
}

func (s *DefaultForumActionStore) Exists(id int) bool {
	err := s.exists.QueryRow(id).Scan(&id)
	if err != nil && err != ErrNoRows {
		LogError(err)
	}
	return err != ErrNoRows
}

func (s *DefaultForumActionStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *DefaultForumActionStore) CountInForum(fid int) (count int) {
	return Countf(s.countInForum, fid)
}

/*type ForumActionRunnable struct {
	ID         int
	ActionID   int
	TargetID   int
	TargetType int // 0 = topic
	RunAfter   int //unixtime
}

type ForumActionRunnableStoreInt interface {
	GetAfterTime(unix int) ([]*ForumActionRunnable, error)
	GetInForum(fid int) ([]*ForumActionRunnable, error)
	Delete(faid int) error
	DeleteInForum(fid int) error
	DeleteByActionID(faid int) error
	Count() int
	CountInForum(fid int) int
}

type DefaultForumActionRunnableStore struct {
	delete        *sql.Stmt
	deleteInForum *sql.Stmt
	count         *sql.Stmt
	countInForum  *sql.Stmt
}

func NewDefaultForumActionRunnableStore(acc *qgen.Accumulator) (*DefaultForumActionRunnableStore, error) {
	fa := "forums_actions"
	return &DefaultForumActionRunnableStore{
		delete:        acc.Delete(fa).Where("faid=?").Prepare(),
		deleteInForum: acc.Delete(fa).Where("fid=?").Prepare(),
		count:         acc.Count(fa).Prepare(),
		countInForum:  acc.Count(fa).Where("faid=?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultForumActionRunnableStore) Delete(id int) error {
	_, e := s.delete.Exec(id)
	return e
}

func (s *DefaultForumActionRunnableStore) DeleteInForum(fid int) error {
	_, e := s.deleteInForum.Exec(id)
	return e
}

func (s *DefaultForumActionRunnableStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *DefaultForumActionRunnableStore) CountInForum(fid int) (count int) {
	return Countf(s.countInForum, fid)
}
*/
