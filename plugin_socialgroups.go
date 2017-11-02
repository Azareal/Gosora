package main

import (
	//"fmt"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"html"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"./query_gen/lib"
)

var socialgroupsListStmt *sql.Stmt
var socialgroupsMemberListStmt *sql.Stmt
var socialgroupsMemberListJoinStmt *sql.Stmt
var socialgroupsGetMemberStmt *sql.Stmt
var socialgroupsGetGroupStmt *sql.Stmt
var socialgroupsCreateGroupStmt *sql.Stmt
var socialgroupsAttachForumStmt *sql.Stmt
var socialgroupsUnattachForumStmt *sql.Stmt
var socialgroupsAddMemberStmt *sql.Stmt

// TODO: Add a better way of splitting up giant plugins like this

// SocialGroup is a struct representing a social group
type SocialGroup struct {
	ID      int
	Link    string
	Name    string
	Desc    string
	Active  bool
	Privacy int /* 0: Public, 1: Protected, 2: Private */

	// Who should be able to accept applications and create invites? Mods+ or just admins? Mods is a good start, we can ponder over whether we should make this more flexible in the future.
	Joinable int /* 0: Private, 1: Anyone can join, 2: Applications, 3: Invite-only */

	MemberCount    int
	Owner          int
	Backdrop       string
	CreatedAt      string
	LastUpdateTime string

	MainForumID int
	MainForum   *Forum
	Forums      []*Forum
	ExtData     ExtData
}

type SocialGroupPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	ItemList    []*TopicsRow
	Forum       *Forum
	SocialGroup *SocialGroup
	Page        int
	LastPage    int
}

// SocialGroupListPage is a page struct for constructing a list of every social group
type SocialGroupListPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	GroupList   []*SocialGroup
}

type SocialGroupMemberListPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	ItemList    []SocialGroupMember
	SocialGroup *SocialGroup
	Page        int
	LastPage    int
}

// SocialGroupMember is a struct representing a specific member of a group, not to be confused with the global User struct.
type SocialGroupMember struct {
	Link       string
	Rank       int    /* 0: Member. 1: Mod. 2: Admin. */
	RankString string /* Member, Mod, Admin, Owner */
	PostCount  int
	JoinedAt   string
	Offline    bool // TODO: Need to track the online states of members when WebSockets are enabled

	User User
}

// TODO: Add a plugin interface instead of having a bunch of argument to AddPlugin?
func init() {
	plugins["socialgroups"] = NewPlugin("socialgroups", "Social Groups", "Azareal", "http://github.com/Azareal", "", "", "", initSocialgroups, nil, deactivateSocialgroups, installSocialgroups, nil)
}

