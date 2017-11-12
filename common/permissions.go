package common

import (
	"database/sql"
	"encoding/json"
	"log"
	"sync"

	"../query_gen/lib"
)

// TODO: Refactor the perms system

var PermUpdateMutex sync.Mutex
var BlankPerms Perms
var BlankForumPerms ForumPerms
var GuestPerms Perms
var ReadForumPerms ForumPerms
var ReadReplyForumPerms ForumPerms
var ReadWriteForumPerms ForumPerms

// AllPerms is a set of global permissions with everything set to true
var AllPerms Perms

// AllForumPerms is a set of forum local permissions with everything set to true
var AllForumPerms ForumPerms
var AllPluginPerms = make(map[string]bool)

// ? - Can we avoid duplicating the items in this list in a bunch of places?

var LocalPermList = []string{
	"ViewTopic",
	"LikeItem",
	"CreateTopic",
	"EditTopic",
	"DeleteTopic",
	"CreateReply",
	"EditReply",
	"DeleteReply",
	"PinTopic",
	"CloseTopic",
}

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
	// TODO: Add a permission for enabling avatars

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
	PinTopic    bool
	CloseTopic  bool
	//CloseOwnTopic bool

	//ExtData map[string]bool
}

/* Inherit from group permissions for ones we don't have */
type ForumPerms struct {
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
	PinTopic    bool
	CloseTopic  bool
	//CloseOwnTopic bool

	Overrides bool
	ExtData   map[string]bool
}

func init() {
	BlankPerms = Perms{
	//ExtData: make(map[string]bool),
	}

	BlankForumPerms = ForumPerms{
		ExtData: make(map[string]bool),
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

		//ExtData: make(map[string]bool),
	}

	AllForumPerms = ForumPerms{
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

		Overrides: true,
		ExtData:   make(map[string]bool),
	}

	ReadWriteForumPerms = ForumPerms{
		ViewTopic:   true,
		LikeItem:    true,
		CreateTopic: true,
		CreateReply: true,
		Overrides:   true,
		ExtData:     make(map[string]bool),
	}

	ReadReplyForumPerms = ForumPerms{
		ViewTopic:   true,
		LikeItem:    true,
		CreateReply: true,
		Overrides:   true,
		ExtData:     make(map[string]bool),
	}

	ReadForumPerms = ForumPerms{
		ViewTopic: true,
		Overrides: true,
		ExtData:   make(map[string]bool),
	}

	GuestUser.Perms = GuestPerms

	if Dev.DebugMode {
		log.Printf("Guest Perms: %+v\n", GuestPerms)
		log.Printf("All Perms: %+v\n", AllPerms)
	}
}

func PresetToPermmap(preset string) (out map[string]ForumPerms) {
	out = make(map[string]ForumPerms)
	switch preset {
	case "all":
		out["guests"] = ReadForumPerms
		out["members"] = ReadWriteForumPerms
		out["staff"] = AllForumPerms
		out["admins"] = AllForumPerms
	case "announce":
		out["guests"] = ReadForumPerms
		out["members"] = ReadReplyForumPerms
		out["staff"] = AllForumPerms
		out["admins"] = AllForumPerms
	case "members":
		out["guests"] = BlankForumPerms
		out["members"] = ReadWriteForumPerms
		out["staff"] = AllForumPerms
		out["admins"] = AllForumPerms
	case "staff":
		out["guests"] = BlankForumPerms
		out["members"] = BlankForumPerms
		out["staff"] = ReadWriteForumPerms
		out["admins"] = AllForumPerms
	case "admins":
		out["guests"] = BlankForumPerms
		out["members"] = BlankForumPerms
		out["staff"] = BlankForumPerms
		out["admins"] = AllForumPerms
	case "archive":
		out["guests"] = ReadForumPerms
		out["members"] = ReadForumPerms
		out["staff"] = ReadForumPerms
		out["admins"] = ReadForumPerms //CurateForumPerms. Delete / Edit but no create?
	default:
		out["guests"] = BlankForumPerms
		out["members"] = BlankForumPerms
		out["staff"] = BlankForumPerms
		out["admins"] = BlankForumPerms
	}
	return out
}

