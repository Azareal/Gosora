package main

import (
	"encoding/json"
	"log"
)

var fpstore *ForumPermsStore

type ForumPermsStore struct {
}

func NewForumPermsStore() *ForumPermsStore {
	return &ForumPermsStore{}
}

func (fps *ForumPermsStore) Init() error {
	fids, err := fstore.GetAllIDs()
	if err != nil {
		return err
	}
	if dev.SuperDebug {
		log.Print("fids: ", fids)
	}

	rows, err := stmts.getForumsPermissions.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	if dev.DebugMode {
		log.Print("Adding the forum permissions")
		if dev.SuperDebug {
			log.Print("forumPerms[gid][fid]")
		}
	}

	// Temporarily store the forum perms in a map before transferring it to a much faster and thread-safe slice
	forumPerms = make(map[int]map[int]ForumPerms)
	for rows.Next() {
		var gid, fid int
		var perms []byte
		var pperms ForumPerms
		err = rows.Scan(&gid, &fid, &perms)
		if err != nil {
			return err
		}

		if dev.SuperDebug {
			log.Print("perms: ", string(perms))
		}
		err = json.Unmarshal(perms, &pperms)
		if err != nil {
			return err
		}
		pperms.ExtData = make(map[string]bool)
		pperms.Overrides = true
		_, ok := forumPerms[gid]
		if !ok {
			forumPerms[gid] = make(map[int]ForumPerms)
		}

		if dev.SuperDebug {
			log.Print("gid: ", gid)
			log.Print("fid: ", fid)
			log.Printf("perms: %+v\n", pperms)
		}
		forumPerms[gid][fid] = pperms
	}

	return fps.cascadePermSetToGroups(forumPerms, fids)
}

// TODO: Need a more thread-safe way of doing this. Possibly with sync.Map?
func (fps *ForumPermsStore) Reload(fid int) error {
	if dev.DebugMode {
		log.Printf("Reloading the forum permissions for forum #%d", fid)
	}
	fids, err := fstore.GetAllIDs()
	if err != nil {
		return err
	}

	rows, err := db.Query("select gid, permissions from forums_permissions where fid = ? order by gid asc", fid)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var gid int
		var perms []byte
		var pperms ForumPerms
		err := rows.Scan(&gid, &perms)
		if err != nil {
			return err
		}
		err = json.Unmarshal(perms, &pperms)
		if err != nil {
			return err
		}
		pperms.ExtData = make(map[string]bool)
		pperms.Overrides = true
		_, ok := forumPerms[gid]
		if !ok {
			forumPerms[gid] = make(map[int]ForumPerms)
		}
		forumPerms[gid][fid] = pperms
	}

	return fps.cascadePermSetToGroups(forumPerms, fids)
}

func (fps *ForumPermsStore) cascadePermSetToGroups(forumPerms map[int]map[int]ForumPerms, fids []int) error {
	groups, err := gstore.GetAll()
	if err != nil {
		return err
	}

	for _, group := range groups {
		if dev.DebugMode {
			log.Printf("Updating the forum permissions for Group #%d", group.ID)
		}
		group.Forums = []ForumPerms{BlankForumPerms}
		group.CanSee = []int{}
		fps.cascadePermSetToGroup(forumPerms, group, fids)

		if dev.SuperDebug {
			log.Printf("group.CanSee (length %d): %+v \n", len(group.CanSee), group.CanSee)
			log.Printf("group.Forums (length %d): %+v\n", len(group.Forums), group.Forums)
		}
	}
	return nil
}

func (fps *ForumPermsStore) cascadePermSetToGroup(forumPerms map[int]map[int]ForumPerms, group *Group, fids []int) {
	for _, fid := range fids {
		if dev.SuperDebug {
			log.Printf("Forum #%+v\n", fid)
		}
		forumPerm, ok := forumPerms[group.ID][fid]
		if ok {
			//log.Print("Overriding permissions for forum #%d",fid)
			group.Forums = append(group.Forums, forumPerm)
		} else {
			//log.Printf("Inheriting from group defaults for forum #%d",fid)
			forumPerm = BlankForumPerms
			group.Forums = append(group.Forums, forumPerm)
		}
		if forumPerm.Overrides {
			if forumPerm.ViewTopic {
				group.CanSee = append(group.CanSee, fid)
			}
		} else if group.Perms.ViewTopic {
			group.CanSee = append(group.CanSee, fid)
		}

		if dev.SuperDebug {
			log.Print("group.ID: ", group.ID)
			log.Printf("forumPerm: %+v\n", forumPerm)
			log.Print("group.CanSee: ", group.CanSee)
		}
	}
}

func (fps *ForumPermsStore) Get(fid int, gid int) (fperms ForumPerms, err error) {
	// TODO: Add a hook here and have plugin_guilds use it
	group, err := gstore.Get(gid)
	if err != nil {
		return fperms, ErrNoRows
	}
	return group.Forums[fid], nil
}
