/* Under Heavy Construction */
package main

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
)

var groupCreateMutex sync.Mutex
var groupUpdateMutex sync.Mutex
var gstore GroupStore

type GroupStore interface {
	LoadGroups() error
	DirtyGet(id int) *Group
	Get(id int) (*Group, error)
	GetCopy(id int) (Group, error)
	Exists(id int) bool
	Create(groupName string, tag string, isAdmin bool, isMod bool, isBanned bool) (int, error)
	GetAll() ([]*Group, error)
	GetRange(lower int, higher int) ([]*Group, error)
}

type MemoryGroupStore struct {
	groups        []*Group // TODO: Use a sync.Map instead of a slice
	groupCapCount int
}

func NewMemoryGroupStore() *MemoryGroupStore {
	return &MemoryGroupStore{}
}

func (mgs *MemoryGroupStore) LoadGroups() error {
	mgs.groups = []*Group{&Group{ID: 0, Name: "Unknown"}}

	rows, err := getGroupsStmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 1
	for ; rows.Next(); i++ {
		group := Group{ID: 0}
		err := rows.Scan(&group.ID, &group.Name, &group.PermissionsText, &group.PluginPermsText, &group.IsMod, &group.IsAdmin, &group.IsBanned, &group.Tag)
		if err != nil {
			return err
		}

		err = json.Unmarshal(group.PermissionsText, &group.Perms)
		if err != nil {
			return err
		}
		if dev.DebugMode {
			log.Print(group.Name + ": ")
			log.Printf("%+v\n", group.Perms)
		}

		err = json.Unmarshal(group.PluginPermsText, &group.PluginPerms)
		if err != nil {
			return err
		}
		if dev.DebugMode {
			log.Print(group.Name + ": ")
			log.Printf("%+v\n", group.PluginPerms)
		}

		//group.Perms.ExtData = make(map[string]bool)
		mgs.groups = append(mgs.groups, &group)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	mgs.groupCapCount = i

	if dev.DebugMode {
		log.Print("Binding the Not Loggedin Group")
	}
	GuestPerms = mgs.groups[6].Perms
	return nil
}

func (mgs *MemoryGroupStore) DirtyGet(gid int) *Group {
	if !mgs.Exists(gid) {
		return &blankGroup
	}
	return mgs.groups[gid]
}

func (mgs *MemoryGroupStore) Get(gid int) (*Group, error) {
	if !mgs.Exists(gid) {
		return nil, ErrNoRows
	}
	return mgs.groups[gid], nil
}

func (mgs *MemoryGroupStore) GetCopy(gid int) (Group, error) {
	if !mgs.Exists(gid) {
		return blankGroup, ErrNoRows
	}
	return *mgs.groups[gid], nil
}

func (mgs *MemoryGroupStore) Exists(gid int) bool {
	return (gid <= mgs.groupCapCount) && (gid > -1) && mgs.groups[gid].Name != ""
}

func (mgs *MemoryGroupStore) Create(groupName string, tag string, isAdmin bool, isMod bool, isBanned bool) (int, error) {
	groupCreateMutex.Lock()
	defer groupCreateMutex.Unlock()

	var permstr = "{}"
	res, err := createGroupStmt.Exec(groupName, tag, isAdmin, isMod, isBanned, permstr)
	if err != nil {
		return 0, err
	}

	gid64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	var gid = int(gid64)

	var perms = BlankPerms
	var blankForums []ForumPerms
	var blankIntList []int
	var pluginPerms = make(map[string]bool)
	var pluginPermsBytes = []byte("{}")
	if vhooks["create_group_preappend"] != nil {
		runVhook("create_group_preappend", &pluginPerms, &pluginPermsBytes)
	}

	mgs.groups = append(mgs.groups, &Group{gid, groupName, isMod, isAdmin, isBanned, tag, perms, []byte(permstr), pluginPerms, pluginPermsBytes, blankForums, blankIntList})

	// Generate the forum permissions based on the presets...
	fdata, err := fstore.GetAll()
	if err != nil {
		return 0, err
	}

	permUpdateMutex.Lock()
	defer permUpdateMutex.Unlock()
	for _, forum := range fdata {
		var thePreset string
		if isAdmin {
			thePreset = "admins"
		} else if isMod {
			thePreset = "staff"
		} else if isBanned {
			thePreset = "banned"
		} else {
			thePreset = "members"
		}

		permmap := presetToPermmap(forum.Preset)
		permitem := permmap[thePreset]
		permitem.Overrides = true
		permstr, err := json.Marshal(permitem)
		if err != nil {
			return gid, err
		}
		perms := string(permstr)
		_, err = addForumPermsToGroupStmt.Exec(gid, forum.ID, forum.Preset, perms)
		if err != nil {
			return gid, err
		}

		err = rebuildForumPermissions(forum.ID)
		if err != nil {
			return gid, err
		}
	}
	return gid, nil
}

// ! NOT CONCURRENT
func (mgs *MemoryGroupStore) GetAll() ([]*Group, error) {
	return mgs.groups, nil
}

// ? - It's currently faster to use GetAll(), but we'll be dropping the guarantee that the slices are ordered soon
// ? - Set the lower and higher numbers to 0 to remove the bounds
// ? - Currently uses slicing for efficiency, so it might behave a little weirdly
func (mgs *MemoryGroupStore) GetRange(lower int, higher int) (groups []*Group, err error) {
	if lower == 0 && higher == 0 {
		return mgs.GetAll()
	} else if lower == 0 {
		if higher < 0 {
			return nil, errors.New("higher may not be lower than 0")
		}
		if higher > len(mgs.groups) {
			higher = len(mgs.groups)
		}
		groups = mgs.groups[:higher]
	} else if higher == 0 {
		if lower < 0 {
			return nil, errors.New("lower may not be lower than 0")
		}
		groups = mgs.groups[lower:]
	}
	return groups, nil
}
