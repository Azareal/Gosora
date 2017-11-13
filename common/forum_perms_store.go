package common

import (
	"database/sql"
	"encoding/json"
	"log"
	"sync"

	"../query_gen/lib"
)

var Fpstore ForumPermsStore

type ForumPermsStore interface {
	Init() error
	Get(fid int, gid int) (fperms *ForumPerms, err error)
	Reload(id int) error
	ReloadGroup(fid int, gid int) error
}

type ForumPermsCache interface {
}

type MemoryForumPermsStore struct {
	get             *sql.Stmt
	getByForum      *sql.Stmt
	getByForumGroup *sql.Stmt

	updateMutex sync.Mutex
}

func NewMemoryForumPermsStore() (*MemoryForumPermsStore, error) {
	acc := qgen.Builder.Accumulator()
	return &MemoryForumPermsStore{
		get:             acc.Select("forums_permissions").Columns("gid, fid, permissions").Orderby("gid ASC, fid ASC").Prepare(),
		getByForum:      acc.Select("forums_permissions").Columns("gid, permissions").Where("fid = ?").Orderby("gid ASC").Prepare(),
		getByForumGroup: acc.Select("forums_permissions").Columns("permissions").Where("fid = ? AND gid = ?").Prepare(),
	}, acc.FirstError()
}

func (fps *MemoryForumPermsStore) Init() error {
	fps.updateMutex.Lock()
	defer fps.updateMutex.Unlock()
	fids, err := Fstore.GetAllIDs()
	if err != nil {
		return err
	}
	debugDetail("fids: ", fids)

	rows, err := fps.get.Query()
	if err != nil {
		return err
	}
	defer rows.Close()
	debugLog("Adding the forum permissions")
	debugDetail("forumPerms[gid][fid]")

	// Temporarily store the forum perms in a map before transferring it to a much faster and thread-safe slice
	forumPerms = make(map[int]map[int]*ForumPerms)
	for rows.Next() {
		var gid, fid int
		var perms []byte
		err = rows.Scan(&gid, &fid, &perms)
		if err != nil {
			return err
		}

		pperms, err := fps.parseForumPerm(perms)
		if err != nil {
			return err
		}
		_, ok := forumPerms[gid]
		if !ok {
			forumPerms[gid] = make(map[int]*ForumPerms)
		}

		debugDetail("gid: ", gid)
		debugDetail("fid: ", fid)
		debugDetailf("perms: %+v\n", pperms)
		forumPerms[gid][fid] = pperms
	}

	return fps.cascadePermSetToGroups(forumPerms, fids)
}

func (fps *MemoryForumPermsStore) parseForumPerm(perms []byte) (pperms *ForumPerms, err error) {
	debugDetail("perms: ", string(perms))
	pperms = BlankForumPerms()
	err = json.Unmarshal(perms, &pperms)
	pperms.ExtData = make(map[string]bool)
	pperms.Overrides = true
	return pperms, err
}

// TODO: Need a more thread-safe way of doing this. Possibly with sync.Map?
func (fps *MemoryForumPermsStore) Reload(fid int) error {
	fps.updateMutex.Lock()
	defer fps.updateMutex.Unlock()
	debugLogf("Reloading the forum permissions for forum #%d", fid)
	fids, err := Fstore.GetAllIDs()
	if err != nil {
		return err
	}

	rows, err := fps.getByForum.Query(fid)
	if err != nil {
		return err
	}
	defer rows.Close()

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
		_, ok := forumPerms[gid]
		if !ok {
			forumPerms[gid] = make(map[int]*ForumPerms)
		}

		forumPerms[gid][fid] = pperms
	}

	return fps.cascadePermSetToGroups(forumPerms, fids)
}

func (fps *MemoryForumPermsStore) ReloadGroup(fid int, gid int) (err error) {
	fps.updateMutex.Lock()
	defer fps.updateMutex.Unlock()
	var perms []byte
	err = fps.getByForumGroup.QueryRow(fid, gid).Scan(&perms)
	if err != nil {
		return err
	}
	fperms, err := fps.parseForumPerm(perms)
	if err != nil {
		return err
	}
	group, err := Gstore.Get(gid)
	if err != nil {
		return err
	}
	// TODO: Refactor this
	group.Forums[fid] = fperms
	return nil
}

func (fps *MemoryForumPermsStore) cascadePermSetToGroups(forumPerms map[int]map[int]*ForumPerms, fids []int) error {
	groups, err := Gstore.GetAll()
	if err != nil {
		return err
	}

	for _, group := range groups {
		debugLogf("Updating the forum permissions for Group #%d", group.ID)
		group.Forums = []*ForumPerms{BlankForumPerms()}
		group.CanSee = []int{}
		fps.cascadePermSetToGroup(forumPerms, group, fids)

		if Dev.SuperDebug {
			log.Printf("group.CanSee (length %d): %+v \n", len(group.CanSee), group.CanSee)
			log.Printf("group.Forums (length %d): %+v\n", len(group.Forums), group.Forums)
		}
	}
	return nil
}

func (fps *MemoryForumPermsStore) cascadePermSetToGroup(forumPerms map[int]map[int]*ForumPerms, group *Group, fids []int) {
	for _, fid := range fids {
		debugDetailf("Forum #%+v\n", fid)
		forumPerm, ok := forumPerms[group.ID][fid]
		if ok {
			//log.Printf("Overriding permissions for forum #%d",fid)
			group.Forums = append(group.Forums, forumPerm)
		} else {
			//log.Printf("Inheriting from group defaults for forum #%d",fid)
			forumPerm = BlankForumPerms()
			group.Forums = append(group.Forums, forumPerm)
		}
		if forumPerm.Overrides {
			if forumPerm.ViewTopic {
				group.CanSee = append(group.CanSee, fid)
			}
		} else if group.Perms.ViewTopic {
			group.CanSee = append(group.CanSee, fid)
		}

		debugDetail("group.ID: ", group.ID)
		debugDetailf("forumPerm: %+v\n", forumPerm)
		debugDetail("group.CanSee: ", group.CanSee)
	}
}

// TODO: Add a hook here and have plugin_guilds use it
func (fps *MemoryForumPermsStore) Get(fid int, gid int) (fperms *ForumPerms, err error) {
	group, err := Gstore.Get(gid)
	if err != nil {
		return fperms, ErrNoRows
	}
	return group.Forums[fid], nil
}