func initSocialgroups() (err error) {
	plugins["socialgroups"].AddHook("intercept_build_widgets", socialgroupsWidgets)
	plugins["socialgroups"].AddHook("trow_assign", socialgroupsTrowAssign)
	plugins["socialgroups"].AddHook("topic_create_pre_loop", socialgroupsTopicCreatePreLoop)
	plugins["socialgroups"].AddHook("pre_render_view_forum", socialgroupsPreRenderViewForum)
	plugins["socialgroups"].AddHook("simple_forum_check_pre_perms", socialgroupsForumCheck)
	plugins["socialgroups"].AddHook("forum_check_pre_perms", socialgroupsForumCheck)
	// TODO: Auto-grant this perm to admins upon installation?
	registerPluginPerm("CreateSocialGroup")
	router.HandleFunc("/groups/", socialgroupsGroupList)
	router.HandleFunc("/group/", socialgroupsViewGroup)
	router.HandleFunc("/group/create/", socialgroupsCreateGroup)
	router.HandleFunc("/group/create/submit/", socialgroupsCreateGroupSubmit)
	router.HandleFunc("/group/members/", socialgroupsMemberList)

	socialgroupsListStmt, err = qgen.Builder.SimpleSelect("socialgroups", "sgid, name, desc, active, privacy, joinable, owner, memberCount, createdAt, lastUpdateTime", "", "", "")
	if err != nil {
		return err
	}
	socialgroupsGetGroupStmt, err = qgen.Builder.SimpleSelect("socialgroups", "name, desc, active, privacy, joinable, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime", "sgid = ?", "", "")
	if err != nil {
		return err
	}
	socialgroupsMemberListStmt, err = qgen.Builder.SimpleSelect("socialgroups_members", "sgid, uid, rank, posts, joinedAt", "", "", "")
	if err != nil {
		return err
	}
	socialgroupsMemberListJoinStmt, err = qgen.Builder.SimpleLeftJoin("socialgroups_members", "users", "users.uid, socialgroups_members.rank, socialgroups_members.posts, socialgroups_members.joinedAt, users.name, users.avatar", "socialgroups_members.uid = users.uid", "socialgroups_members.sgid = ?", "socialgroups_members.rank DESC, socialgroups_members.joinedat ASC", "")
	if err != nil {
		return err
	}
	socialgroupsGetMemberStmt, err = qgen.Builder.SimpleSelect("socialgroups_members", "rank, posts, joinedAt", "sgid = ? AND uid = ?", "", "")
	if err != nil {
		return err
	}
	socialgroupsCreateGroupStmt, err = qgen.Builder.SimpleInsert("socialgroups", "name, desc, active, privacy, joinable, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime", "?,?,?,?,1,?,1,?,'',UTC_TIMESTAMP(),UTC_TIMESTAMP()")
	if err != nil {
		return err
	}
	socialgroupsAttachForumStmt, err = qgen.Builder.SimpleUpdate("forums", "parentID = ?, parentType = 'socialgroup'", "fid = ?")
	if err != nil {
		return err
	}
	socialgroupsUnattachForumStmt, err = qgen.Builder.SimpleUpdate("forums", "parentID = 0, parentType = ''", "fid = ?")
	if err != nil {
		return err
	}
	socialgroupsAddMemberStmt, err = qgen.Builder.SimpleInsert("socialgroups_members", "sgid, uid, rank, posts, joinedAt", "?,?,?,0,UTC_TIMESTAMP()")
	if err != nil {
		return err
	}

	return nil
}

func deactivateSocialgroups() {
	plugins["socialgroups"].RemoveHook("intercept_build_widgets", socialgroupsWidgets)
	plugins["socialgroups"].RemoveHook("trow_assign", socialgroupsTrowAssign)
	plugins["socialgroups"].RemoveHook("topic_create_pre_loop", socialgroupsTopicCreatePreLoop)
	plugins["socialgroups"].RemoveHook("pre_render_view_forum", socialgroupsPreRenderViewForum)
	plugins["socialgroups"].RemoveHook("simple_forum_check_pre_perms", socialgroupsForumCheck)
	plugins["socialgroups"].RemoveHook("forum_check_pre_perms", socialgroupsForumCheck)
	deregisterPluginPerm("CreateSocialGroup")
	_ = router.RemoveFunc("/groups/")
	_ = router.RemoveFunc("/group/")
	_ = router.RemoveFunc("/group/create/")
	_ = router.RemoveFunc("/group/create/submit/")
	_ = socialgroupsListStmt.Close()
	_ = socialgroupsMemberListStmt.Close()
	_ = socialgroupsMemberListJoinStmt.Close()
	_ = socialgroupsGetMemberStmt.Close()
	_ = socialgroupsGetGroupStmt.Close()
	_ = socialgroupsCreateGroupStmt.Close()
	_ = socialgroupsAttachForumStmt.Close()
	_ = socialgroupsUnattachForumStmt.Close()
	_ = socialgroupsAddMemberStmt.Close()
}

