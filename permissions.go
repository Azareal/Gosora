package main

import "log"
import "fmt"
import "sync"
import "strconv"
import "encoding/json"

var permupdate_mutex sync.Mutex
var BlankPerms Perms
var BlankForumPerms ForumPerms
var GuestPerms Perms
var ReadForumPerms ForumPerms
var ReadReplyForumPerms ForumPerms
var ReadWriteForumPerms ForumPerms
var AllPerms Perms
var AllForumPerms ForumPerms

var LocalPermList []string = []string{
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

var GlobalPermList []string = []string{
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
	EditGroup bool
	EditGroupLocalPerms bool
	EditGroupGlobalPerms bool
	EditGroupSuperMod bool
	EditGroupAdmin bool
	ManageForums bool // This could be local, albeit limited for per-forum managers?
	EditSettings bool
	ManageThemes bool
	ManagePlugins bool
	ViewAdminLogs bool
	ViewIPs bool

	// Forum permissions
	ViewTopic bool
	LikeItem bool
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
	LikeItem bool
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

	Overrides bool
	ExtData map[string]bool
}

func init() {
	BlankPerms = Perms{
		ExtData: make(map[string]bool),
	}

	BlankForumPerms = ForumPerms{
		ExtData: make(map[string]bool),
	}

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
		EditGroup: true,
		EditGroupLocalPerms: true,
		EditGroupGlobalPerms: true,
		EditGroupSuperMod: true,
		EditGroupAdmin: true,
		ManageForums: true,
		EditSettings: true,
		ManageThemes: true,
		ManagePlugins: true,
		ViewAdminLogs: true,
		ViewIPs: true,

		ViewTopic: true,
		LikeItem: true,
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

	AllForumPerms = ForumPerms{
		ViewTopic: true,
		LikeItem: true,
		CreateTopic: true,
		EditTopic: true,
		DeleteTopic: true,
		CreateReply: true,
		EditReply: true,
		DeleteReply: true,
		PinTopic: true,
		CloseTopic: true,

		Overrides: true,
		ExtData: make(map[string]bool),
	}

	ReadWriteForumPerms = ForumPerms{
		ViewTopic: true,
		LikeItem: true,
		CreateTopic: true,
		CreateReply: true,
		Overrides: true,
		ExtData: make(map[string]bool),
	}

	ReadReplyForumPerms = ForumPerms{
		ViewTopic: true,
		LikeItem: true,
		CreateReply: true,
		Overrides: true,
		ExtData: make(map[string]bool),
	}

	ReadForumPerms = ForumPerms{
		ViewTopic: true,
		Overrides: true,
		ExtData: make(map[string]bool),
	}

	guest_user.Perms = GuestPerms

	if debug {
		fmt.Printf("Guest Perms: ")
		fmt.Printf("%+v\n", GuestPerms)
		fmt.Printf("All Perms: ")
		fmt.Printf("%+v\n", AllPerms)
	}
}

func preset_to_permmap(preset string) (out map[string]ForumPerms) {
	out = make(map[string]ForumPerms)
	switch(preset) {
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

func permmap_to_query(permmap map[string]ForumPerms, fid int) error {
	permupdate_mutex.Lock()
	defer permupdate_mutex.Unlock()

	_, err := delete_forum_perms_by_forum_stmt.Exec(fid)
	if err != nil {
		return err
	}

	perms, err := json.Marshal(permmap["admins"])
	if err != nil {
		return err
	}
	_, err = add_forum_perms_to_forum_admins_stmt.Exec(fid,"",perms)
	if err != nil {
		return err
	}

	perms, err = json.Marshal(permmap["staff"])
	if err != nil {
		return err
	}
	_, err = add_forum_perms_to_forum_staff_stmt.Exec(fid,"",perms)
	if err != nil {
		return err
	}

	perms, err = json.Marshal(permmap["members"])
	if err != nil {
		return err
	}
	_, err = add_forum_perms_to_forum_members_stmt.Exec(fid,"",perms)
	if err != nil {
		return err
	}

	perms, err = json.Marshal(permmap["guests"])
	if err != nil {
		return err
	}
	_, err = add_forum_perms_to_group_stmt.Exec(6,fid,"",perms)
	if err != nil {
		return err
	}

	return rebuild_forum_permissions(fid)
}

func rebuild_forum_permissions(fid int) error {
	log.Print("Loading the forum permissions")
	rows, err := db.Query("select gid, permissions from forums_permissions where fid = ? order by gid asc", fid)
	if err != nil {
		return err
	}
	defer rows.Close()

	log.Print("Updating the forum permissions")
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
		_, ok := forum_perms[gid]
		if !ok {
			forum_perms[gid] = make(map[int]ForumPerms)
		}
		forum_perms[gid][fid] = pperms
	}
	for gid, _ := range groups {
		log.Print("Updating the forum permissions for Group #" + strconv.Itoa(gid))
		var blank_list []ForumPerms
		var blank_int_list []int
		groups[gid].Forums = blank_list
		groups[gid].CanSee = blank_int_list

		for ffid, _ := range forums {
			forum_perm, ok := forum_perms[gid][ffid]
			if ok {
				//log.Print("Overriding permissions for forum #" + strconv.Itoa(fid))
				groups[gid].Forums = append(groups[gid].Forums,forum_perm)
			} else {
				//log.Print("Inheriting from default for forum #" + strconv.Itoa(fid))
				forum_perm = BlankForumPerms
				groups[gid].Forums = append(groups[gid].Forums,forum_perm)
			}

			if forum_perm.Overrides {
				if forum_perm.ViewTopic {
					groups[gid].CanSee = append(groups[gid].CanSee, ffid)
				}
			} else if groups[gid].Perms.ViewTopic {
				groups[gid].CanSee = append(groups[gid].CanSee, ffid)
			}
		}
		//fmt.Printf("%+v\n", groups[gid].CanSee)
		//fmt.Printf("%+v\n", groups[gid].Forums)
		//fmt.Println(len(groups[gid].Forums))
	}
	return nil
}

func build_forum_permissions() error {
	return nil
}

func forum_perms_to_group_forum_preset(fperms ForumPerms) string {
	if !fperms.Overrides {
		return "default"
	}
	if !fperms.ViewTopic {
		return "no_access"
	}
	var can_post bool = (fperms.LikeItem && fperms.CreateTopic && fperms.CreateReply)
	var can_moderate bool = (can_post && fperms.EditTopic && fperms.DeleteTopic && fperms.EditReply && fperms.DeleteReply && fperms.PinTopic && fperms.CloseTopic)
	if can_moderate {
		return "can_moderate"
	}
	if (fperms.EditTopic || fperms.DeleteTopic || fperms.EditReply || fperms.DeleteReply || fperms.PinTopic || fperms.CloseTopic) {
		if !can_post {
			return "custom"
		}
		return "quasi_mod"
	}

	if can_post {
		return "can_post"
	}
	if fperms.ViewTopic && !fperms.LikeItem && !fperms.CreateTopic && !fperms.CreateReply {
		return "read_only"
	}
	return "custom"
}

func group_forum_preset_to_forum_perms(preset string) (fperms ForumPerms, changed bool) {
	switch(preset) {
		case "read_only":
			return ReadForumPerms, true
		case "can_post":
			return ReadWriteForumPerms, true
		case "can_moderate":
			return AllForumPerms, true
		case "no_access":
			return ForumPerms{Overrides: true}, true
		case "default":
			return BlankForumPerms, true
		//case "custom": return fperms, false
	}
	return fperms, false
}

func strip_invalid_group_forum_preset(preset string) string {
	switch(preset) {
	case "read_only","can_post","can_moderate","no_access","default","custom":
			return preset
	}
	return ""
}

func strip_invalid_preset(preset string) string {
	switch(preset) {
		case "all","announce","members","staff","admins","archive","custom":
			break
		default: return ""
	}
	return preset
}

func preset_to_lang(preset string) string {
	switch(preset) {
		case "all": return "Public"
		case "announce": return "Announcements"
		case "members": return "Member Only"
		case "staff": return "Staff Only"
		case "admins": return "Admin Only"
		case "archive": return "Archive"
		case "custom": return "Custom"
	}
	return ""
}

func rebuild_group_permissions(gid int) error {
	var permstr []byte
	log.Print("Reloading a group")
	err := db.QueryRow("select permissions from users_groups where gid = ?",gid).Scan(&permstr)
	if err != nil {
		return err
	}

	tmp_perms := Perms{ExtData: make(map[string]bool)}
	err = json.Unmarshal(permstr, &tmp_perms)
	if err != nil {
		return err
	}

	groups[gid].Perms = tmp_perms
	return nil
}
