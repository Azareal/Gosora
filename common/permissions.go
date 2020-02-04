package common

import (
	"encoding/json"
	"log"

	"github.com/Azareal/Gosora/common/phrases"
	qgen "github.com/Azareal/Gosora/query_gen"
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
	"UseConvos",
	"CreateProfileReply",
	"AutoEmbed",
}

// Permission Structure: ActionComponent[Subcomponent]Flag
type Perms struct {
	// Global Permissions
	BanUsers              bool `json:",omitempty"`
	ActivateUsers         bool `json:",omitempty"`
	EditUser              bool `json:",omitempty"`
	EditUserEmail         bool `json:",omitempty"`
	EditUserPassword      bool `json:",omitempty"`
	EditUserGroup         bool `json:",omitempty"`
	EditUserGroupSuperMod bool `json:",omitempty"`
	EditUserGroupAdmin    bool `json:",omitempty"`
	EditGroup             bool `json:",omitempty"`
	EditGroupLocalPerms   bool `json:",omitempty"`
	EditGroupGlobalPerms  bool `json:",omitempty"`
	EditGroupSuperMod     bool `json:",omitempty"`
	EditGroupAdmin        bool `json:",omitempty"`
	ManageForums          bool `json:",omitempty"` // This could be local, albeit limited for per-forum managers?
	EditSettings          bool `json:",omitempty"`
	ManageThemes          bool `json:",omitempty"`
	ManagePlugins         bool `json:",omitempty"`
	ViewAdminLogs         bool `json:",omitempty"`
	ViewIPs               bool `json:",omitempty"`

	// Global non-staff permissions
	UploadFiles        bool `json:",omitempty"`
	UploadAvatars      bool `json:",omitempty"`
	UseConvos          bool `json:",omitempty"`
	CreateProfileReply bool `json:",omitempty"`
	AutoEmbed          bool `json:",omitempty"`

	// Forum permissions
	ViewTopic bool `json:",omitempty"`
	//ViewOwnTopic bool `json:",omitempty"`
	LikeItem    bool `json:",omitempty"`
	CreateTopic bool `json:",omitempty"`
	EditTopic   bool `json:",omitempty"`
	DeleteTopic bool `json:",omitempty"`
	CreateReply bool `json:",omitempty"`
	//CreateReplyToOwn bool `json:",omitempty"`
	EditReply bool `json:",omitempty"`
	//EditOwnReply bool `json:",omitempty"`
	DeleteReply bool `json:",omitempty"`
	//DeleteOwnReply bool `json:",omitempty"`
	PinTopic   bool `json:",omitempty"`
	CloseTopic bool `json:",omitempty"`
	//CloseOwnTopic bool `json:",omitempty"`
	MoveTopic bool `json:",omitempty"`

	//ExtData map[string]bool `json:",omitempty"`
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

		UploadFiles:        true,
		UploadAvatars:      true,
		UseConvos:          true,
		CreateProfileReply: true,
		AutoEmbed:          true,

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
func RebuildGroupPermissions(g *Group) error {
	var permstr []byte
	log.Print("Reloading a group")

	// TODO: Avoid re-initting this all the time
	getGroupPerms, err := qgen.Builder.SimpleSelect("users_groups", "permissions", "gid=?", "", "")
	if err != nil {
		return err
	}
	defer getGroupPerms.Close()

	err = getGroupPerms.QueryRow(g.ID).Scan(&permstr)
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
	g.Perms = tmpPerms
	return nil
}

func OverridePerms(p *Perms, status bool) {
	if status {
		*p = AllPerms
	} else {
		*p = BlankPerms
	}
}

// TODO: We need a better way of overriding forum perms rather than setting them one by one
func OverrideForumPerms(p *Perms, status bool) {
	p.ViewTopic = status
	p.LikeItem = status
	p.CreateTopic = status
	p.EditTopic = status
	p.DeleteTopic = status
	p.CreateReply = status
	p.EditReply = status
	p.DeleteReply = status
	p.PinTopic = status
	p.CloseTopic = status
	p.MoveTopic = status
}

func RegisterPluginPerm(name string) {
	AllPluginPerms[name] = true
}

func DeregisterPluginPerm(name string) {
	delete(AllPluginPerms, name)
}
