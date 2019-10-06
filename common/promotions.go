package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var GroupPromotions GroupPromotionStore

type GroupPromotion struct {
	ID     int
	From   int
	To     int
	TwoWay bool

	Level   int
	Posts   int
	MinTime int
}

type GroupPromotionStore interface {
	GetByGroup(gid int) (gps []*GroupPromotion, err error)
	Get(id int) (*GroupPromotion, error)
	PromoteIfEligible(u *User, level int, posts int) error
	Delete(id int) error
	Create(from int, to int, twoWay bool, level int, posts int) (int, error)
}

type DefaultGroupPromotionStore struct {
	getByGroup *sql.Stmt
	get        *sql.Stmt
	delete     *sql.Stmt
	create     *sql.Stmt

	getByUser  *sql.Stmt
	updateUser *sql.Stmt
}

func NewDefaultGroupPromotionStore(acc *qgen.Accumulator) (*DefaultGroupPromotionStore, error) {
	ugp := "users_groups_promotions"
	return &DefaultGroupPromotionStore{
		getByGroup: acc.Select(ugp).Columns("pid, from_gid, to_gid, two_way, level, posts, minTime").Where("from_gid=? OR to_gid=?").Prepare(),
		get:        acc.Select(ugp).Columns("from_gid, to_gid, two_way, level, posts, minTime").Where("pid = ?").Prepare(),
		delete:     acc.Delete(ugp).Where("pid = ?").Prepare(),
		create:     acc.Insert(ugp).Columns("from_gid, to_gid, two_way, level, posts, minTime").Fields("?,?,?,?,?,?").Prepare(),

		getByUser:  acc.Select(ugp).Columns("pid, to_gid, two_way, level, posts, minTime").Where("from_gid=? AND level<=? AND posts<=?").Orderby("level DESC").Limit("1").Prepare(),
		updateUser: acc.Update("users").Set("group = ?").Where("level >= ? AND posts >= ?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultGroupPromotionStore) GetByGroup(gid int) (gps []*GroupPromotion, err error) {
	rows, err := s.getByGroup.Query(gid, gid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		g := &GroupPromotion{}
		err := rows.Scan(&g.ID, &g.From, &g.To, &g.TwoWay, &g.Level, &g.Posts, &g.MinTime)
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
	err := s.get.QueryRow(id).Scan(&g.From, &g.To, &g.TwoWay, &g.Level, &g.Posts, &g.MinTime)
	if err == nil {
		//s.cache.Set(u)
	}
	return g, err
}

func (s *DefaultGroupPromotionStore) PromoteIfEligible(u *User, level int, posts int) error {
	g := &GroupPromotion{From: u.Group}
	err := s.getByUser.QueryRow(u.Group, level, posts).Scan(&g.ID, &g.To, &g.TwoWay, &g.Level, &g.Posts, &g.MinTime)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}
	_, err = s.updateUser.Exec(g.To, g.Level, g.Posts)
	return err
}

func (s *DefaultGroupPromotionStore) Delete(id int) error {
	_, err := s.delete.Exec(id)
	return err
}

func (s *DefaultGroupPromotionStore) Create(from int, to int, twoWay bool, level int, posts int) (int, error) {
	res, err := s.create.Exec(from, to, twoWay, level, posts, 0)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	return int(lastID), err
}
