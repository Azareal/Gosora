package common

import (
	"database/sql"
	"encoding/json"
	"sync"

	"github.com/Azareal/Gosora/query_gen"
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

func (s *MemoryForumPermsStore) Init() error {
	DebugLog("Initialising the forum perms store")
	return s.ReloadAll()
}

func (s *MemoryForumPermsStore) ReloadAll() error {
	DebugLog("Reloading the forum perms")
	fids, err := Forums.GetAllIDs()
	if err != nil {
		return err
	}
	for _, fid := range fids {
		err := s.Reload(fid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryForumPermsStore) parseForumPerm(perms []byte) (pperms *ForumPerms, err error) {
	DebugDetail("perms: ", string(perms))
	pperms = BlankForumPerms()
	err = json.Unmarshal(perms, &pperms)
	pperms.ExtData = make(map[string]bool)
	pperms.Overrides = true
	return pperms, err
}

// TODO: Need a more thread-safe way of doing this. Possibly with sync.Map?
func (s *MemoryForumPermsStore) Reload(fid int) error {
	DebugLogf("Reloading the forum permissions for forum #%d", fid)
	rows, err := s.getByForum.Query(fid)
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

		DebugLog("gid: ", gid)
		DebugLogf("perms: %+v\n", perms)
		pperms, err := s.parseForumPerm(perms)
		if err != nil {
			return err
		}
		DebugLogf("pperms: %+v\n", pperms)
		forumPerms[gid] = pperms
	}
	DebugLogf("forumPerms: %+v\n", forumPerms)
	if fid%2 == 0 {
		s.evenLock.Lock()
		s.evenForums[fid] = forumPerms
		s.evenLock.Unlock()
	} else {
		s.oddLock.Lock()
		s.oddForums[fid] = forumPerms
		s.oddLock.Unlock()
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
				s.evenLock.RLock()
				forumPerms, ok = s.evenForums[fid]
				s.evenLock.RUnlock()
			} else {
				s.oddLock.RLock()
				forumPerms, ok = s.oddForums[fid]
				s.oddLock.RUnlock()
			}

			var forumPerm *ForumPerms
			if !ok {
				continue
			}
			forumPerm, ok = forumPerms[group.ID]
			if !ok {
				if group.Perms.ViewTopic {
					group.CanSee = append(group.CanSee, fid)
				}
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
	TopicListThaw.Thaw()
	return nil
}

// ! Throughput on this might be bad due to the excessive locking
func (s *MemoryForumPermsStore) GetAllMap() (bigMap map[int]map[int]*ForumPerms) {
	bigMap = make(map[int]map[int]*ForumPerms)
	s.evenLock.RLock()
	for fid, subMap := range s.evenForums {
		bigMap[fid] = subMap
	}
	s.evenLock.RUnlock()
	s.oddLock.RLock()
	for fid, subMap := range s.oddForums {
		bigMap[fid] = subMap
	}
	s.oddLock.RUnlock()
	return bigMap
}

// TODO: Add a hook here and have plugin_guilds use it
// TODO: Check if the forum exists?
// TODO: Fix the races
func (s *MemoryForumPermsStore) Get(fid int, gid int) (fperms *ForumPerms, err error) {
	var fmap map[int]*ForumPerms
	var ok bool
	if fid%2 == 0 {
		s.evenLock.RLock()
		fmap, ok = s.evenForums[fid]
		s.evenLock.RUnlock()
	} else {
		s.oddLock.RLock()
		fmap, ok = s.oddForums[fid]
		s.oddLock.RUnlock()
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
func (s *MemoryForumPermsStore) GetCopy(fid int, gid int) (fperms ForumPerms, err error) {
	fPermsPtr, err := s.Get(fid, gid)
	if err != nil {
		return fperms, err
	}
	return *fPermsPtr, nil
}