// TODO: Stop accessing the query builder directly and add a feature in Gosora which is more easily reversed, if an error comes up during the installation process
func installSocialgroups() error {
	sgTableStmt, err := qgen.Builder.CreateTable("socialgroups", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"sgid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"name", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"desc", "varchar", 200, false, false, ""},
			qgen.DB_Table_Column{"active", "boolean", 1, false, false, ""},
			qgen.DB_Table_Column{"privacy", "smallint", 0, false, false, ""},
			qgen.DB_Table_Column{"joinable", "smallint", 0, false, false, "0"},
			qgen.DB_Table_Column{"owner", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"memberCount", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"mainForum", "int", 0, false, false, "0"}, // The board the user lands on when they click on a group, we'll make it possible for group admins to change what users land on
			//qgen.DB_Table_Column{"boards","varchar",255,false,false,""}, // Cap the max number of boards at 8 to avoid overflowing the confines of a 64-bit integer?
			qgen.DB_Table_Column{"backdrop", "varchar", 200, false, false, ""}, // File extension for the uploaded file, or an external link
			qgen.DB_Table_Column{"createdAt", "createdAt", 0, false, false, ""},
			qgen.DB_Table_Column{"lastUpdateTime", "datetime", 0, false, false, ""},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"sgid", "primary"},
		},
	)
	if err != nil {
		return err
	}

	_, err = sgTableStmt.Exec()
	if err != nil {
		return err
	}

	sgMembersTableStmt, err := qgen.Builder.CreateTable("socialgroups_members", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"sgid", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"uid", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"rank", "int", 0, false, false, "0"},  /* 0: Member. 1: Mod. 2: Admin. */
			qgen.DB_Table_Column{"posts", "int", 0, false, false, "0"}, /* Per-Group post count. Should we do some sort of score system? */
			qgen.DB_Table_Column{"joinedAt", "datetime", 0, false, false, ""},
		},
		[]qgen.DB_Table_Key{},
	)
	if err != nil {
		return err
	}

	_, err = sgMembersTableStmt.Exec()
	return err
}

// TO-DO; Implement an uninstallation system into Gosora. And a better installation system.
func uninstallSocialgroups() error {
	return nil
}

// TODO: Do this properly via the widget system
func socialgroupsCommonAreaWidgets(headerVars *HeaderVars) {
	// TODO: Hot Groups? Featured Groups? Official Groups?
	var b bytes.Buffer
	var menu = WidgetMenu{"Social Groups", []WidgetMenuItem{
		WidgetMenuItem{"Create Group", "/group/create/", false},
	}}

	err := templates.ExecuteTemplate(&b, "widget_menu.html", menu)
	if err != nil {
		LogError(err)
		return
	}

	if themes[headerVars.ThemeName].Sidebars == "left" {
		headerVars.Widgets.LeftSidebar = template.HTML(string(b.Bytes()))
	} else if themes[headerVars.ThemeName].Sidebars == "right" || themes[headerVars.ThemeName].Sidebars == "both" {
		headerVars.Widgets.RightSidebar = template.HTML(string(b.Bytes()))
	}
}

// TODO: Do this properly via the widget system
// TODO: Make a better more customisable group widget system
func socialgroupsGroupWidgets(headerVars *HeaderVars, sgItem *SocialGroup) (success bool) {
	return false // Disabled until the next commit

	/*var b bytes.Buffer
	var menu WidgetMenu = WidgetMenu{"Group Options", []WidgetMenuItem{
		WidgetMenuItem{"Join", "/group/join/" + strconv.Itoa(sgItem.ID), false},
		WidgetMenuItem{"Members", "/group/members/" + strconv.Itoa(sgItem.ID), false},
	}}

	err := templates.ExecuteTemplate(&b, "widget_menu.html", menu)
	if err != nil {
		LogError(err)
		return false
	}

	if themes[headerVars.ThemeName].Sidebars == "left" {
		headerVars.Widgets.LeftSidebar = template.HTML(string(b.Bytes()))
	} else if themes[headerVars.ThemeName].Sidebars == "right" || themes[headerVars.ThemeName].Sidebars == "both" {
		headerVars.Widgets.RightSidebar = template.HTML(string(b.Bytes()))
	} else {
		return false
	}
	return true*/
}

/*
	Custom Pages
*/

