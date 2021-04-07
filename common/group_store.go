/* Under Heavy Construction */
package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strconv"
	"sync"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Groups GroupStore

// ? - We could fallback onto the database when an item can't be found in the cache?
type GroupStore interface {
	LoadGroups() error
	DirtyGet(id int) *Group
	Get(id int) (*Group, error)
	GetCopy(id int) (Group, error)
	Exists(id int) bool
	Create(name, tag string, isAdmin, isMod, isBanned bool) (id int, err error)
	GetAll() ([]*Group, error)
	GetRange(lower, higher int) ([]*Group, error)
	Reload(id int) error // ? - Should we move this to GroupCache? It might require us to do some unnecessary casting though
	Count() int
}

type GroupCache interface {
	CacheSet(g *Group) error
	SetCanSee(gid int, canSee []int) error
	CacheAdd(g *Group) error
	Length() int
}

type MemoryGroupStore struct {
	groups     map[int]*Group // TODO: Use a sync.Map instead of a map?
	groupCount int
	getAll     *sql.Stmt
	get        *sql.Stmt
	count      *sql.Stmt
	userCount  *sql.Stmt

	sync.RWMutex
}

func NewMemoryGroupStore() (*MemoryGroupStore, error) {
	acc := qgen.NewAcc()
	ug := "users_groups"
	return &MemoryGroupStore{
		groups:     make(map[int]*Group),
		groupCount: 0,
		getAll:     acc.Select(ug).Columns("gid,name,permissions,plugin_perms,is_mod,is_admin,is_banned,tag").Prepare(),
		get:        acc.Select(ug).Columns("name,permissions,plugin_perms,is_mod,is_admin,is_banned,tag").Where("gid=?").Prepare(),
		count:      acc.Count(ug).Prepare(),
		userCount:  acc.Count("users").Where("group=?").Prepare(),
	}, acc.FirstError()
}

// TODO: Move this query from the global stmt store into this store
func (s *MemoryGroupStore) LoadGroups() error {
	s.Lock()
	defer s.Unlock()
	s.groups[0] = &Group{ID: 0, Name: "Unknown"}

	rows, err := s.getAll.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 1
	for ; rows.Next(); i++ {
		g := &Group{ID: 0}
		err := rows.Scan(&g.ID, &g.Name, &g.PermissionsText, &g.PluginPermsText, &g.IsMod, &g.IsAdmin, &g.IsBanned, &g.Tag)
		if err != nil {
			return err
		}

		err = s.initGroup(g)
		if err != nil {
			return err
		}
		s.groups[g.ID] = g
	}
	if err = rows.Err(); err != nil {
		return err
	}
	s.groupCount = i

	DebugLog("Binding the Not Loggedin Group")
	GuestPerms = s.dirtyGetUnsafe(6).Perms // ! Race?
	TopicListThaw.Thaw()
	return nil
}

// TODO: Hit the database when the item isn't in memory
func (s *MemoryGroupStore) dirtyGetUnsafe(id int) *Group {
	group, ok := s.groups[id]
	if !ok {
		return &blankGroup
	}
	return group
}

// TODO: Hit the database when the item isn't in memory
func (s *MemoryGroupStore) DirtyGet(id int) *Group {
	s.RLock()
	group, ok := s.groups[id]
	s.RUnlock()
	if !ok {
		return &blankGroup
	}
	return group
}

// TODO: Hit the database when the item isn't in memory
func (s *MemoryGroupStore) Get(id int) (*Group, error) {
	s.RLock()
	group, ok := s.groups[id]
	s.RUnlock()
	if !ok {
		return nil, ErrNoRows
	}
	return group, nil
}

// TODO: Hit the database when the item isn't in memory
func (s *MemoryGroupStore) GetCopy(id int) (Group, error) {
	s.RLock()
	group, ok := s.groups[id]
	s.RUnlock()
	if !ok {
		return blankGroup, ErrNoRows
	}
	return *group, nil
}

func (s *MemoryGroupStore) Reload(id int) error {
	// TODO: Reload this data too
	g, e := s.Get(id)
	if e != nil {
		LogError(errors.New("can't get cansee data for group #" + strconv.Itoa(id)))
		return nil
	}
	canSee := g.CanSee

	g = &Group{ID: id, CanSee: canSee}
	e = s.get.QueryRow(id).Scan(&g.Name, &g.PermissionsText, &g.PluginPermsText, &g.IsMod, &g.IsAdmin, &g.IsBanned, &g.Tag)
	if e != nil {
		return e
	}
	if e = s.initGroup(g); e != nil {
		LogError(e)
		return nil
	}

	s.CacheSet(g)
	TopicListThaw.Thaw()
	return nil
}

