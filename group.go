package main
import "fmt"

var GuestPerms Perms
var AllPerms Perms

type Group struct
{
	ID int
	Name string
	Perms Perms
	PermissionsText []byte
	Is_Mod bool
	Is_Admin bool
	Is_Banned bool
	Tag string
}

// Permission Structure: ActionComponent[Subcomponent]Flag
type Perms struct
{
	// Global Permissions
	BanUsers bool
	ActivateUsers bool
	EditUser bool
	EditUserEmail bool
	EditUserPassword bool
	EditUserGroup bool
	EditUserGroupSuperMod bool
	EditUserGroupAdmin bool
	ManageForums bool // This could be local, albeit limited for per-forum managers
	EditSettings bool
	ManageThemes bool
	ManagePlugins bool
	ViewIPs bool
	
	// Forum permissions
	ViewTopic bool
	CreateTopic bool
	EditTopic bool
	DeleteTopic bool
	CreateReply bool
	//CreateReplyToOwn bool
	EditReply bool
	//EditOwnReply bool
	DeleteReply bool
	PinTopic bool
	CloseTopic bool
	//CloseOwnTopic bool
	
	ExtData interface{}
}

/* Inherit from group permissions for ones we don't have */
type ForumPerms struct
{
	ViewTopic bool
	CreateTopic bool
	EditTopic bool
	DeleteTopic bool
	CreateReply bool
	//CreateReplyToOwn bool
	EditReply bool
	//EditOwnReply bool
	DeleteReply bool
	PinTopic bool
	CloseTopic bool
	//CloseOwnTopic bool
	
	ExtData map[string]bool
}

func init() {
	GuestPerms = Perms{
		ViewTopic: true,
		ExtData: make(map[string]bool),
	}
	
	AllPerms = Perms{
		BanUsers: true,
		ActivateUsers: true,
		EditUser: true,
		EditUserEmail: true,
		EditUserPassword: true,
		EditUserGroup: true,
		EditUserGroupSuperMod: true,
		EditUserGroupAdmin: true,
		ManageForums: true,
		EditSettings: true,
		ManageThemes: true,
		ManagePlugins: true,
		ViewIPs: true,
		
		ViewTopic: true,
		CreateTopic: true,
		EditTopic: true,
		DeleteTopic: true,
		CreateReply: true,
		EditReply: true,
		DeleteReply: true,
		PinTopic: true,
		CloseTopic: true,
		
		ExtData: make(map[string]bool),
	}
	
	if debug {
		fmt.Printf("Guest Perms: ")
		fmt.Printf("%+v\n", GuestPerms)
		fmt.Printf("All Perms: ")
		fmt.Printf("%+v\n", AllPerms)
	}
}