func socialgroupsGroupList(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	socialgroupsCommonAreaWidgets(headerVars)

	rows, err := socialgroupsListStmt.Query()
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}

	var sgList []*SocialGroup
	for rows.Next() {
		sgItem := &SocialGroup{ID: 0}
		err := rows.Scan(&sgItem.ID, &sgItem.Name, &sgItem.Desc, &sgItem.Active, &sgItem.Privacy, &sgItem.Joinable, &sgItem.Owner, &sgItem.MemberCount, &sgItem.CreatedAt, &sgItem.LastUpdateTime)
		if err != nil {
			return InternalError(err, w, r)
		}
		sgItem.Link = socialgroupsBuildGroupURL(nameToSlug(sgItem.Name), sgItem.ID)
		sgList = append(sgList, sgItem)
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}
	rows.Close()

	pi := SocialGroupListPage{"Group List", user, headerVars, sgList}
	err = templates.ExecuteTemplate(w, "socialgroups_group_list.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func socialgroupsGetGroup(sgid int) (sgItem *SocialGroup, err error) {
	sgItem = &SocialGroup{ID: sgid}
	err = socialgroupsGetGroupStmt.QueryRow(sgid).Scan(&sgItem.Name, &sgItem.Desc, &sgItem.Active, &sgItem.Privacy, &sgItem.Joinable, &sgItem.Owner, &sgItem.MemberCount, &sgItem.MainForumID, &sgItem.Backdrop, &sgItem.CreatedAt, &sgItem.LastUpdateTime)
	return sgItem, err
}

func socialgroupsViewGroup(w http.ResponseWriter, r *http.Request, user User) RouteError {
	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/group/"):], ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	sgid, err := strconv.Atoi(halves[1])
	if err != nil {
		return PreError("Not a valid group ID", w, r)
	}

	sgItem, err := socialgroupsGetGroup(sgid)
	if err != nil {
		return LocalError("Bad group", w, r, user)
	}
	if !sgItem.Active {
		return NotFound(w, r)
	}

	// Re-route the request to routeForums
	var ctx = context.WithValue(r.Context(), "socialgroups_current_group", sgItem)
	return routeForum(w, r.WithContext(ctx), user, strconv.Itoa(sgItem.MainForumID))
}

