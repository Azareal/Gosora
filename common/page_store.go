package common

import (
	"database/sql"
	"strconv"
	"strings"

	"../query_gen/lib"
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
	GlobalCount() (pageCount int)
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
	return &DefaultPageStore{
		get:       acc.Select("pages").Columns("name, title, body, allowedGroups, menuID").Where("pid = ?").Prepare(),
		getByName: acc.Select("pages").Columns("pid, name, title, body, allowedGroups, menuID").Where("name = ?").Prepare(),
		getOffset: acc.Select("pages").Columns("pid, name, title, body, allowedGroups, menuID").Orderby("pid DESC").Limit("?,?").Prepare(),
		count:     acc.Count("pages").Prepare(),
		delete:    acc.Delete("pages").Where("pid = ?").Prepare(),
	}, acc.FirstError()
}

func (store *DefaultPageStore) GlobalCount() (pageCount int) {
	err := store.count.QueryRow().Scan(&pageCount)
	if err != nil {
		LogError(err)
	}
	return pageCount
}

func (store *DefaultPageStore) parseAllowedGroups(raw string, page *CustomPage) error {
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

func (store *DefaultPageStore) Get(id int) (*CustomPage, error) {
	page := &CustomPage{ID: id}
	rawAllowedGroups := ""
	err := store.get.QueryRow(id).Scan(&page.Name, &page.Title, &page.Body, &rawAllowedGroups, &page.MenuID)
	if err != nil {
		return nil, err
	}
	return page, store.parseAllowedGroups(rawAllowedGroups, page)
}

func (store *DefaultPageStore) GetByName(name string) (*CustomPage, error) {
	page := BlankCustomPage()
	rawAllowedGroups := ""
	err := store.getByName.QueryRow(name).Scan(&page.ID, &page.Name, &page.Title, &page.Body, &rawAllowedGroups, &page.MenuID)
	if err != nil {
		return nil, err
	}
	return page, store.parseAllowedGroups(rawAllowedGroups, page)
}

func (store *DefaultPageStore) GetOffset(offset int, perPage int) (pages []*CustomPage, err error) {
	rows, err := store.getOffset.Query(offset, perPage)
	if err != nil {
		return pages, err
	}
	defer rows.Close()

	for rows.Next() {
		page := &CustomPage{ID: 0}
		rawAllowedGroups := ""
		err := rows.Scan(&page.ID, &page.Name, &page.Title, &page.Body, &rawAllowedGroups, &page.MenuID)
		if err != nil {
			return pages, err
		}
		err = store.parseAllowedGroups(rawAllowedGroups, page)
		if err != nil {
			return pages, err
		}
		pages = append(pages, page)
	}
	return pages, rows.Err()
}

// Always returns nil as there's currently no cache
func (store *DefaultPageStore) Reload(id int) error {
	return nil
}

func (store *DefaultPageStore) Delete(id int) error {
	_, err := store.delete.Exec(id)
	return err
}
