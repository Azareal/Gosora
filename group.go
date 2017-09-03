package main

import "sync"
import "encoding/json"

var groupUpdateMutex sync.Mutex

type GroupAdmin struct {
	ID        int
	Name      string
	Rank      string
	RankClass string
	CanEdit   bool
	CanDelete bool
}

type Group struct {
	ID              int
	Name            string
	IsMod           bool
	IsAdmin         bool
	IsBanned        bool
	Tag             string
	Perms           Perms
	PermissionsText []byte
	PluginPerms     map[string]bool // Custom permissions defined by plugins. What if two plugins declare the same permission, but they handle them in incompatible ways? Very unlikely, we probably don't need to worry about this, the plugin authors should be aware of each other to some extent
	PluginPermsText []byte
	Forums          []ForumPerms
	CanSee          []int // The IDs of the forums this group can see
}

var groupCreateMutex sync.Mutex

func createGroup(groupName string, tag string, isAdmin bool, isMod bool, isBanned bool) (int, error) {
	var gid int
	err := group_entry_exists_stmt.QueryRow().Scan(&gid)
	if err != nil && err != ErrNoRows {
		return 0, err
	}
	if err != ErrNoRows {
		groupUpdateMutex.Lock()
		_, err = update_group_rank_stmt.Exec(isAdmin, isMod, isBanned, gid)
		if err != nil {
			return gid, err
		}
		_, err = update_group_stmt.Exec(groupName, tag, gid)
		if err != nil {
			return gid, err
		}

		groups[gid].Name = groupName
		groups[gid].Tag = tag
		groups[gid].IsBanned = isBanned
		groups[gid].IsMod = isMod
		groups[gid].IsAdmin = isAdmin

		groupUpdateMutex.Unlock()
		return gid, nil
	}

	groupCreateMutex.Lock()
	var permstr = "{}"
	res, err := create_group_stmt.Exec(groupName, tag, isAdmin, isMod, isBanned, permstr)
	if err != nil {
		return 0, err
	}

	gid64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	gid = int(gid64)

	var perms = BlankPerms
	var blankForums []ForumPerms
	var blankIntList []int
	var pluginPerms = make(map[string]bool)
	var pluginPermsBytes = []byte("{}")
	if vhooks["create_group_preappend"] != nil {
		runVhook("create_group_preappend", &pluginPerms, &pluginPermsBytes)
	}

	groups = append(groups, Group{gid, groupName, isMod, isAdmin, isBanned, tag, perms, []byte(permstr), pluginPerms, pluginPermsBytes, blankForums, blankIntList})
	groupCreateMutex.Unlock()

	// Generate the forum permissions based on the presets...
	fdata, err := fstore.GetAll()
	if err != nil {
		return 0, err
	}

	permUpdateMutex.Lock()
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
		_, err = add_forum_perms_to_group_stmt.Exec(gid, forum.ID, forum.Preset, perms)
		if err != nil {
			return gid, err
		}

		err = rebuildForumPermissions(forum.ID)
		if err != nil {
			return gid, err
		}
	}
	permUpdateMutex.Unlock()
	return gid, nil
}

func groupExists(gid int) bool {
	return (gid <= groupCapCount) && (gid > 0) && groups[gid].Name != ""
}
