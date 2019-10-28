package common

import (
	"database/sql"
	"strconv"
	"sync/atomic"

	"github.com/Azareal/Gosora/query_gen"
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
func (s *DefaultMenuStore) GetAllMap() (out map[int]*MenuListHolder) {
	out = make(map[int]*MenuListHolder)
	for mid, atom := range s.menus {
		out[mid] = atom.Load().(*MenuListHolder)
	}
	return out
}

func (s *DefaultMenuStore) Get(mid int) (*MenuListHolder, error) {
	aStore, ok := s.menus[mid]
	if ok {
		return aStore.Load().(*MenuListHolder), nil
	}
	return nil, ErrNoRows
}

func (s *DefaultMenuStore) Items(mid int) (mlist MenuItemList, err error) {
	err = qgen.NewAcc().Select("menu_items").Columns("miid,name,htmlID,cssClass,position,path,aria,tooltip,order,tmplName,guestOnly,memberOnly,staffOnly,adminOnly").Where("mid = " + strconv.Itoa(mid)).Orderby("order ASC").Each(func(rows *sql.Rows) error {
		i := MenuItem{MenuID: mid}
		err := rows.Scan(&i.ID, &i.Name, &i.HTMLID, &i.CSSClass, &i.Position, &i.Path, &i.Aria, &i.Tooltip, &i.Order, &i.TmplName, &i.GuestOnly, &i.MemberOnly, &i.SuperModOnly, &i.AdminOnly)
		if err != nil {
			return err
		}
		s.itemStore.Add(i)
		mlist = append(mlist, i)
		return nil
	})
	return mlist, err
}

func (s *DefaultMenuStore) Load(mid int) error {
	mlist, err := s.Items(mid)
	if err != nil {
		return err
	}
	hold := &MenuListHolder{mid, mlist, make(map[int]menuTmpl)}
	err = hold.Preparse()
	if err != nil {
		return err
	}

	aStore := &atomic.Value{}
	aStore.Store(hold)
	s.menus[mid] = aStore
	return nil
}

func (s *DefaultMenuStore) ItemStore() *DefaultMenuItemStore {
	return s.itemStore
}
