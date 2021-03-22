package common

import (
	"database/sql"
	"encoding/json"
	"sync"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var FPStore ForumPermsStore

type ForumPermsStore interface {
	Init() error
	GetAllMap() (bigMap map[int]map[int]*ForumPerms)
	Get(fid, gid int) (fp *ForumPerms, err error)
	GetCopy(fid, gid int) (fp ForumPerms, err error)
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
	fp := "forums_permissions"
	return &MemoryForumPermsStore{
		getByForum:      acc.Select(fp).Columns("gid,permissions").Where("fid=?").Orderby("gid ASC").Prepare(),
		getByForumGroup: acc.Select(fp).Columns("permissions").Where("fid=? AND gid=?").Prepare(),

		evenForums: make(map[int]map[int]*ForumPerms),
		oddForums:  make(map[int]map[int]*ForumPerms),
	}, acc.FirstError()
}

func (s *MemoryForumPermsStore) Init() error {
	DebugLog("Initialising the forum perms store")
	return s.ReloadAll()
}

// TODO: Optimise this?
func (s *MemoryForumPermsStore) ReloadAll() error {
	DebugLog("Reloading the forum perms")
	fids, err := Forums.GetAllIDs()
	if err != nil {
		return err
	}
	for _, fid := range fids {
		if e := s.reload(fid); e != nil {
			return e
		}
	}
	if e := s.recalcCanSeeAll(); e != nil {
		return e
	}
	TopicListThaw.Thaw()
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

func (s *MemoryForumPermsStore) Reload(fid int) error {
	e := s.reload(fid)
	if e != nil {
		return e
	}
	if e = s.recalcCanSeeAll(); e != nil {
		return e
	}
	TopicListThaw.Thaw()
	return nil
}

// TODO: Need a more thread-safe way of doing this. Possibly with sync.Map?
func (s *MemoryForumPermsStore) reload(fid int) error {
	DebugLogf("Reloading the forum permissions for forum #%d", fid)
	rows, err := s.getByForum.Query(fid)
	if err != nil {
		return err
	}
	defer rows.Close()

	forumPerms := make(map[int]*ForumPerms)
	for rows.Next() {
		var gid int
		var perms []byte
		err := rows.Scan(&gid, &perms)
		if err != nil {
			return err
		}

		DebugLog("gid:", gid)
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
	return nil
}

func (s *MemoryForumPermsStore) recalcCanSeeAll() error {
	groups, err := Groups.GetAll()
	if err != nil {
		return err
	}
	fids, err := Forums.GetAllIDs()
	if err != nil {
		return err
	}

	gc, ok := Groups.(GroupCache)
	if !ok {
		TopicListThaw.Thaw()
		return nil
	}

	// A separate loop to avoid contending on the odd-even locks as much
	fForumPerms := make(map[int]map[int]*ForumPerms)
	for _, fid := range fids {
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
		if ok {
			fForumPerms[fid] = forumPerms
		}
	}

	// TODO: Can we recalculate CanSee without calculating every other forum?
	for _, g := range groups {
		DebugLogf("Updating the forum permissions for Group #%d", g.ID)
		canSee := []int{}
		for _, fid := range fids {
			DebugDetailf("Forum #%+v\n", fid)
			forumPerms, ok := fForumPerms[fid]
			if !ok {
				continue
			}
			fp, ok := forumPerms[g.ID]
			if !ok {
				if g.Perms.ViewTopic {
					canSee = append(canSee, fid)
				}
				continue
			}

			if fp.Overrides {
				if fp.ViewTopic {
					canSee = append(canSee, fid)
				}
			} else if g.Perms.ViewTopic {
				canSee = append(canSee, fid)
			}
			//DebugDetail("g.ID: ", g.ID)
			DebugDetailf("forumPerm: %+v\n", fp)
			DebugDetail("canSee: ", canSee)
		}
		DebugDetailf("canSee (length %d): %+v \n", len(canSee), canSee)
		gc.SetCanSee(g.ID, canSee)
	}

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
// TODO: Return BlankForumPerms() when the forum permission set doesn't exist?
func (s *MemoryForumPermsStore) Get(fid, gid int) (fp *ForumPerms, err error) {
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
		return fp, ErrNoRows
	}

	fp, ok = fmap[gid]
	if !ok {
		return fp, ErrNoRows
	}
	return fp, nil
}

// TODO: Check if the forum exists?
// TODO: Fix the races
func (s *MemoryForumPermsStore) GetCopy(fid, gid int) (fp ForumPerms, err error) {
	fPermsPtr, err := s.Get(fid, gid)
	if err != nil {
		return fp, err
	}
	return *fPermsPtr, nil
}
