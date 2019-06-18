package common

import (
	"encoding/json"
	"log"

	"github.com/Azareal/Gosora/query_gen"
	"github.com/Azareal/Gosora/common/phrases"
)

// TODO: Refactor the perms system
var BlankPerms Perms
var GuestPerms Perms

// AllPerms is a set of global permissions with everything set to true
var AllPerms Perms
var AllPluginPerms = make(map[string]bool)

// ? - Can we avoid duplicating the items in this list in a bunch of places?
var GlobalPermList = []string{
	"BanUsers",
	"ActivateUsers",
	"EditUser",
	"EditUserEmail",
	"EditUserPassword",
	"EditUserGroup",
	"EditUserGroupSuperMod",
	"EditUserGroupAdmin",
	"EditGroup",
	"EditGroupLocalPerms",
	"EditGroupGlobalPerms",
	"EditGroupSuperMod",
	"EditGroupAdmin",
	"ManageForums",
	"EditSettings",
	"ManageThemes",
	"ManagePlugins",
	"ViewAdminLogs",
	"ViewIPs",
	"UploadFiles",
	"UploadAvatars",
}

// Permission Structure: ActionComponent[Subcomponent]Flag
type Perms struct {
	// Global Permissions
	BanUsers              bool
	ActivateUsers         bool
	EditUser              bool
	EditUserEmail         bool
	EditUserPassword      bool
	EditUserGroup         bool
	EditUserGroupSuperMod bool
	EditUserGroupAdmin    bool
	EditGroup             bool
	EditGroupLocalPerms   bool
	EditGroupGlobalPerms  bool
	EditGroupSuperMod     bool
	EditGroupAdmin        bool
	ManageForums          bool // This could be local, albeit limited for per-forum managers?
	EditSettings          bool
	ManageThemes          bool
	ManagePlugins         bool
	ViewAdminLogs         bool
	ViewIPs               bool

	// Global non-staff permissions
	UploadFiles bool
	UploadAvatars bool

	// Forum permissions
	ViewTopic bool
	//ViewOwnTopic bool
	LikeItem    bool
	CreateTopic bool
	EditTopic   bool
	DeleteTopic bool
	CreateReply bool
	//CreateReplyToOwn bool
	EditReply bool
	//EditOwnReply bool
	DeleteReply bool
	//DeleteOwnReply bool
	PinTopic   bool
	CloseTopic bool
	//CloseOwnTopic bool
	MoveTopic bool

	//ExtData map[string]bool
}

func init() {
	BlankPerms = Perms{
	//ExtData: make(map[string]bool),
	}

	GuestPerms = Perms{
		ViewTopic: true,
		//ExtData: make(map[string]bool),
	}

	AllPerms = Perms{
		BanUsers:              true,
		ActivateUsers:         true,
		EditUser:              true,
		EditUserEmail:         true,
		EditUserPassword:      true,
		EditUserGroup:         true,
		EditUserGroupSuperMod: true,
		EditUserGroupAdmin:    true,
		EditGroup:             true,
		EditGroupLocalPerms:   true,
		EditGroupGlobalPerms:  true,
		EditGroupSuperMod:     true,
		EditGroupAdmin:        true,
		ManageForums:          true,
		EditSettings:          true,
		ManageThemes:          true,
		ManagePlugins:         true,
		ViewAdminLogs:         true,
		ViewIPs:               true,

		UploadFiles: true,
		UploadAvatars: true,

		ViewTopic:   true,
		LikeItem:    true,
		CreateTopic: true,
		EditTopic:   true,
		DeleteTopic: true,
		CreateReply: true,
		EditReply:   true,
		DeleteReply: true,
		PinTopic:    true,
		CloseTopic:  true,
		MoveTopic:   true,

		//ExtData: make(map[string]bool),
	}

	GuestUser.Perms = GuestPerms
	DebugLogf("Guest Perms: %+v\n", GuestPerms)
	DebugLogf("All Perms: %+v\n", AllPerms)
}

func StripInvalidGroupForumPreset(preset string) string {
	switch preset {
	case "read_only", "can_post", "can_moderate", "no_access", "default", "custom":
		return preset
	}
	return ""
}

func StripInvalidPreset(preset string) string {
	switch preset {
	case "all", "announce", "members", "staff", "admins", "archive", "custom":
		return preset
	}
	return ""
}

// TODO: Move this into the phrase system?
func PresetToLang(preset string) string {
	phrases := phrases.GetAllPermPresets()
	phrase, ok := phrases[preset]
	if !ok {
		phrase = phrases["unknown"]
	}
	return phrase
}

// TODO: Is this racey?
// TODO: Test this along with the rest of the perms system
func RebuildGroupPermissions(group *Group) error {
	var permstr []byte
	log.Print("Reloading a group")

	// TODO: Avoid re-initting this all the time
	getGroupPerms, err := qgen.Builder.SimpleSelect("users_groups", "permissions", "gid = ?", "", "")
	if err != nil {
		return err
	}
	defer getGroupPerms.Close()

	err = getGroupPerms.QueryRow(group.ID).Scan(&permstr)
	if err != nil {
		return err
	}

	tmpPerms := Perms{
	//ExtData: make(map[string]bool),
	}
	err = json.Unmarshal(permstr, &tmpPerms)
	if err != nil {
		return err
	}
	group.Perms = tmpPerms
	return nil
}

func OverridePerms(perms *Perms, status bool) {
	if status {
		*perms = AllPerms
	} else {
		*perms = BlankPerms
	}
}

// TODO: We need a better way of overriding forum perms rather than setting them one by one
func OverrideForumPerms(perms *Perms, status bool) {
	perms.ViewTopic = status
	perms.LikeItem = status
	perms.CreateTopic = status
	perms.EditTopic = status
	perms.DeleteTopic = status
	perms.CreateReply = status
	perms.EditReply = status
	perms.DeleteReply = status
	perms.PinTopic = status
	perms.CloseTopic = status
	perms.MoveTopic = status
}

func RegisterPluginPerm(name string) {
	AllPluginPerms[name] = true
}

func DeregisterPluginPerm(name string) {
	delete(AllPluginPerms, name)
}
