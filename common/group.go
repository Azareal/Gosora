package common

import "database/sql"
import "../query_gen/lib"

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
	Forums          []ForumPerms
	CanSee          []int // The IDs of the forums this group can see
}

type GroupStmts struct {
	updateGroupRank *sql.Stmt
}

var groupStmts GroupStmts

func init() {
	DbInits.Add(func() error {
		acc := qgen.Builder.Accumulator()
		groupStmts = GroupStmts{
			updateGroupRank: acc.SimpleUpdate("users_groups", "is_admin = ?, is_mod = ?, is_banned = ?", "gid = ?"),
		}
		return acc.FirstError()
	})
}

func (group *Group) ChangeRank(isAdmin bool, isMod bool, isBanned bool) (err error) {
	_, err = groupStmts.updateGroupRank.Exec(isAdmin, isMod, isBanned, group.ID)
	if err != nil {
		return err
	}

	Gstore.Reload(group.ID)
	return nil
}

// Copy gives you a non-pointer concurrency safe copy of the group
func (group *Group) Copy() Group {
	return *group
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
