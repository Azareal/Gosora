package common

import (
	"database/sql"
	"encoding/json"

	"../query_gen/lib"
)

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
	"MoveTopic",
}

// TODO: Rename this to ForumPermSet?
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
	MoveTopic bool

	Overrides bool
	ExtData   map[string]bool
}

func PresetToPermmap(preset string) (out map[string]*ForumPerms) {
	out = make(map[string]*ForumPerms)
	switch preset {
	case "all":
		out["guests"] = ReadForumPerms()
		out["members"] = ReadWriteForumPerms()
		out["staff"] = AllForumPerms()
		out["admins"] = AllForumPerms()
	case "announce":
		out["guests"] = ReadForumPerms()
		out["members"] = ReadReplyForumPerms()
		out["staff"] = AllForumPerms()
		out["admins"] = AllForumPerms()
	case "members":
		out["guests"] = BlankForumPerms()
		out["members"] = ReadWriteForumPerms()
		out["staff"] = AllForumPerms()
		out["admins"] = AllForumPerms()
	case "staff":
		out["guests"] = BlankForumPerms()
		out["members"] = BlankForumPerms()
		out["staff"] = ReadWriteForumPerms()
		out["admins"] = AllForumPerms()
	case "admins":
		out["guests"] = BlankForumPerms()
		out["members"] = BlankForumPerms()
		out["staff"] = BlankForumPerms()
		out["admins"] = AllForumPerms()
	case "archive":
		out["guests"] = ReadForumPerms()
		out["members"] = ReadForumPerms()
		out["staff"] = ReadForumPerms()
		out["admins"] = ReadForumPerms() //CurateForumPerms. Delete / Edit but no create?
	default:
		out["guests"] = BlankForumPerms()
		out["members"] = BlankForumPerms()
		out["staff"] = BlankForumPerms()
		out["admins"] = BlankForumPerms()
	}
	return out
}

func PermmapToQuery(permmap map[string]*ForumPerms, fid int) error {
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

	// TODO: The group ID is probably a variable somewhere. Find it and use it.
	// Group 5 is the Awaiting Activation group
	err = ReplaceForumPermsForGroupTx(tx, 5, map[int]string{fid: ""}, map[int]*ForumPerms{fid: permmap["guests"]})
	if err != nil {
		return err
	}

	// TODO: Consult a config setting instead of GuestUser?
	err = ReplaceForumPermsForGroupTx(tx, GuestUser.Group, map[int]string{fid: ""}, map[int]*ForumPerms{fid: permmap["guests"]})
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	err = FPStore.Reload(fid)
	if err != nil {
		return err
	}
	return TopicList.RebuildPermTree()
}

// TODO: FPStore.Reload?
func ReplaceForumPermsForGroup(gid int, presetSet map[int]string, permSets map[int]*ForumPerms) error {
	tx, err := qgen.Builder.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = ReplaceForumPermsForGroupTx(tx, gid, presetSet, permSets)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return TopicList.RebuildPermTree()
}

func ReplaceForumPermsForGroupTx(tx *sql.Tx, gid int, presetSets map[int]string, permSets map[int]*ForumPerms) error {
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
// TODO: We really need to improve the thread safety of this
func ForumPermsToGroupForumPreset(fperms *ForumPerms) string {
	if !fperms.Overrides {
		return "default"
	}
	if !fperms.ViewTopic {
		return "no_access"
	}
	var canPost = (fperms.LikeItem && fperms.CreateTopic && fperms.CreateReply)
	var canModerate = (canPost && fperms.EditTopic && fperms.DeleteTopic && fperms.EditReply && fperms.DeleteReply && fperms.PinTopic && fperms.CloseTopic && fperms.MoveTopic)
	if canModerate {
		return "can_moderate"
	}
	if fperms.EditTopic || fperms.DeleteTopic || fperms.EditReply || fperms.DeleteReply || fperms.PinTopic || fperms.CloseTopic || fperms.MoveTopic {
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

func GroupForumPresetToForumPerms(preset string) (fperms *ForumPerms, changed bool) {
	switch preset {
	case "read_only":
		return ReadForumPerms(), true
	case "can_post":
		return ReadWriteForumPerms(), true
	case "can_moderate":
		return AllForumPerms(), true
	case "no_access":
		return &ForumPerms{Overrides: true, ExtData: make(map[string]bool)}, true
	case "default":
		return BlankForumPerms(), true
	}
	return fperms, false
}

func BlankForumPerms() *ForumPerms {
	return &ForumPerms{ViewTopic: false}
}

func ReadWriteForumPerms() *ForumPerms {
	return &ForumPerms{
		ViewTopic:   true,
		LikeItem:    true,
		CreateTopic: true,
		CreateReply: true,
		Overrides:   true,
		ExtData:     make(map[string]bool),
	}
}

func ReadReplyForumPerms() *ForumPerms {
	return &ForumPerms{
		ViewTopic:   true,
		LikeItem:    true,
		CreateReply: true,
		Overrides:   true,
		ExtData:     make(map[string]bool),
	}
}

func ReadForumPerms() *ForumPerms {
	return &ForumPerms{
		ViewTopic: true,
		Overrides: true,
		ExtData:   make(map[string]bool),
	}
}

// AllForumPerms is a set of forum local permissions with everything set to true
func AllForumPerms() *ForumPerms {
	return &ForumPerms{
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

		Overrides: true,
		ExtData:   make(map[string]bool),
	}
}
