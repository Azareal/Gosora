package common

import (
	"errors"
	"time"

	//"log"

	"database/sql"
	"strconv"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Convos ConversationStore
var convoStmts ConvoStmts

type ConvoStmts struct {
	fetchPost  *sql.Stmt
	getPosts   *sql.Stmt
	countPosts *sql.Stmt
	edit       *sql.Stmt
	create     *sql.Stmt
	delete     *sql.Stmt
	has        *sql.Stmt

	editPost   *sql.Stmt
	createPost *sql.Stmt
	deletePost *sql.Stmt

	getUsers *sql.Stmt
}

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		cpo := "conversations_posts"
		convoStmts = ConvoStmts{
			fetchPost:  acc.Select(cpo).Columns("cid,body,post,createdBy").Where("pid=?").Prepare(),
			getPosts:   acc.Select(cpo).Columns("pid,body,post,createdBy").Where("cid=?").Limit("?,?").Prepare(),
			countPosts: acc.Count(cpo).Where("cid=?").Prepare(),
			edit:       acc.Update("conversations").Set("lastReplyBy=?,lastReplyAt=?").Where("cid=?").Prepare(),
			create:     acc.Insert("conversations").Columns("createdAt,lastReplyAt").Fields("UTC_TIMESTAMP(),UTC_TIMESTAMP()").Prepare(),
			has:        acc.Count("conversations_participants").Where("uid=? AND cid=?").Prepare(),

			editPost:   acc.Update(cpo).Set("body=?,post=?").Where("pid=?").Prepare(),
			createPost: acc.Insert(cpo).Columns("cid,body,post,createdBy").Fields("?,?,?,?").Prepare(),
			deletePost: acc.Delete(cpo).Where("pid=?").Prepare(),

			getUsers: acc.Select("conversations_participants").Columns("uid").Where("cid=?").Prepare(),
		}
		return acc.FirstError()
	})
}

type Conversation struct {
	ID          int
	Link        string
	CreatedBy   int
	CreatedAt   time.Time
	LastReplyBy int
	LastReplyAt time.Time
}

