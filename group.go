package main

import "sync"

var group_update_mutex sync.Mutex

type GroupAdmin struct
{
	ID int
	Name string
	Rank string
	RankEmoji string
	CanEdit bool
	CanDelete bool
}

type Group struct
{
	ID int
	Name string
	Is_Mod bool
	Is_Admin bool
	Is_Banned bool
	Tag string
	Perms Perms
	PermissionsText []byte
	Forums []ForumPerms
	CanSee []int // The IDs of the forums this group can see
}

func group_exists(gid int) bool {
	//fmt.Println(gid <= groupCapCount)
	//fmt.Println(gid > 0)
	//fmt.Println(groups[gid].Name!="")
	return (gid <= groupCapCount) && (gid > 0) && groups[gid].Name!=""
}