func socialgroupsCreateGroup(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	// TODO: Add an approval queue mode for group creation
	if !user.Loggedin || !user.PluginPerms["CreateSocialGroup"] {
		return NoPermissions(w, r, user)
	}
	socialgroupsCommonAreaWidgets(headerVars)

	pi := Page{"Create Group", user, headerVars, tList, nil}
	err := templates.ExecuteTemplate(w, "socialgroups_create_group.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func socialgroupsCreateGroupSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	// TODO: Add an approval queue mode for group creation
	if !user.Loggedin || !user.PluginPerms["CreateSocialGroup"] {
		return NoPermissions(w, r, user)
	}

	var groupActive = true
	var groupName = html.EscapeString(r.PostFormValue("group_name"))
	var groupDesc = html.EscapeString(r.PostFormValue("group_desc"))
	var gprivacy = r.PostFormValue("group_privacy")

	var groupPrivacy int
	switch gprivacy {
	case "0":
		groupPrivacy = 0 // Public
	case "1":
		groupPrivacy = 1 // Protected
	case "2":
		groupPrivacy = 2 // private
	default:
		groupPrivacy = 0
	}

	// Create the backing forum
	fid, err := fstore.Create(groupName, "", true, "")
	if err != nil {
		return InternalError(err, w, r)
	}

	res, err := socialgroupsCreateGroupStmt.Exec(groupName, groupDesc, groupActive, groupPrivacy, user.ID, fid)
	if err != nil {
		return InternalError(err, w, r)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return InternalError(err, w, r)
	}

	// Add the main backing forum to the forum list
	err = socialgroupsAttachForum(int(lastID), fid)
	if err != nil {
		return InternalError(err, w, r)
	}

	_, err = socialgroupsAddMemberStmt.Exec(lastID, user.ID, 2)
	if err != nil {
		return InternalError(err, w, r)
	}

	http.Redirect(w, r, socialgroupsBuildGroupURL(nameToSlug(groupName), int(lastID)), http.StatusSeeOther)
	return nil
}

func socialgroupsMemberList(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/group/members/"):], ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	sgid, err := strconv.Atoi(halves[1])
	if err != nil {
		return PreError("Not a valid group ID", w, r)
	}

	var sgItem = &SocialGroup{ID: sgid}
	var mainForum int // Unused
	err = socialgroupsGetGroupStmt.QueryRow(sgid).Scan(&sgItem.Name, &sgItem.Desc, &sgItem.Active, &sgItem.Privacy, &sgItem.Joinable, &sgItem.Owner, &sgItem.MemberCount, &mainForum, &sgItem.Backdrop, &sgItem.CreatedAt, &sgItem.LastUpdateTime)
	if err != nil {
		return LocalError("Bad group", w, r, user)
	}
	sgItem.Link = socialgroupsBuildGroupURL(nameToSlug(sgItem.Name), sgItem.ID)

	socialgroupsGroupWidgets(headerVars, sgItem)

	rows, err := socialgroupsMemberListJoinStmt.Query(sgid)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}

	var sgMembers []SocialGroupMember
	for rows.Next() {
		sgMember := SocialGroupMember{PostCount: 0}
		err := rows.Scan(&sgMember.User.ID, &sgMember.Rank, &sgMember.PostCount, &sgMember.JoinedAt, &sgMember.User.Name, &sgMember.User.Avatar)
		if err != nil {
			return InternalError(err, w, r)
		}
		sgMember.Link = buildProfileURL(nameToSlug(sgMember.User.Name), sgMember.User.ID)
		if sgMember.User.Avatar != "" {
			if sgMember.User.Avatar[0] == '.' {
				sgMember.User.Avatar = "/uploads/avatar_" + strconv.Itoa(sgMember.User.ID) + sgMember.User.Avatar
			}
		} else {
			sgMember.User.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(sgMember.User.ID), 1)
		}
		sgMember.JoinedAt, _ = relativeTimeFromString(sgMember.JoinedAt)
		if sgItem.Owner == sgMember.User.ID {
			sgMember.RankString = "Owner"
		} else {
			switch sgMember.Rank {
			case 0:
				sgMember.RankString = "Member"
			case 1:
				sgMember.RankString = "Mod"
			case 2:
				sgMember.RankString = "Admin"
			}
		}
		sgMembers = append(sgMembers, sgMember)
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}
	rows.Close()

	pi := SocialGroupMemberListPage{"Group Member List", user, headerVars, sgMembers, sgItem, 0, 0}
	// A plugin with plugins. Pluginception!
	if preRenderHooks["pre_render_socialgroups_member_list"] != nil {
		if runPreRenderHook("pre_render_socialgroups_member_list", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "socialgroups_member_list.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func socialgroupsAttachForum(sgid int, fid int) error {
	_, err := socialgroupsAttachForumStmt.Exec(sgid, fid)
	return err
}

func socialgroupsUnattachForum(fid int) error {
	_, err := socialgroupsAttachForumStmt.Exec(fid)
	return err
}

func socialgroupsBuildGroupURL(slug string, id int) string {
	if slug == "" {
		return "/group/" + slug + "." + strconv.Itoa(id)
	}
	return "/group/" + strconv.Itoa(id)
}

/*
	Hooks
*/

func socialgroupsPreRenderViewForum(w http.ResponseWriter, r *http.Request, user *User, data interface{}) (halt bool) {
	pi := data.(*ForumPage)
	if pi.Header.ExtData.items != nil {
		if sgData, ok := pi.Header.ExtData.items["socialgroups_current_group"]; ok {
			sgItem := sgData.(*SocialGroup)

			sgpi := SocialGroupPage{pi.Title, pi.CurrentUser, pi.Header, pi.ItemList, pi.Forum, sgItem, pi.Page, pi.LastPage}
			err := templates.ExecuteTemplate(w, "socialgroups_view_group.html", sgpi)
			if err != nil {
				LogError(err)
				return false
			}
			return true
		}
	}
	return false
}

func socialgroupsTrowAssign(args ...interface{}) interface{} {
	var forum = args[1].(*Forum)
	if forum.ParentType == "socialgroup" {
		var topicItem = args[0].(*TopicsRow)
		topicItem.ForumLink = "/group/" + strings.TrimPrefix(topicItem.ForumLink, getForumURLPrefix())
	}
	return nil
}

// TODO: It would be nice, if you could select one of the boards in the group from that drop-down rather than just the one you got linked from
func socialgroupsTopicCreatePreLoop(args ...interface{}) interface{} {
	var fid = args[2].(int)
	if fstore.DirtyGet(fid).ParentType == "socialgroup" {
		var strictmode = args[5].(*bool)
		*strictmode = true
	}
	return nil
}

// TODO: Add privacy options
// TODO: Add support for multiple boards and add per-board simplified permissions
// TODO: Take isJs into account for routes which expect JSON responses
func socialgroupsForumCheck(args ...interface{}) (skip bool, rerr RouteError) {
	var r = args[1].(*http.Request)
	var fid = args[3].(*int)
	var forum = fstore.DirtyGet(*fid)

	if forum.ParentType == "socialgroup" {
		var err error
		var w = args[0].(http.ResponseWriter)
		sgItem, ok := r.Context().Value("socialgroups_current_group").(*SocialGroup)
		if !ok {
			sgItem, err = socialgroupsGetGroup(forum.ParentID)
			if err != nil {
				return true, InternalError(errors.New("Unable to find the parent group for a forum"), w, r)
			}
			if !sgItem.Active {
				return true, NotFound(w, r)
			}
			r = r.WithContext(context.WithValue(r.Context(), "socialgroups_current_group", sgItem))
		}

		var user = args[2].(*User)
		var rank int
		var posts int
		var joinedAt string

		// TODO: Group privacy settings. For now, groups are all globally visible

		// Clear the default group permissions
		// TODO: Do this more efficiently, doing it quick and dirty for now to get this out quickly
		overrideForumPerms(&user.Perms, false)
		user.Perms.ViewTopic = true

		err = socialgroupsGetMemberStmt.QueryRow(sgItem.ID, user.ID).Scan(&rank, &posts, &joinedAt)
		if err != nil && err != ErrNoRows {
			return true, InternalError(err, w, r)
		} else if err != nil {
			// TODO: Should we let admins / guests into public groups?
			return true, LocalError("You're not part of this group!", w, r, *user)
		}

		// TODO: Implement bans properly by adding the Local Ban API in the next commit
		// TODO: How does this even work? Refactor it along with the rest of this plugin!
		if rank < 0 {
			return true, LocalError("You've been banned from this group!", w, r, *user)
		}

		// Basic permissions for members, more complicated permissions coming in the next commit!
		if sgItem.Owner == user.ID {
			overrideForumPerms(&user.Perms, true)
		} else if rank == 0 {
			user.Perms.LikeItem = true
			user.Perms.CreateTopic = true
			user.Perms.CreateReply = true
		} else {
			overrideForumPerms(&user.Perms, true)
		}
		return true, nil
	}

	return false, nil
}

// TODO: Override redirects? I don't think this is needed quite yet

func socialgroupsWidgets(args ...interface{}) interface{} {
	var zone = args[0].(string)
	var headerVars = args[2].(*HeaderVars)
	var request = args[3].(*http.Request)

	if zone != "view_forum" {
		return false
	}

	var forum = args[1].(*Forum)
	if forum.ParentType == "socialgroup" {
		// This is why I hate using contexts, all the daisy chains and interface casts x.x
		sgItem, ok := request.Context().Value("socialgroups_current_group").(*SocialGroup)
		if !ok {
			LogError(errors.New("Unable to find a parent group in the context data"))
			return false
		}

		if headerVars.ExtData.items == nil {
			headerVars.ExtData.items = make(map[string]interface{})
		}
		headerVars.ExtData.items["socialgroups_current_group"] = sgItem

		return socialgroupsGroupWidgets(headerVars, sgItem)
	}
	return false
}