func PermmapToQuery(permmap map[string]ForumPerms, fid int) error {
	tx, err := qgen.Builder.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteForumPermsByForumTx, err := qgen.Builder.SimpleDeleteTx(tx, "forums_permissions", "fid = ?")
	if err != nil {
		return err
	}

	_, err = deleteForumPermsByForumTx.Exec(fid)
	if err != nil {
		return err
	}

	perms, err := json.Marshal(permmap["admins"])
	if err != nil {
		return err
	}

	addForumPermsToForumAdminsTx, err := qgen.Builder.SimpleInsertSelectTx(tx,
		qgen.DBInsert{"forums_permissions", "gid, fid, preset, permissions", ""},
		qgen.DBSelect{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 1", "", ""},
	)
	if err != nil {
		return err
	}

	_, err = addForumPermsToForumAdminsTx.Exec(fid, "", perms)
	if err != nil {
		return err
	}

	perms, err = json.Marshal(permmap["staff"])
	if err != nil {
		return err
	}

	addForumPermsToForumStaffTx, err := qgen.Builder.SimpleInsertSelectTx(tx,
		qgen.DBInsert{"forums_permissions", "gid, fid, preset, permissions", ""},
		qgen.DBSelect{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 0 AND is_mod = 1", "", ""},
	)
	if err != nil {
		return err
	}
	_, err = addForumPermsToForumStaffTx.Exec(fid, "", perms)
	if err != nil {
		return err
	}

	perms, err = json.Marshal(permmap["members"])
	if err != nil {
		return err
	}

	addForumPermsToForumMembersTx, err := qgen.Builder.SimpleInsertSelectTx(tx,
		qgen.DBInsert{"forums_permissions", "gid, fid, preset, permissions", ""},
		qgen.DBSelect{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 0 AND is_mod = 0 AND is_banned = 0", "", ""},
	)
	if err != nil {
		return err
	}
	_, err = addForumPermsToForumMembersTx.Exec(fid, "", perms)
	if err != nil {
		return err
	}

	// 6 is the ID of the Not Loggedin Group
	// TODO: Use a shared variable rather than a literal for the group ID
	err = ReplaceForumPermsForGroupTx(tx, 6, map[int]string{fid: ""}, map[int]ForumPerms{fid: permmap["guests"]})
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	PermUpdateMutex.Lock()
	defer PermUpdateMutex.Unlock()
	return Fpstore.Reload(fid)
}

func ReplaceForumPermsForGroup(gid int, presetSet map[int]string, permSets map[int]ForumPerms) error {
	tx, err := qgen.Builder.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = ReplaceForumPermsForGroupTx(tx, gid, presetSet, permSets)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func ReplaceForumPermsForGroupTx(tx *sql.Tx, gid int, presetSets map[int]string, permSets map[int]ForumPerms) error {
	deleteForumPermsForGroupTx, err := qgen.Builder.SimpleDeleteTx(tx, "forums_permissions", "gid = ? AND fid = ?")
	if err != nil {
		return err
	}

	addForumPermsToGroupTx, err := qgen.Builder.SimpleInsertTx(tx, "forums_permissions", "gid, fid, preset, permissions", "?,?,?,?")
	if err != nil {
		return err
	}
	for fid, permSet := range permSets {
		permstr, err := json.Marshal(permSet)
		if err != nil {
			return err
		}
		_, err = deleteForumPermsForGroupTx.Exec(gid, fid)
		if err != nil {
			return err
		}
		_, err = addForumPermsToGroupTx.Exec(gid, fid, presetSets[fid], string(permstr))
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Refactor this and write tests for it
func ForumPermsToGroupForumPreset(fperms ForumPerms) string {
	if !fperms.Overrides {
		return "default"
	}
	if !fperms.ViewTopic {
		return "no_access"
	}
	var canPost = (fperms.LikeItem && fperms.CreateTopic && fperms.CreateReply)
	var canModerate = (canPost && fperms.EditTopic && fperms.DeleteTopic && fperms.EditReply && fperms.DeleteReply && fperms.PinTopic && fperms.CloseTopic)
	if canModerate {
		return "can_moderate"
	}
	if fperms.EditTopic || fperms.DeleteTopic || fperms.EditReply || fperms.DeleteReply || fperms.PinTopic || fperms.CloseTopic {
		if !canPost {
			return "custom"
		}
		return "quasi_mod"
	}

	if canPost {
		return "can_post"
	}
	if fperms.ViewTopic && !fperms.LikeItem && !fperms.CreateTopic && !fperms.CreateReply {
		return "read_only"
	}
	return "custom"
}

func GroupForumPresetToForumPerms(preset string) (fperms ForumPerms, changed bool) {
	switch preset {
	case "read_only":
		return ReadForumPerms, true
	case "can_post":
		return ReadWriteForumPerms, true
	case "can_moderate":
		return AllForumPerms, true
	case "no_access":
		return ForumPerms{Overrides: true, ExtData: make(map[string]bool)}, true
	case "default":
		return BlankForumPerms, true
		//case "custom": return fperms, false
	}
	return fperms, false
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
	default:
		return ""
	}
}

// TODO: Move this into the phrase system?
func PresetToLang(preset string) string {
	phrases := GetAllPermPresets()
	phrase, ok := phrases[preset]
	if !ok {
		phrase = phrases["unknown"]
	}
	return phrase
}

// TODO: Is this racey?
// TODO: Test this along with the rest of the perms system
func RebuildGroupPermissions(gid int) error {
	var permstr []byte
	log.Print("Reloading a group")

	// TODO: Avoid re-initting this all the time
	getGroupPerms, err := qgen.Builder.SimpleSelect("users_groups", "permissions", "gid = ?", "", "")
	if err != nil {
		return err
	}
	defer getGroupPerms.Close()

	err = getGroupPerms.QueryRow(gid).Scan(&permstr)
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

	group, err := Gstore.Get(gid)
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
}

func RegisterPluginPerm(name string) {
	AllPluginPerms[name] = true
}

func DeregisterPluginPerm(name string) {
	delete(AllPluginPerms, name)
}
