package common

import (
	"database/sql"
	"encoding/json"

	"../query_gen/lib"
)

var blankGroup = Group{ID: 0, Name: ""}

type GroupAdmin struct {
	ID        int
	Name      string
	Rank      string
	RankClass string
	CanEdit   bool
	CanDelete bool
}

// ! Fix the data races in the fperms
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
	CanSee          []int // The IDs of the forums this group can see
	UserCount       int   // ! Might be temporary as I might want to lean on the database instead for this
}

type GroupStmts struct {
	updateGroup      *sql.Stmt
	updateGroupRank  *sql.Stmt
	updateGroupPerms *sql.Stmt
}

var groupStmts GroupStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		groupStmts = GroupStmts{
			updateGroup:      acc.Update("users_groups").Set("name = ?, tag = ?").Where("gid = ?").Prepare(),
			updateGroupRank:  acc.Update("users_groups").Set("is_admin = ?, is_mod = ?, is_banned = ?").Where("gid = ?").Prepare(),
			updateGroupPerms: acc.Update("users_groups").Set("permissions = ?").Where("gid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

func (group *Group) ChangeRank(isAdmin bool, isMod bool, isBanned bool) (err error) {
	_, err = groupStmts.updateGroupRank.Exec(isAdmin, isMod, isBanned, group.ID)
	if err != nil {
		return err
	}

	Groups.Reload(group.ID)
	return nil
}

func (group *Group) Update(name string, tag string) (err error) {
	_, err = groupStmts.updateGroup.Exec(name, tag, group.ID)
	if err != nil {
		return err
	}

	Groups.Reload(group.ID)
	return nil
}

// Please don't pass arbitrary inputs to this method
func (group *Group) UpdatePerms(perms map[string]bool) (err error) {
	pjson, err := json.Marshal(perms)
	if err != nil {
		return err
	}
	_, err = groupStmts.updateGroupPerms.Exec(pjson, group.ID)
	if err != nil {
		return err
	}
	return RebuildGroupPermissions(group.ID)
}

// Copy gives you a non-pointer concurrency safe copy of the group
func (group *Group) Copy() Group {
	return *group
}

func (group *Group) CopyPtr() (co *Group) {
	co = new(Group)
	*co = *group
	return co
}

// TODO: Replace this sorting mechanism with something a lot more efficient
// ? - Use sort.Slice instead?
type SortGroup []*Group

func (sg SortGroup) Len() int {
	return len(sg)
}
func (sg SortGroup) Swap(i, j int) {
	sg[i], sg[j] = sg[j], sg[i]
}
func (sg SortGroup) Less(i, j int) bool {
	return sg[i].ID < sg[j].ID
}