func (co *Conversation) Posts(offset, itemsPerPage int) (posts []*ConversationPost, err error) {
	rows, err := convoStmts.getPosts.Query(co.ID, offset, itemsPerPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := &ConversationPost{CID: co.ID}
		err := rows.Scan(&p.ID, &p.Body, &p.Post, &p.CreatedBy)
		if err != nil {
			return nil, err
		}
		p, err = ConvoPostProcess.OnLoad(p)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

func (co *Conversation) PostsCount() (count int) {
	err := convoStmts.countPosts.QueryRow(co.ID).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (co *Conversation) Uids() (ids []int, err error) {
	rows, err := convoStmts.getUsers.Query(co.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (co *Conversation) Has(uid int) (in bool) {
	var count int
	err := convoStmts.has.QueryRow(uid, co.ID).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count > 0
}

func (co *Conversation) Update() error {
	_, err := convoStmts.edit.Exec(co.CreatedAt, co.LastReplyBy, co.LastReplyAt, co.ID)
	return err
}

func (co *Conversation) Create() (int, error) {
	res, err := convoStmts.create.Exec()
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

func BuildConvoURL(coid int) string {
	return "/user/convo/" + strconv.Itoa(coid)
}

type ConversationExtra struct {
	*Conversation
	Users []*User
}

type ConversationStore interface {
	Get(id int) (*Conversation, error)
	GetUser(uid, offset int) (cos []*Conversation, err error)
	GetUserExtra(uid, offset int) (cos []*ConversationExtra, err error)
	GetUserCount(uid int) (count int)
	Delete(id int) error
	Count() (count int)
	Create(content string, createdBy int, participants []int) (int, error)
}

type DefaultConversationStore struct {
	get                *sql.Stmt
	getUser            *sql.Stmt
	getUserCount       *sql.Stmt
	delete             *sql.Stmt
	deletePosts        *sql.Stmt
	deleteParticipants *sql.Stmt
	create             *sql.Stmt
	addParticipant     *sql.Stmt
	count              *sql.Stmt
}

func NewDefaultConversationStore(acc *qgen.Accumulator) (*DefaultConversationStore, error) {
	co := "conversations"
	return &DefaultConversationStore{
		get:                acc.Select(co).Columns("createdBy,createdAt,lastReplyBy,lastReplyAt").Where("cid=?").Prepare(),
		getUser:            acc.SimpleInnerJoin("conversations_participants AS cp", "conversations AS c", "cp.cid, c.createdBy, c.createdAt, c.lastReplyBy, c.lastReplyAt", "cp.cid=c.cid", "cp.uid=?", "c.lastReplyAt DESC, c.createdAt DESC, c.cid DESC", "?,?"),
		getUserCount:       acc.Count("conversations_participants").Where("uid=?").Prepare(),
		delete:             acc.Delete(co).Where("cid=?").Prepare(),
		deletePosts:        acc.Delete("conversations_posts").Where("cid=?").Prepare(),
		deleteParticipants: acc.Delete("conversations_participants").Where("cid=?").Prepare(),
		create:             acc.Insert(co).Columns("createdBy,createdAt,lastReplyBy,lastReplyAt").Fields("?,UTC_TIMESTAMP(),?,UTC_TIMESTAMP()").Prepare(),
		addParticipant:     acc.Insert("conversations_participants").Columns("uid,cid").Fields("?,?").Prepare(),
		count:              acc.Count(co).Prepare(),
	}, acc.FirstError()
}

func (s *DefaultConversationStore) Get(id int) (*Conversation, error) {
	co := &Conversation{ID: id}
	err := s.get.QueryRow(id).Scan(&co.CreatedBy, &co.CreatedAt, &co.LastReplyBy, &co.LastReplyAt)
	co.Link = BuildConvoURL(co.ID)
	return co, err
}

func (s *DefaultConversationStore) GetUser(uid, offset int) (cos []*Conversation, err error) {
	rows, err := s.getUser.Query(uid, offset, Config.ItemsPerPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		co := &Conversation{}
		err := rows.Scan(&co.ID, &co.CreatedBy, &co.CreatedAt, &co.LastReplyBy, &co.LastReplyAt)
		if err != nil {
			return nil, err
		}
		co.Link = BuildConvoURL(co.ID)
		cos = append(cos, co)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	if len(cos) == 0 {
		err = sql.ErrNoRows
	}
	return cos, err
}

func (s *DefaultConversationStore) GetUserExtra(uid, offset int) (cos []*ConversationExtra, err error) {
	raw, err := s.GetUser(uid, offset)
	if err != nil {
		return nil, err
	}
	//log.Printf("raw: %+v\n", raw)

	if len(raw) == 1 {
		//log.Print("r0b2")
		uids, err := raw[0].Uids()
		if err != nil {
			return nil, err
		}
		//log.Println("r1b2")
		umap, err := Users.BulkGetMap(uids)
		if err != nil {
			return nil, err
		}
		//log.Println("r2b2")
		users := make([]*User, len(umap))
		var i int
		for _, user := range umap {
			users[i] = user
			i++
		}
		return []*ConversationExtra{{raw[0], users}}, nil
	}
	//log.Println("1")

	cmap := make(map[int]*ConversationExtra, len(raw))
	for _, co := range raw {
		cmap[co.ID] = &ConversationExtra{co, nil}
	}

	// TODO: Add a function for the q stuff
	var q string
	idList := make([]interface{}, len(raw))
	for i, co := range raw {
		if i == 0 {
			q = "?"
		} else {
			q += ",?"
		}
		idList[i] = strconv.Itoa(co.ID)
	}

	rows, err := qgen.NewAcc().Select("conversations_participants").Columns("uid,cid").Where("cid IN(" + q + ")").Query(idList...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	//log.Println("2")

	idmap := make(map[int][]int) // cid: []uid
	puidmap := make(map[int]struct{})
	for rows.Next() {
		var uid, cid int
		err := rows.Scan(&uid, &cid)
		if err != nil {
			return nil, err
		}
		idmap[cid] = append(idmap[cid], uid)
		puidmap[uid] = struct{}{}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	//log.Println("3")
	//log.Printf("idmap: %+v\n", idmap)
	//log.Printf("puidmap: %+v\n",puidmap)

	puids := make([]int, len(puidmap))
	var i int
	for puid, _ := range puidmap {
		puids[i] = puid
		i++
	}
	umap, err := Users.BulkGetMap(puids)
	if err != nil {
		return nil, err
	}
	//log.Println("4")
	//log.Printf("umap: %+v\n", umap)
	for cid, uids := range idmap {
		co := cmap[cid]
		for _, uid := range uids {
			co.Users = append(co.Users, umap[uid])
		}
		//log.Printf("co.Conversation: %+v\n", co.Conversation)
		//log.Printf("co.Users: %+v\n", co.Users)
		cmap[cid] = co
	}
	//log.Printf("cmap: %+v\n", cmap)
	for _, ra := range raw {
		cos = append(cos, cmap[ra.ID])
	}
	//log.Printf("cos: %+v\n", cos)

	return cos, rows.Err()
}

func (s *DefaultConversationStore) GetUserCount(uid int) (count int) {
	err := s.getUserCount.QueryRow(uid).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

// TODO: Use a foreign key or transaction
func (s *DefaultConversationStore) Delete(id int) error {
	_, err := s.delete.Exec(id)
	if err != nil {
		return err
	}
	_, err = s.deletePosts.Exec(id)
	if err != nil {
		return err
	}
	_, err = s.deleteParticipants.Exec(id)
	return err
}

func (s *DefaultConversationStore) Create(content string, createdBy int, participants []int) (int, error) {
	if len(participants) == 0 {
		return 0, errors.New("no participants set")
	}
	res, err := s.create.Exec(createdBy, createdBy)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	post := &ConversationPost{CID: int(lastID), Body: content, CreatedBy: createdBy}
	_, err = post.Create()
	if err != nil {
		return 0, err
	}

	for _, p := range participants {
		if p == createdBy {
			continue
		}
		_, err := s.addParticipant.Exec(p, lastID)
		if err != nil {
			return 0, err
		}
	}
	_, err = s.addParticipant.Exec(createdBy, lastID)
	if err != nil {
		return 0, err
	}

	return int(lastID), err
}

// Count returns the total number of topics on these forums
func (s *DefaultConversationStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}
