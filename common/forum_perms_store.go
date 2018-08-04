package common

import (
	"database/sql"
	"encoding/json"
	"sync"

	"../query_gen/lib"
)

var FPStore ForumPermsStore

type ForumPermsStore interface {
	Init() error
	GetAllMap() (bigMap map[int]map[int]*ForumPerms)
	Get(fid int, gid int) (fperms *ForumPerms, err error)
	GetCopy(fid int, gid int) (fperms ForumPerms, err error)
	ReloadAll() error
	Reload(id int) error
}

type ForumPermsCache interface {
}

type MemoryForumPermsStore struct {
	getByForum      *sql.Stmt
	getByForumGroup *sql.Stmt

	evenForums map[int]map[int]*ForumPerms
	oddForums  map[int]map[int]*ForumPerms // [fid][gid]*ForumPerms
	evenLock   sync.RWMutex
	oddLock    sync.RWMutex
}

func NewMemoryForumPermsStore() (*MemoryForumPermsStore, error) {
	acc := qgen.NewAcc()
	return &MemoryForumPermsStore{
		getByForum:      acc.Select("forums_permissions").Columns("gid, permissions").Where("fid = ?").Orderby("gid ASC").Prepare(),
		getByForumGroup: acc.Select("forums_permissions").Columns("permissions").Where("fid = ? AND gid = ?").Prepare(),

		evenForums: make(map[int]map[int]*ForumPerms),
		oddForums:  make(map[int]map[int]*ForumPerms),
	}, acc.FirstError()
}

func (fps *MemoryForumPermsStore) Init() error {
	DebugLog("Initialising the forum perms store")
	return fps.ReloadAll()
}

func (fps *MemoryForumPermsStore) ReloadAll() error {
	DebugLog("Reloading the forum perms")
	fids, err := Forums.GetAllIDs()
	if err != nil {
		return err
	}
	for _, fid := range fids {
		err := fps.Reload(fid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fps *MemoryForumPermsStore) parseForumPerm(perms []byte) (pperms *ForumPerms, err error) {
	DebugDetail("perms: ", string(perms))
	pperms = BlankForumPerms()
	err = json.Unmarshal(perms, &pperms)
	pperms.ExtData = make(map[string]bool)
	pperms.Overrides = true
	return pperms, err
}

// TODO: Need a more thread-safe way of doing this. Possibly with sync.Map?
func (fps *MemoryForumPermsStore) Reload(fid int) error {
	DebugLogf("Reloading the forum permissions for forum #%d", fid)
	rows, err := fps.getByForum.Query(fid)
	if err != nil {
		return err
	}
	defer rows.Close()

	var forumPerms = make(map[int]*ForumPerms)
	for rows.Next() {
		var gid int
		var perms []byte
		err := rows.Scan(&gid, &perms)
		if err != nil {
			return err
		}

		pperms, err := fps.parseForumPerm(perms)
		if err != nil {
			return err
		}
		forumPerms[gid] = pperms
	}
	DebugLogf("forumPerms: %+v\n", forumPerms)
	if fid%2 == 0 {
		fps.evenLock.Lock()
		fps.evenForums[fid] = forumPerms
		fps.evenLock.Unlock()
	} else {
		fps.oddLock.Lock()
		fps.oddForums[fid] = forumPerms
		fps.oddLock.Unlock()
	}

	groups, err := Groups.GetAll()
	if err != nil {
		return err
	}
	fids, err := Forums.GetAllIDs()
	if err != nil {
		return err
	}

	for _, group := range groups {
		DebugLogf("Updating the forum permissions for Group #%d", group.ID)
		group.CanSee = []int{}
		for _, fid := range fids {
			DebugDetailf("Forum #%+v\n", fid)
			var forumPerms map[int]*ForumPerms
			var ok bool
			if fid%2 == 0 {
				fps.evenLock.RLock()
				forumPerms, ok = fps.evenForums[fid]
				fps.evenLock.RUnlock()
			} else {
				fps.oddLock.RLock()
				forumPerms, ok = fps.oddForums[fid]
				fps.oddLock.RUnlock()
			}

			var forumPerm *ForumPerms
			if !ok {
				continue
			}
			forumPerm, ok = forumPerms[group.ID]
			if !ok {
				continue
			}

			if forumPerm.Overrides {
				if forumPerm.ViewTopic {
					group.CanSee = append(group.CanSee, fid)
				}
			} else if group.Perms.ViewTopic {
				group.CanSee = append(group.CanSee, fid)
			}
			DebugDetail("group.ID: ", group.ID)
			DebugDetailf("forumPerm: %+v\n", forumPerm)
			DebugDetail("group.CanSee: ", group.CanSee)
		}
		DebugDetailf("group.CanSee (length %d): %+v \n", len(group.CanSee), group.CanSee)
	}
	return nil
}

// ! Throughput on this might be bad due to the excessive locking
func (fps *MemoryForumPermsStore) GetAllMap() (bigMap map[int]map[int]*ForumPerms) {
	bigMap = make(map[int]map[int]*ForumPerms)
	fps.evenLock.RLock()
	for fid, subMap := range fps.evenForums {
		bigMap[fid] = subMap
	}
	fps.evenLock.RUnlock()
	fps.oddLock.RLock()
	for fid, subMap := range fps.oddForums {
		bigMap[fid] = subMap
	}
	fps.oddLock.RUnlock()
	return bigMap
}

// TODO: Add a hook here and have plugin_guilds use it
// TODO: Check if the forum exists?
// TODO: Fix the races
func (fps *MemoryForumPermsStore) Get(fid int, gid int) (fperms *ForumPerms, err error) {
	var fmap map[int]*ForumPerms
	var ok bool
	if fid%2 == 0 {
		fps.evenLock.RLock()
		fmap, ok = fps.evenForums[fid]
		fps.evenLock.RUnlock()
	} else {
		fps.oddLock.RLock()
		fmap, ok = fps.oddForums[fid]
		fps.oddLock.RUnlock()
	}
	if !ok {
		return fperms, ErrNoRows
	}

	fperms, ok = fmap[gid]
	if !ok {
		return fperms, ErrNoRows
	}
	return fperms, nil
}

// TODO: Check if the forum exists?
// TODO: Fix the races
func (fps *MemoryForumPermsStore) GetCopy(fid int, gid int) (fperms ForumPerms, err error) {
	fPermsPtr, err := fps.Get(fid, gid)
	if err != nil {
		return fperms, err
	}
	return *fPermsPtr, nil
}
