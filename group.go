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

func (group *Group) Copy() Group {
	return *group
}
