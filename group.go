package main

var blankGroup = Group{ID: 0, Name: ""}

type GroupAdmin struct {
	ID        int
	Name      string
	Rank      string
	RankClass string
	CanEdit   bool
	CanDelete bool
}

// ! Fix the data races
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

// TODO: Reload the group from the database rather than modifying it via it's pointer
func (group *Group) ChangeRank(isAdmin bool, isMod bool, isBanned bool) (err error) {
	_, err = updateGroupRankStmt.Exec(isAdmin, isMod, isBanned, group.ID)
	if err != nil {
		return err
	}

	group.IsAdmin = isAdmin
	group.IsMod = isMod
	if isAdmin || isMod {
		group.IsBanned = false
	} else {
		group.IsBanned = isBanned
	}

	return nil
}

// ! Ahem, don't listen to the comment below. It's not concurrency safe right now.
// Copy gives you a non-pointer concurrency safe copy of the group
func (group *Group) Copy() Group {
	return *group
}
