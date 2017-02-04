package main

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
