/* Under Heavy Construction */
package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"sort"
	"sync"
	"strconv"

	"github.com/Azareal/Gosora/query_gen"
)

var Groups GroupStore

// ? - We could fallback onto the database when an item can't be found in the cache?
type GroupStore interface {
	LoadGroups() error
	DirtyGet(id int) *Group
	Get(id int) (*Group, error)
	GetCopy(id int) (Group, error)
	Exists(id int) bool
	Create(name string, tag string, isAdmin bool, isMod bool, isBanned bool) (id int, err error)
	GetAll() ([]*Group, error)
	GetRange(lower int, higher int) ([]*Group, error)
	Reload(id int) error // ? - Should we move this to GroupCache? It might require us to do some unnecessary casting though
	Count() int
}

type GroupCache interface {
	CacheSet(group *Group) error
	SetCanSee(gid int, canSee []int) error
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
	return &MemoryGroupStore{
		groups:     make(map[int]*Group),
		groupCount: 0,
		getAll:     acc.Select("users_groups").Columns("gid, name, permissions, plugin_perms, is_mod, is_admin, is_banned, tag").Prepare(),
		get:        acc.Select("users_groups").Columns("name, permissions, plugin_perms, is_mod, is_admin, is_banned, tag").Where("gid = ?").Prepare(),
		count:      acc.Count("users_groups").Prepare(),
		userCount:  acc.Count("users").Where("group = ?").Prepare(),
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
	err = rows.Err()
	if err != nil {
		return err
	}
	s.groupCount = i

	DebugLog("Binding the Not Loggedin Group")
	GuestPerms = s.dirtyGetUnsafe(6).Perms
	TopicListThaw.Thaw()
	return nil
}

// TODO: Hit the database when the item isn't in memory
func (s *MemoryGroupStore) dirtyGetUnsafe(gid int) *Group {
	group, ok := s.groups[gid]
	if !ok {
		return &blankGroup
	}
	return group
}

// TODO: Hit the database when the item isn't in memory
func (s *MemoryGroupStore) DirtyGet(gid int) *Group {
	s.RLock()
	group, ok := s.groups[gid]
	s.RUnlock()
	if !ok {
		return &blankGroup
	}
	return group
}

// TODO: Hit the database when the item isn't in memory
func (s *MemoryGroupStore) Get(gid int) (*Group, error) {
	s.RLock()
	group, ok := s.groups[gid]
	s.RUnlock()
	if !ok {
		return nil, ErrNoRows
	}
	return group, nil
}

// TODO: Hit the database when the item isn't in memory
func (s *MemoryGroupStore) GetCopy(gid int) (Group, error) {
	s.RLock()
	group, ok := s.groups[gid]
	s.RUnlock()
	if !ok {
		return blankGroup, ErrNoRows
	}
	return *group, nil
}

func (s *MemoryGroupStore) Reload(id int) error {
	// TODO: Reload this data too
	g, err := s.Get(id)
	if err != nil {
		LogError(errors.New("can't get cansee data for group #" + strconv.Itoa(id)))
		return nil
	}
	canSee := g.CanSee
	
	g = &Group{ID: id, CanSee: canSee}
	err = s.get.QueryRow(id).Scan(&g.Name, &g.PermissionsText, &g.PluginPermsText, &g.IsMod, &g.IsAdmin, &g.IsBanned, &g.Tag)
	if err != nil {
		return err
	}

	err = s.initGroup(g)
	if err != nil {
		LogError(err)
		return nil
	}
	
	s.CacheSet(g)
	TopicListThaw.Thaw()
	return nil
}

func (s *MemoryGroupStore) initGroup(group *Group) error {
	err := json.Unmarshal(group.PermissionsText, &group.Perms)
	if err != nil {
		log.Printf("group: %+v\n", group)
		log.Print("bad group perms: ", group.PermissionsText)
		return err
	}
	DebugLogf(group.Name+": %+v\n", group.Perms)

	err = json.Unmarshal(group.PluginPermsText, &group.PluginPerms)
	if err != nil {
		log.Printf("group: %+v\n", group)
		log.Print("bad group plugin perms: ", group.PluginPermsText)
		return err
	}
	DebugLogf(group.Name+": %+v\n", group.PluginPerms)

	//group.Perms.ExtData = make(map[string]bool)
	// TODO: Can we optimise the bit where this cascades down to the user now?
	if group.IsAdmin || group.IsMod {
		group.IsBanned = false
	}

	err = s.userCount.QueryRow(group.ID).Scan(&group.UserCount)
	if err != sql.ErrNoRows {
		return err
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

func (s *MemoryGroupStore) CacheSet(group *Group) error {
	s.Lock()
	s.groups[group.ID] = group
	s.Unlock()
	return nil
}

// TODO: Hit the database when the item isn't in memory
func (s *MemoryGroupStore) Exists(gid int) bool {
	s.RLock()
	group, ok := s.groups[gid]
	s.RUnlock()
	return ok && group.Name != ""
}

// ? Allow two groups with the same name?
// TODO: Refactor this
func (s *MemoryGroupStore) Create(name string, tag string, isAdmin bool, isMod bool, isBanned bool) (gid int, err error) {
	permstr := "{}"
	tx, err := qgen.Builder.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	insertTx, err := qgen.Builder.SimpleInsertTx(tx, "users_groups", "name, tag, is_admin, is_mod, is_banned, permissions, plugin_perms", "?,?,?,?,?,?,'{}'")
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
	fdata, err := Forums.GetAll()
	if err != nil {
		return 0, err
	}

	presetSet := make(map[int]string)
	permSet := make(map[int]*ForumPerms)
	for _, forum := range fdata {
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

		permmap := PresetToPermmap(forum.Preset)
		permItem := permmap[thePreset]
		permItem.Overrides = true

		permSet[forum.ID] = permItem
		presetSet[forum.ID] = forum.Preset
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

	s.Lock()
	s.groups[gid] = &Group{gid, name, isMod, isAdmin, isBanned, tag, perms, []byte(permstr), pluginPerms, pluginPermsBytes, blankIntList, 0}
	s.groupCount++
	s.Unlock()

	TopicListThaw.Thaw()
	return gid, FPStore.ReloadAll()
	//return gid, TopicList.RebuildPermTree()
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
func (s *MemoryGroupStore) GetRange(lower int, higher int) (groups []*Group, err error) {
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
