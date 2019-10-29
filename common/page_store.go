package common

import (
	"database/sql"
	"strconv"
	"strings"

	qgen "github.com/Azareal/Gosora/query_gen"
)

type CustomPageStmts struct {
	update *sql.Stmt
	create *sql.Stmt
}

var customPageStmts CustomPageStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		customPageStmts = CustomPageStmts{
			update: acc.Update("pages").Set("name = ?, title = ?, body = ?, allowedGroups = ?, menuID = ?").Where("pid = ?").Prepare(),
			create: acc.Insert("pages").Columns("name, title, body, allowedGroups, menuID").Fields("?,?,?,?,?").Prepare(),
		}
		return acc.FirstError()
	})
}

type CustomPage struct {
	ID            int
	Name          string // TODO: Let admins put pages in "virtual subdirectories"
	Title         string
	Body          string
	AllowedGroups []int
	MenuID        int
}

func BlankCustomPage() *CustomPage {
	return new(CustomPage)
}

func (page *CustomPage) AddAllowedGroup(gid int) {
	page.AllowedGroups = append(page.AllowedGroups, gid)
}

func (page *CustomPage) getRawAllowedGroups() (rawAllowedGroups string) {
	for _, group := range page.AllowedGroups {
		rawAllowedGroups += strconv.Itoa(group) + ","
	}
	if len(rawAllowedGroups) > 0 {
		rawAllowedGroups = rawAllowedGroups[:len(rawAllowedGroups)-1]
	}
	return rawAllowedGroups
}

func (page *CustomPage) Commit() error {
	_, err := customPageStmts.update.Exec(page.Name, page.Title, page.Body, page.getRawAllowedGroups(), page.MenuID, page.ID)
	Pages.Reload(page.ID)
	return err
}

func (page *CustomPage) Create() (int, error) {
	res, err := customPageStmts.create.Exec(page.Name, page.Title, page.Body, page.getRawAllowedGroups(), page.MenuID)
	if err != nil {
		return 0, err
	}

	pid64, err := res.LastInsertId()
	return int(pid64), err
}

var Pages PageStore

// Holds the custom pages, but doesn't include the template pages in /pages/ which are a lot more flexible yet harder to use and which are too risky security-wise to make editable in the Control Panel
type PageStore interface {
	Count() (count int)
	Get(id int) (*CustomPage, error)
	GetByName(name string) (*CustomPage, error)
	GetOffset(offset int, perPage int) (pages []*CustomPage, err error)
	Reload(id int) error
	Delete(id int) error
}

// TODO: Add a cache to this to save on the queries
type DefaultPageStore struct {
	get       *sql.Stmt
	getByName *sql.Stmt
	getOffset *sql.Stmt
	count     *sql.Stmt
	delete    *sql.Stmt
}

func NewDefaultPageStore(acc *qgen.Accumulator) (*DefaultPageStore, error) {
	pa := "pages"
	return &DefaultPageStore{
		get:       acc.Select(pa).Columns("name, title, body, allowedGroups, menuID").Where("pid = ?").Prepare(),
		getByName: acc.Select(pa).Columns("pid, name, title, body, allowedGroups, menuID").Where("name = ?").Prepare(),
		getOffset: acc.Select(pa).Columns("pid, name, title, body, allowedGroups, menuID").Orderby("pid DESC").Limit("?,?").Prepare(),
		count:     acc.Count(pa).Prepare(),
		delete:    acc.Delete(pa).Where("pid = ?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultPageStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *DefaultPageStore) parseAllowedGroups(raw string, page *CustomPage) error {
	if raw == "" {
		return nil
	}
	for _, sgroup := range strings.Split(raw, ",") {
		group, err := strconv.Atoi(sgroup)
		if err != nil {
			return err
		}
		page.AddAllowedGroup(group)
	}
	return nil
}

func (s *DefaultPageStore) Get(id int) (*CustomPage, error) {
	p := &CustomPage{ID: id}
	rawAllowedGroups := ""
	err := s.get.QueryRow(id).Scan(&p.Name, &p.Title, &p.Body, &rawAllowedGroups, &p.MenuID)
	if err != nil {
		return nil, err
	}
	return p, s.parseAllowedGroups(rawAllowedGroups, p)
}

func (s *DefaultPageStore) GetByName(name string) (*CustomPage, error) {
	p := BlankCustomPage()
	rawAllowedGroups := ""
	err := s.getByName.QueryRow(name).Scan(&p.ID, &p.Name, &p.Title, &p.Body, &rawAllowedGroups, &p.MenuID)
	if err != nil {
		return nil, err
	}
	return p, s.parseAllowedGroups(rawAllowedGroups, p)
}

func (s *DefaultPageStore) GetOffset(offset int, perPage int) (pages []*CustomPage, err error) {
	rows, err := s.getOffset.Query(offset, perPage)
	if err != nil {
		return pages, err
	}
	defer rows.Close()

	for rows.Next() {
		p := &CustomPage{ID: 0}
		rawAllowedGroups := ""
		err := rows.Scan(&p.ID, &p.Name, &p.Title, &p.Body, &rawAllowedGroups, &p.MenuID)
		if err != nil {
			return pages, err
		}
		err = s.parseAllowedGroups(rawAllowedGroups, p)
		if err != nil {
			return pages, err
		}
		pages = append(pages, p)
	}
	return pages, rows.Err()
}

// Always returns nil as there's currently no cache
func (s *DefaultPageStore) Reload(id int) error {
	return nil
}

func (s *DefaultPageStore) Delete(id int) error {
	_, err := s.delete.Exec(id)
	return err
}
