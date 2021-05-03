package common

import (
	"database/sql"
	//"log"
	"time"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var GroupPromotions GroupPromotionStore

type GroupPromotion struct {
	ID     int
	From   int
	To     int
	TwoWay bool

	Level         int
	Posts         int
	MinTime       int
	RegisteredFor int
}

type GroupPromotionStore interface {
	GetByGroup(gid int) (gps []*GroupPromotion, err error)
	Get(id int) (*GroupPromotion, error)
	PromoteIfEligible(u *User, level, posts int, registeredAt time.Time) error
	Delete(id int) error
	Create(from, to int, twoWay bool, level, posts, registeredFor int) (int, error)
}

type DefaultGroupPromotionStore struct {
	getByGroup *sql.Stmt
	get        *sql.Stmt
	delete     *sql.Stmt
	create     *sql.Stmt

	getByUser     *sql.Stmt
	getByUserMins *sql.Stmt
	updateUser    *sql.Stmt
	updateGeneric *sql.Stmt
}

func NewDefaultGroupPromotionStore(acc *qgen.Accumulator) (*DefaultGroupPromotionStore, error) {
	ugp := "users_groups_promotions"
	prs := &DefaultGroupPromotionStore{
		getByGroup: acc.Select(ugp).Columns("pid, from_gid, to_gid, two_way, level, posts, minTime, registeredFor").Where("from_gid=? OR to_gid=?").Prepare(),
		get:        acc.Select(ugp).Columns("from_gid, to_gid, two_way, level, posts, minTime, registeredFor").Where("pid=?").Prepare(),
		delete:     acc.Delete(ugp).Where("pid=?").Prepare(),
		create:     acc.Insert(ugp).Columns("from_gid, to_gid, two_way, level, posts, minTime, registeredFor").Fields("?,?,?,?,?,?,?").Prepare(),

		getByUserMins: acc.Select(ugp).Columns("pid, to_gid, two_way, level, posts, minTime, registeredFor").Where("from_gid=? AND level<=? AND posts<=? AND registeredFor<=?").Orderby("level DESC").Limit("1").Prepare(),
		getByUser:     acc.Select(ugp).Columns("pid, to_gid, two_way, level, posts, minTime, registeredFor").Where("from_gid=? AND level<=? AND posts<=?").Orderby("level DESC").Limit("1").Prepare(),
		updateUser:    acc.Update("users").Set("group=?").Where("group=? AND uid=?").Prepare(),
		updateGeneric: acc.Update("users").Set("group=?").Where("group=? AND level>=? AND posts>=?").Prepare(),
	}
	Tasks.FifteenMin.Add(prs.Tick)
	return prs, acc.FirstError()
}

func (s *DefaultGroupPromotionStore) Tick() error {
	return nil
}

func (s *DefaultGroupPromotionStore) GetByGroup(gid int) (gps []*GroupPromotion, err error) {
	rows, err := s.getByGroup.Query(gid, gid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		g := &GroupPromotion{}
		err := rows.Scan(&g.ID, &g.From, &g.To, &g.TwoWay, &g.Level, &g.Posts, &g.MinTime, &g.RegisteredFor)
		if err != nil {
			return nil, err
		}
		gps = append(gps, g)
	}
	return gps, rows.Err()
}

// TODO: Cache the group promotions to avoid hitting the database as much
func (s *DefaultGroupPromotionStore) Get(id int) (*GroupPromotion, error) {
	/*g, err := s.cache.Get(id)
	if err == nil {
		return u, nil
	}*/

	g := &GroupPromotion{ID: id}
	err := s.get.QueryRow(id).Scan(&g.From, &g.To, &g.TwoWay, &g.Level, &g.Posts, &g.MinTime, &g.RegisteredFor)
	if err == nil {
		//s.cache.Set(u)
	}
	return g, err
}

// TODO: Optimise this to avoid the query
func (s *DefaultGroupPromotionStore) PromoteIfEligible(u *User, level, posts int, registeredAt time.Time) error {
	mins := time.Since(registeredAt).Minutes()
	g := &GroupPromotion{From: u.Group}
	//log.Printf("pre getByUserMins: %+v\n", u)
	err := s.getByUserMins.QueryRow(u.Group, level, posts, mins).Scan(&g.ID, &g.To, &g.TwoWay, &g.Level, &g.Posts, &g.MinTime, &g.RegisteredFor)
	if err == sql.ErrNoRows {
		//log.Print("no matches found")
		return nil
	} else if err != nil {
		return err
	}
	//log.Printf("g: %+v\n", g)
	if g.RegisteredFor == 0 {
		_, err = s.updateGeneric.Exec(g.To, g.From, g.Level, g.Posts)
	} else {
		_, err = s.updateUser.Exec(g.To, g.From, u.ID)
	}
	return err
}

func (s *DefaultGroupPromotionStore) Delete(id int) error {
	_, err := s.delete.Exec(id)
	return err
}

func (s *DefaultGroupPromotionStore) Create(from, to int, twoWay bool, level, posts, registeredFor int) (int, error) {
	res, err := s.create.Exec(from, to, twoWay, level, posts, 0, registeredFor)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	return int(lastID), err
}
