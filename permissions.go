package main

import "log"
import "sync"
import "strconv"
import "encoding/json"

var permUpdateMutex sync.Mutex
var BlankPerms Perms
var BlankForumPerms ForumPerms
var GuestPerms Perms
var ReadForumPerms ForumPerms
var ReadReplyForumPerms ForumPerms
var ReadWriteForumPerms ForumPerms
var AllPerms Perms
var AllForumPerms ForumPerms
var AllPluginPerms = make(map[string]bool)

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

	// Forum permissions
	ViewTopic   bool
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
	ViewTopic   bool
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

	guestUser.Perms = GuestPerms

	if dev.DebugMode {
		log.Printf("Guest Perms: %+v\n", GuestPerms)
		log.Printf("All Perms: %+v\n", AllPerms)
	}
}

func presetToPermmap(preset string) (out map[string]ForumPerms) {
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

func permmapToQuery(permmap map[string]ForumPerms, fid int) error {
	permUpdateMutex.Lock()
	defer permUpdateMutex.Unlock()

	_, err := deleteForumPermsByForumStmt.Exec(fid)
	if err != nil {
		return err
	}

	perms, err := json.Marshal(permmap["admins"])
	if err != nil {
		return err
	}
	_, err = addForumPermsToForumAdminsStmt.Exec(fid, "", perms)
	if err != nil {
		return err
	}

	perms, err = json.Marshal(permmap["staff"])
	if err != nil {
		return err
	}
	_, err = addForumPermsToForumStaffStmt.Exec(fid, "", perms)
	if err != nil {
		return err
	}

	perms, err = json.Marshal(permmap["members"])
	if err != nil {
		return err
	}
	_, err = addForumPermsToForumMembersStmt.Exec(fid, "", perms)
	if err != nil {
		return err
	}

	perms, err = json.Marshal(permmap["guests"])
	if err != nil {
		return err
	}
	_, err = addForumPermsToGroupStmt.Exec(6, fid, "", perms)
	if err != nil {
		return err
	}

	return rebuildForumPermissions(fid)
}

// TODO: Need a more thread-safe way of doing this. Possibly with sync.Map?
func rebuildForumPermissions(fid int) error {
	if dev.DebugMode {
		log.Print("Loading the forum permissions")
	}
	forums, err := fstore.GetAll()
	if err != nil {
		return err
	}

	rows, err := db.Query("select gid, permissions from forums_permissions where fid = ? order by gid asc", fid)
	if err != nil {
		return err
	}
	defer rows.Close()

	if dev.DebugMode {
		log.Print("Updating the forum permissions")
	}
	for rows.Next() {
		var gid int
		var perms []byte
		var pperms ForumPerms
		err := rows.Scan(&gid, &perms)
		if err != nil {
			return err
		}
		err = json.Unmarshal(perms, &pperms)
		if err != nil {
			return err
		}
		pperms.ExtData = make(map[string]bool)
		pperms.Overrides = true
		_, ok := forumPerms[gid]
		if !ok {
			forumPerms[gid] = make(map[int]ForumPerms)
		}
		forumPerms[gid][fid] = pperms
	}

	groups, err := gstore.GetAll()
	if err != nil {
		return err
	}

	for _, group := range groups {
		if dev.DebugMode {
			log.Print("Updating the forum permissions for Group #" + strconv.Itoa(group.ID))
		}
		var blankList []ForumPerms
		var blankIntList []int
		group.Forums = blankList
		group.CanSee = blankIntList

		for ffid := range forums {
			forumPerm, ok := forumPerms[group.ID][ffid]
			if ok {
				//log.Print("Overriding permissions for forum #" + strconv.Itoa(fid))
				group.Forums = append(group.Forums, forumPerm)
			} else {
				//log.Print("Inheriting from default for forum #" + strconv.Itoa(fid))
				forumPerm = BlankForumPerms
				group.Forums = append(group.Forums, forumPerm)
			}

			if forumPerm.Overrides {
				if forumPerm.ViewTopic {
					group.CanSee = append(group.CanSee, ffid)
				}
			} else if group.Perms.ViewTopic {
				group.CanSee = append(group.CanSee, ffid)
			}
		}
		if dev.SuperDebug {
			log.Printf("group.CanSee %+v\n", group.CanSee)
			log.Printf("group.Forums %+v\n", group.Forums)
			log.Print("len(group.Forums)", len(group.Forums))
		}
	}
	return nil
}

// ? - We could have buildForumPermissions and rebuildForumPermissions call a third function containing common logic?
func buildForumPermissions() error {
	forums, err := fstore.GetAll()
	if err != nil {
		return err
	}

	rows, err := getForumsPermissionsStmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	if dev.DebugMode {
		log.Print("Adding the forum permissions")
	}
	// Temporarily store the forum perms in a map before transferring it to a much faster and thread-safe slice
	forumPerms = make(map[int]map[int]ForumPerms)
	for rows.Next() {
		var gid, fid int
		var perms []byte
		var pperms ForumPerms
		err = rows.Scan(&gid, &fid, &perms)
		if err != nil {
			return err
		}
		err = json.Unmarshal(perms, &pperms)
		if err != nil {
			return err
		}
		pperms.ExtData = make(map[string]bool)
		pperms.Overrides = true
		_, ok := forumPerms[gid]
		if !ok {
			forumPerms[gid] = make(map[int]ForumPerms)
		}
		forumPerms[gid][fid] = pperms
	}

	groups, err := gstore.GetAll()
	if err != nil {
		return err

	}
	for _, group := range groups {
		if dev.DebugMode {
			log.Print("Adding the forum permissions for Group #" + strconv.Itoa(group.ID) + " - " + group.Name)
		}
		//groups[gid].Forums = append(groups[gid].Forums,BlankForumPerms) // GID 0. No longer needed now that Uncategorised occupies that slot
		for fid := range forums {
			forumPerm, ok := forumPerms[group.ID][fid]
			if ok {
				// Override group perms
				//log.Print("Overriding permissions for forum #" + strconv.Itoa(fid))
				group.Forums = append(group.Forums, forumPerm)
			} else {
				// Inherit from Group
				//log.Print("Inheriting from default for forum #" + strconv.Itoa(fid))
				forumPerm = BlankForumPerms
				group.Forums = append(group.Forums, forumPerm)
			}

			if forumPerm.Overrides {
				if forumPerm.ViewTopic {
					group.CanSee = append(group.CanSee, fid)
				}
			} else if group.Perms.ViewTopic {
				group.CanSee = append(group.CanSee, fid)
			}
		}
		if dev.SuperDebug {
			//log.Printf("groups[gid].CanSee %+v\n", groups[gid].CanSee)
			//log.Printf("groups[gid].Forums %+v\n", groups[gid].Forums)
			//log.Print("len(groups[gid].CanSee)",len(groups[gid].CanSee))
			//log.Print("len(groups[gid].Forums)",len(groups[gid].Forums))
		}
	}
	return nil
}

func forumPermsToGroupForumPreset(fperms ForumPerms) string {
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

func groupForumPresetToForumPerms(preset string) (fperms ForumPerms, changed bool) {
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

func stripInvalidGroupForumPreset(preset string) string {
	switch preset {
	case "read_only", "can_post", "can_moderate", "no_access", "default", "custom":
		return preset
	}
	return ""
}

func stripInvalidPreset(preset string) string {
	switch preset {
	case "all", "announce", "members", "staff", "admins", "archive", "custom":
		return preset
	default:
		return ""
	}
}

// TODO: Move this into the phrase system?
func presetToLang(preset string) string {
	switch preset {
	case "all":
		return "Public"
	case "announce":
		return "Announcements"
	case "members":
		return "Member Only"
	case "staff":
		return "Staff Only"
	case "admins":
		return "Admin Only"
	case "archive":
		return "Archive"
	case "custom":
		return "Custom"
	default:
		return ""
	}
}

// TODO: Is this racey?
func rebuildGroupPermissions(gid int) error {
	var permstr []byte
	log.Print("Reloading a group")
	err := db.QueryRow("select permissions from users_groups where gid = ?", gid).Scan(&permstr)
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

	group, err := gstore.Get(gid)
	if err != nil {
		return err
	}
	group.Perms = tmpPerms
	return nil
}

func overridePerms(perms *Perms, status bool) {
	if status {
		*perms = AllPerms
	} else {
		*perms = BlankPerms
	}
}

// TODO: We need a better way of overriding forum perms rather than setting them one by one
func overrideForumPerms(perms *Perms, status bool) {
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

func registerPluginPerm(name string) {
	AllPluginPerms[name] = true
}

func deregisterPluginPerm(name string) {
	delete(AllPluginPerms, name)
}
