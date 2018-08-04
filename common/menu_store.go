package common

import (
	"database/sql"
	"strconv"
	"sync/atomic"

	"../query_gen/lib"
)

var Menus *DefaultMenuStore

type DefaultMenuStore struct {
	menus     map[int]*atomic.Value
	itemStore *DefaultMenuItemStore
}

func NewDefaultMenuStore() *DefaultMenuStore {
	return &DefaultMenuStore{
		make(map[int]*atomic.Value),
		NewDefaultMenuItemStore(),
	}
}

// TODO: Add actual support for multiple menus
func (store *DefaultMenuStore) GetAllMap() (out map[int]*MenuListHolder) {
	out = make(map[int]*MenuListHolder)
	for mid, atom := range store.menus {
		out[mid] = atom.Load().(*MenuListHolder)
	}
	return out
}

func (store *DefaultMenuStore) Get(mid int) (*MenuListHolder, error) {
	aStore, ok := store.menus[mid]
	if ok {
		return aStore.Load().(*MenuListHolder), nil
	}
	return nil, ErrNoRows
}

func (store *DefaultMenuStore) Items(mid int) (mlist MenuItemList, err error) {
	err = qgen.NewAcc().Select("menu_items").Columns("miid, name, htmlID, cssClass, position, path, aria, tooltip, order, tmplName, guestOnly, memberOnly, staffOnly, adminOnly").Where("mid = " + strconv.Itoa(mid)).Orderby("order ASC").Each(func(rows *sql.Rows) error {
		var mitem = MenuItem{MenuID: mid}
		err := rows.Scan(&mitem.ID, &mitem.Name, &mitem.HTMLID, &mitem.CSSClass, &mitem.Position, &mitem.Path, &mitem.Aria, &mitem.Tooltip, &mitem.Order, &mitem.TmplName, &mitem.GuestOnly, &mitem.MemberOnly, &mitem.SuperModOnly, &mitem.AdminOnly)
		if err != nil {
			return err
		}
		store.itemStore.Add(mitem)
		mlist = append(mlist, mitem)
		return nil
	})
	return mlist, err
}

func (store *DefaultMenuStore) Load(mid int) error {
	mlist, err := store.Items(mid)
	if err != nil {
		return err
	}
	hold := &MenuListHolder{mid, mlist, make(map[int]menuTmpl)}
	err = hold.Preparse()
	if err != nil {
		return err
	}

	var aStore = &atomic.Value{}
	aStore.Store(hold)
	store.menus[mid] = aStore
	return nil
}

func (store *DefaultMenuStore) ItemStore() *DefaultMenuItemStore {
	return store.itemStore
}
