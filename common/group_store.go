/* Under Heavy Construction */
package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"sort"
	"sync"

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
	Create(name string, tag string, isAdmin bool, isMod bool, isBanned bool) (int, error)
	GetAll() ([]*Group, error)
	GetRange(lower int, higher int) ([]*Group, error)
	Reload(id int) error // ? - Should we move this to GroupCache? It might require us to do some unnecessary casting though
	GlobalCount() int
}

type GroupCache interface {
	CacheSet(group *Group) error
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
func (mgs *MemoryGroupStore) LoadGroups() error {
	mgs.Lock()
	defer mgs.Unlock()
	mgs.groups[0] = &Group{ID: 0, Name: "Unknown"}

	rows, err := mgs.getAll.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 1
	for ; rows.Next(); i++ {
		group := &Group{ID: 0}
		err := rows.Scan(&group.ID, &group.Name, &group.PermissionsText, &group.PluginPermsText, &group.IsMod, &group.IsAdmin, &group.IsBanned, &group.Tag)
		if err != nil {
			return err
		}

		err = mgs.initGroup(group)
		if err != nil {
			return err
		}
		mgs.groups[group.ID] = group
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	mgs.groupCount = i

	DebugLog("Binding the Not Loggedin Group")
	GuestPerms = mgs.dirtyGetUnsafe(6).Perms
	return nil
}

// TODO: Hit the database when the item isn't in memory
func (mgs *MemoryGroupStore) dirtyGetUnsafe(gid int) *Group {
	group, ok := mgs.groups[gid]
	if !ok {
		return &blankGroup
	}
	return group
}

// TODO: Hit the database when the item isn't in memory
func (mgs *MemoryGroupStore) DirtyGet(gid int) *Group {
	mgs.RLock()
	group, ok := mgs.groups[gid]
	mgs.RUnlock()
	if !ok {
		return &blankGroup
	}
	return group
}

// TODO: Hit the database when the item isn't in memory
func (mgs *MemoryGroupStore) Get(gid int) (*Group, error) {
	mgs.RLock()
	group, ok := mgs.groups[gid]
	mgs.RUnlock()
	if !ok {
		return nil, ErrNoRows
	}
	return group, nil
}

// TODO: Hit the database when the item isn't in memory
func (mgs *MemoryGroupStore) GetCopy(gid int) (Group, error) {
	mgs.RLock()
	group, ok := mgs.groups[gid]
	mgs.RUnlock()
	if !ok {
		return blankGroup, ErrNoRows
	}
	return *group, nil
}

func (mgs *MemoryGroupStore) Reload(id int) error {
	var group = &Group{ID: id}
	err := mgs.get.QueryRow(id).Scan(&group.Name, &group.PermissionsText, &group.PluginPermsText, &group.IsMod, &group.IsAdmin, &group.IsBanned, &group.Tag)
	if err != nil {
		return err
	}

	err = mgs.initGroup(group)
	if err != nil {
		LogError(err)
	}
	mgs.CacheSet(group)

	err = RebuildGroupPermissions(id)
	if err != nil {
		LogError(err)
	}
	return nil
}

func (mgs *MemoryGroupStore) initGroup(group *Group) error {
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

	err = mgs.userCount.QueryRow(group.ID).Scan(&group.UserCount)
	if err != sql.ErrNoRows {
		return err
	}
	return nil
}

func (mgs *MemoryGroupStore) CacheSet(group *Group) error {
	mgs.Lock()
	mgs.groups[group.ID] = group
	mgs.Unlock()
	return nil
}

// TODO: Hit the database when the item isn't in memory
func (mgs *MemoryGroupStore) Exists(gid int) bool {
	mgs.RLock()
	group, ok := mgs.groups[gid]
	mgs.RUnlock()
	return ok && group.Name != ""
}

// ? Allow two groups with the same name?
// TODO: Refactor this
func (mgs *MemoryGroupStore) Create(name string, tag string, isAdmin bool, isMod bool, isBanned bool) (gid int, err error) {
	var permstr = "{}"
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

	var perms = BlankPerms
	var blankIntList []int
	var pluginPerms = make(map[string]bool)
	var pluginPermsBytes = []byte("{}")
	GetHookTable().Vhook("create_group_preappend", &pluginPerms, &pluginPermsBytes)

	// Generate the forum permissions based on the presets...
	fdata, err := Forums.GetAll()
	if err != nil {
		return 0, err
	}

	var presetSet = make(map[int]string)
	var permSet = make(map[int]*ForumPerms)
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

	mgs.Lock()
	mgs.groups[gid] = &Group{gid, name, isMod, isAdmin, isBanned, tag, perms, []byte(permstr), pluginPerms, pluginPermsBytes, blankIntList, 0}
	mgs.groupCount++
	mgs.Unlock()

	err = FPStore.ReloadAll()
	if err != nil {
		return gid, err
	}
	err = TopicList.RebuildPermTree()
	if err != nil {
		return gid, err
	}

	return gid, nil
}

func (mgs *MemoryGroupStore) GetAll() (results []*Group, err error) {
	var i int
	mgs.RLock()
	results = make([]*Group, len(mgs.groups))
	for _, group := range mgs.groups {
		results[i] = group
		i++
	}
	mgs.RUnlock()
	sort.Sort(SortGroup(results))
	return results, nil
}

func (mgs *MemoryGroupStore) GetAllMap() (map[int]*Group, error) {
	mgs.RLock()
	defer mgs.RUnlock()
	return mgs.groups, nil
}

// ? - Set the lower and higher numbers to 0 to remove the bounds
// TODO: Might be a little slow right now, maybe we can cache the groups in a slice or break the map up into chunks
func (mgs *MemoryGroupStore) GetRange(lower int, higher int) (groups []*Group, err error) {
	if lower == 0 && higher == 0 {
		return mgs.GetAll()
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

	mgs.RLock()
	for gid, group := range mgs.groups {
		if gid >= lower && (gid <= higher || higher == 0) {
			groups = append(groups, group)
		}
	}
	mgs.RUnlock()
	sort.Sort(SortGroup(groups))

	return groups, nil
}

func (mgs *MemoryGroupStore) Length() int {
	mgs.RLock()
	defer mgs.RUnlock()
	return mgs.groupCount
}

func (mgs *MemoryGroupStore) GlobalCount() (count int) {
	err := mgs.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}