func (s *MemoryGroupStore) initGroup(g *Group) error {
	e := json.Unmarshal(g.PermissionsText, &g.Perms)
	if e != nil {
		log.Printf("g: %+v\n", g)
		log.Print("bad group perms: ", g.PermissionsText)
		return e
	}
	DebugLogf(g.Name+": %+v\n", g.Perms)

	e = json.Unmarshal(g.PluginPermsText, &g.PluginPerms)
	if e != nil {
		log.Printf("g: %+v\n", g)
		log.Print("bad group plugin perms: ", g.PluginPermsText)
		return e
	}
	DebugLogf(g.Name+": %+v\n", g.PluginPerms)

	//group.Perms.ExtData = make(map[string]bool)
	// TODO: Can we optimise the bit where this cascades down to the user now?
	if g.IsAdmin || g.IsMod {
		g.IsBanned = false
	}

	e = s.userCount.QueryRow(g.ID).Scan(&g.UserCount)
	if e != sql.ErrNoRows {
		return e
	}
	return nil
}

func (s *MemoryGroupStore) SetCanSee(gid int, canSee []int) error {
	s.Lock()
	group, ok := s.groups[gid]
	if !ok {
		s.Unlock()
		return nil
	}
	ngroup := &Group{}
	*ngroup = *group
	ngroup.CanSee = canSee
	s.groups[group.ID] = ngroup
	s.Unlock()
	return nil
}

func (s *MemoryGroupStore) CacheSet(g *Group) error {
	s.Lock()
	s.groups[g.ID] = g
	s.Unlock()
	return nil
}

// TODO: Hit the database when the item isn't in memory
func (s *MemoryGroupStore) Exists(id int) bool {
	s.RLock()
	group, ok := s.groups[id]
	s.RUnlock()
	return ok && group.Name != ""
}

// ? Allow two groups with the same name?
// TODO: Refactor this
func (s *MemoryGroupStore) Create(name, tag string, isAdmin, isMod, isBanned bool) (gid int, err error) {
	permstr := "{}"
	tx, err := qgen.Builder.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	insertTx, err := qgen.Builder.SimpleInsertTx(tx, "users_groups", "name,tag,is_admin,is_mod,is_banned,permissions,plugin_perms", "?,?,?,?,?,?,'{}'")
	if err != nil {
		return 0, err
	}
	res, err := insertTx.Exec(name, tag, isAdmin, isMod, isBanned, permstr)
	if err != nil {
		return 0, err
	}
	gid64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	gid = int(gid64)

	perms := BlankPerms
	blankIntList := []int{}
	pluginPerms := make(map[string]bool)
	pluginPermsBytes := []byte("{}")
	GetHookTable().Vhook("create_group_preappend", &pluginPerms, &pluginPermsBytes)

	// Generate the forum permissions based on the presets...
	forums, err := Forums.GetAll()
	if err != nil {
		return 0, err
	}

	presetSet := make(map[int]string)
	permSet := make(map[int]*ForumPerms)
	for _, f := range forums {
		var thePreset string
		switch {
		case isAdmin:
			thePreset = "admins"
		case isMod:
			thePreset = "staff"
		case isBanned:
			thePreset = "banned"
		default:
			thePreset = "members"
		}

		permmap := PresetToPermmap(f.Preset)
		permItem := permmap[thePreset]
		permItem.Overrides = true

		permSet[f.ID] = permItem
		presetSet[f.ID] = f.Preset
	}

	err = ReplaceForumPermsForGroupTx(tx, gid, presetSet, permSet)
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	// TODO: Can we optimise the bit where this cascades down to the user now?
	if isAdmin || isMod {
		isBanned = false
	}

	s.CacheAdd(&Group{gid, name, isMod, isAdmin, isBanned, tag, perms, []byte(permstr), pluginPerms, pluginPermsBytes, blankIntList, 0})

	TopicListThaw.Thaw()
	return gid, FPStore.ReloadAll()
	//return gid, TopicList.RebuildPermTree()
}

func (s *MemoryGroupStore) CacheAdd(g *Group) error {
	s.Lock()
	s.groups[g.ID] = g
	s.groupCount++
	s.Unlock()
	return nil
}

func (s *MemoryGroupStore) GetAll() (results []*Group, err error) {
	var i int
	s.RLock()
	results = make([]*Group, len(s.groups))
	for _, group := range s.groups {
		results[i] = group
		i++
	}
	s.RUnlock()
	sort.Sort(SortGroup(results))
	return results, nil
}

func (s *MemoryGroupStore) GetAllMap() (map[int]*Group, error) {
	s.RLock()
	defer s.RUnlock()
	return s.groups, nil
}

// ? - Set the lower and higher numbers to 0 to remove the bounds
// TODO: Might be a little slow right now, maybe we can cache the groups in a slice or break the map up into chunks
func (s *MemoryGroupStore) GetRange(lower, higher int) (groups []*Group, err error) {
	if lower == 0 && higher == 0 {
		return s.GetAll()
	}

	// TODO: Simplify these four conditionals into two
	if lower == 0 {
		if higher < 0 {
			return nil, errors.New("higher may not be lower than 0")
		}
	} else if higher == 0 {
		if lower < 0 {
			return nil, errors.New("lower may not be lower than 0")
		}
	}

	s.RLock()
	for gid, group := range s.groups {
		if gid >= lower && (gid <= higher || higher == 0) {
			groups = append(groups, group)
		}
	}
	s.RUnlock()
	sort.Sort(SortGroup(groups))

	return groups, nil
}

func (s *MemoryGroupStore) Length() int {
	s.RLock()
	defer s.RUnlock()
	return s.groupCount
}

func (s *MemoryGroupStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}
