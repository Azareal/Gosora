package main

import (
	//"fmt"
	"bytes"
	"strings"
	"strconv"
	"errors"
	"context"
	"net/http"
	"html"
	"html/template"
	"database/sql"

	"./query_gen/lib"
)

var socialgroups_list_stmt *sql.Stmt
var socialgroups_member_list_stmt *sql.Stmt
var socialgroups_member_list_join_stmt *sql.Stmt
var socialgroups_get_member_stmt *sql.Stmt
var socialgroups_get_group_stmt *sql.Stmt
var socialgroups_create_group_stmt *sql.Stmt
var socialgroups_attach_forum_stmt *sql.Stmt
var socialgroups_unattach_forum_stmt *sql.Stmt
var socialgroups_add_member_stmt *sql.Stmt

// TO-DO: Add a better way of splitting up giant plugins like this
type SocialGroup struct
{
	ID int
	Link string
	Name string
	Desc string
	Active bool
	Privacy int /* 0: Public, 1: Protected, 2: Private */
	MemberCount int
	Owner int
	Backdrop string
	CreatedAt string
	LastUpdateTime string

	MainForum *Forum
	Forums []*Forum
	ExtData ExtData
}

type SocialGroupPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ItemList []TopicUser
	Forum Forum
	SocialGroup SocialGroup
	Page int
	LastPage int
	ExtData ExtData
}

type SocialGroupListPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	GroupList []SocialGroup
	ExtData ExtData
}

type SocialGroupMemberListPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ItemList []SocialGroupMember
	SocialGroup SocialGroup
	Page int
	LastPage int
	ExtData ExtData
}

type SocialGroupMember struct
{
	Link string
	Rank int /* 0: Member. 1: Mod. 2: Admin. */
	RankString string /* Member, Mod, Admin, Owner */
	PostCount int
	JoinedAt string
	Offline bool // TO-DO: Need to track the online states of members when WebSockets are enabled

	User User
}

func init() {
	plugins["socialgroups"] = NewPlugin("socialgroups","Social Groups","Azareal","http://github.com/Azareal","","","",init_socialgroups,nil,deactivate_socialgroups,install_socialgroups,nil)
}

func init_socialgroups() (err error) {
  plugins["socialgroups"].AddHook("intercept_build_widgets", socialgroups_widgets)
	plugins["socialgroups"].AddHook("trow_assign", socialgroups_trow_assign)
	plugins["socialgroups"].AddHook("topic_create_pre_loop", socialgroups_topic_create_pre_loop)
	plugins["socialgroups"].AddHook("pre_render_view_forum", socialgroups_pre_render_view_forum)
	plugins["socialgroups"].AddHook("simple_forum_check_pre_perms", socialgroups_forum_check)
	plugins["socialgroups"].AddHook("forum_check_pre_perms", socialgroups_forum_check)
	// TO-DO: Auto-grant this perm to admins upon installation?
	register_plugin_perm("CreateSocialGroup")
	router.HandleFunc("/groups/", socialgroups_group_list)
	router.HandleFunc("/group/", socialgroups_view_group)
	router.HandleFunc("/group/create/", socialgroups_create_group)
	router.HandleFunc("/group/create/submit/", socialgroups_create_group_submit)
	router.HandleFunc("/group/members/", socialgroups_member_list)

	socialgroups_list_stmt, err = qgen.Builder.SimpleSelect("socialgroups","sgid, name, desc, active, privacy, owner, memberCount, createdAt, lastUpdateTime","","","")
	if err != nil {
		return err
	}
	socialgroups_get_group_stmt, err = qgen.Builder.SimpleSelect("socialgroups","name, desc, active, privacy, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime","sgid = ?","","")
	if err != nil {
		return err
	}
	socialgroups_member_list_stmt, err = qgen.Builder.SimpleSelect("socialgroups_members","sgid, uid, rank, posts, joinedAt","","","")
	if err != nil {
		return err
	}
	socialgroups_member_list_join_stmt, err = qgen.Builder.SimpleLeftJoin("socialgroups_members","users","users.uid, socialgroups_members.rank, socialgroups_members.posts, socialgroups_members.joinedAt, users.name, users.avatar","socialgroups_members.uid = users.uid","socialgroups_members.sgid = ?","socialgroups_members.rank DESC, socialgroups_members.joinedat ASC","")
	if err != nil {
		return err
	}
	socialgroups_get_member_stmt, err = qgen.Builder.SimpleSelect("socialgroups_members","rank, posts, joinedAt","sgid = ? AND uid = ?","","")
	if err != nil {
		return err
	}
	socialgroups_create_group_stmt, err = qgen.Builder.SimpleInsert("socialgroups","name, desc, active, privacy, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime","?,?,?,?,?,1,?,'',NOW(),NOW()")
	if err != nil {
		return err
	}
	socialgroups_attach_forum_stmt, err = qgen.Builder.SimpleUpdate("forums","parentID = ?, parentType = 'socialgroup'","fid = ?")
	if err != nil {
		return err
	}
	socialgroups_unattach_forum_stmt, err = qgen.Builder.SimpleUpdate("forums","parentID = 0, parentType = ''","fid = ?")
	if err != nil {
		return err
	}
	socialgroups_add_member_stmt, err = qgen.Builder.SimpleInsert("socialgroups_members","sgid, uid, rank, posts, joinedAt","?,?,?,0,NOW()")
	if err != nil {
		return err
	}

	return nil
}

func deactivate_socialgroups() {
	plugins["socialgroups"].RemoveHook("intercept_build_widgets", socialgroups_widgets)
	plugins["socialgroups"].RemoveHook("trow_assign", socialgroups_trow_assign)
	plugins["socialgroups"].RemoveHook("topic_create_pre_loop", socialgroups_topic_create_pre_loop)
	plugins["socialgroups"].RemoveHook("pre_render_view_forum", socialgroups_pre_render_view_forum)
	plugins["socialgroups"].RemoveHook("simple_forum_check_pre_perms", socialgroups_forum_check)
	plugins["socialgroups"].RemoveHook("forum_check_pre_perms", socialgroups_forum_check)
	deregister_plugin_perm("CreateSocialGroup")
	_ = router.RemoveFunc("/groups/")
	_ = router.RemoveFunc("/group/")
	_ = router.RemoveFunc("/group/create/")
	_ = router.RemoveFunc("/group/create/submit/")
	_ = socialgroups_list_stmt.Close()
	_ = socialgroups_member_list_stmt.Close()
	_ = socialgroups_member_list_join_stmt.Close()
	_ = socialgroups_get_member_stmt.Close()
	_ = socialgroups_get_group_stmt.Close()
	_ = socialgroups_create_group_stmt.Close()
	_ = socialgroups_attach_forum_stmt.Close()
	_ = socialgroups_unattach_forum_stmt.Close()
	_ = socialgroups_add_member_stmt.Close()
}

// TO-DO: Stop accessing the query builder directly and add a feature in Gosora which is more easily reversed, if an error comes up during the installation process
func install_socialgroups() error {
	sg_table_stmt, err := qgen.Builder.CreateTable("socialgroups","utf8mb4","utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"sgid","int",0,false,true,""},
			qgen.DB_Table_Column{"name","varchar",100,false,false,""},
			qgen.DB_Table_Column{"desc","varchar",200,false,false,""},
			qgen.DB_Table_Column{"active","tinyint",1,false,false,""},
			qgen.DB_Table_Column{"privacy","tinyint",1,false,false,""},
			qgen.DB_Table_Column{"owner","int",0,false,false,""},
			qgen.DB_Table_Column{"memberCount","int",0,false,false,""},
			qgen.DB_Table_Column{"mainForum","int",0,false,false,"0"}, // The board the user lands on when they click ona group, we'll make it possible for group admins to change what users land on
			//qgen.DB_Table_Column{"boards","varchar",200,false,false,""}, // Cap the max number of boards at 8 to avoid overflowing the confines of a 64-bit integer?
			qgen.DB_Table_Column{"backdrop","varchar",200,false,false,""}, // File extension for the uploaded file, or an external link
			qgen.DB_Table_Column{"createdAt","createdAt",0,false,false,""},
			qgen.DB_Table_Column{"lastUpdateTime","datetime",0,false,false,""},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"sgid","primary"},
		},
	)
	if err != nil {
		return err
	}

	_, err = sg_table_stmt.Exec()
	if err != nil {
		return err
	}

	sg_members_table_stmt, err := qgen.Builder.CreateTable("socialgroups_members","","",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"sgid","int",0,false,false,""},
			qgen.DB_Table_Column{"uid","int",0,false,false,""},
			qgen.DB_Table_Column{"rank","int",0,false,false,"0"}, /* 0: Member. 1: Mod. 2: Admin. */
			qgen.DB_Table_Column{"posts","int",0,false,false,"0"}, /* Per-Group post count. Should we do some sort of score system? */
			qgen.DB_Table_Column{"joinedAt","datetime",0,false,false,""},
		},
		[]qgen.DB_Table_Key{},
	)
  if err != nil {
    return err
  }

  _, err = sg_members_table_stmt.Exec()
  return err
}

// TO-DO; Implement an uninstallation system into Gosora. And a better installation system.
func uninstall_socialgroups() error {
	return nil
}

// TO-DO: Do this properly via the widget system
func socialgroups_common_area_widgets(headerVars *HeaderVars) {
	// TO-DO: Hot Groups? Featured Groups? Official Groups?
	var b bytes.Buffer
	var menu WidgetMenu = WidgetMenu{"Social Groups",[]WidgetMenuItem{
		WidgetMenuItem{"Create Group","/group/create/",false},
	}}

	err := templates.ExecuteTemplate(&b,"widget_menu.html",menu)
	if err != nil {
		LogError(err)
		return
	}

	if themes[defaultTheme].Sidebars == "left" {
		headerVars.Widgets.LeftSidebar = template.HTML(string(b.Bytes()))
	} else if themes[defaultTheme].Sidebars == "right" || themes[defaultTheme].Sidebars == "both" {
		headerVars.Widgets.RightSidebar = template.HTML(string(b.Bytes()))
	}
}

// TO-DO: Do this properly via the widget system
// TO-DO: Make a better more customisable group widget system
func socialgroups_group_widgets(headerVars *HeaderVars, sgItem SocialGroup) (success bool) {
	return false // Disabled until the next commit

	var b bytes.Buffer
	var menu WidgetMenu = WidgetMenu{"Group Options",[]WidgetMenuItem{
		WidgetMenuItem{"Join","/group/join/" + strconv.Itoa(sgItem.ID),false},
		WidgetMenuItem{"Members","/group/members/" + strconv.Itoa(sgItem.ID),false},
	}}

	err := templates.ExecuteTemplate(&b,"widget_menu.html",menu)
	if err != nil {
		LogError(err)
		return false
	}

	if themes[defaultTheme].Sidebars == "left" {
		headerVars.Widgets.LeftSidebar = template.HTML(string(b.Bytes()))
	} else if themes[defaultTheme].Sidebars == "right" || themes[defaultTheme].Sidebars == "both" {
		headerVars.Widgets.RightSidebar = template.HTML(string(b.Bytes()))
	} else {
		return false
	}
	return true
}

/*
	Custom Pages
*/

func socialgroups_group_list(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	socialgroups_common_area_widgets(&headerVars)

	rows, err := socialgroups_list_stmt.Query()
  if err != nil && err != ErrNoRows {
		InternalError(err,w,r)
		return
	}

	var sgList []SocialGroup
	for rows.Next() {
		sgItem := SocialGroup{ID:0}
		err := rows.Scan(&sgItem.ID, &sgItem.Name, &sgItem.Desc, &sgItem.Active, &sgItem.Privacy, &sgItem.Owner, &sgItem.MemberCount, &sgItem.CreatedAt, &sgItem.LastUpdateTime)
    if err != nil {
      InternalError(err,w,r)
      return
    }
		sgItem.Link = socialgroups_build_group_url(name_to_slug(sgItem.Name),sgItem.ID)
  	sgList = append(sgList,sgItem)
  }
  err = rows.Err()
  if err != nil {
    InternalError(err,w,r)
    return
  }
  rows.Close()

	pi := SocialGroupListPage{"Group List",user,headerVars,sgList,extData}
	err = templates.ExecuteTemplate(w,"socialgroups_group_list.html", pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func socialgroups_view_group(w http.ResponseWriter, r *http.Request, user User) {
	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/group/"):],".")
	if len(halves) < 2 {
		halves = append(halves,halves[0])
	}
	sgid, err := strconv.Atoi(halves[1])
	if err != nil {
		PreError("Not a valid group ID",w,r)
		return
	}

	var sgItem SocialGroup = SocialGroup{ID:sgid}
	var mainForum int
	err = socialgroups_get_group_stmt.QueryRow(sgid).Scan(&sgItem.Name, &sgItem.Desc, &sgItem.Active, &sgItem.Privacy, &sgItem.Owner, &sgItem.MemberCount, &mainForum, &sgItem.Backdrop, &sgItem.CreatedAt, &sgItem.LastUpdateTime)
	if err != nil {
		LocalError("Bad group",w,r,user)
		return
	}
	if !sgItem.Active {
		NotFound(w,r)
	}

	// Re-route the request to route_forums
	var ctx context.Context = context.WithValue(r.Context(),"socialgroups_current_group",sgItem)
	route_forum(w,r.WithContext(ctx),user,strconv.Itoa(mainForum))
}

func socialgroups_create_group(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	// TO-DO: Add an approval queue mode for group creation
	if !user.Loggedin || !user.PluginPerms["CreateSocialGroup"] {
		NoPermissions(w,r,user)
		return
	}
	socialgroups_common_area_widgets(&headerVars)

	pi := Page{"Create Group",user,headerVars,tList,nil}
	err := templates.ExecuteTemplate(w,"socialgroups_create_group.html", pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func socialgroups_create_group_submit(w http.ResponseWriter, r *http.Request, user User) {
	// TO-DO: Add an approval queue mode for group creation
	if !user.Loggedin || !user.PluginPerms["CreateSocialGroup"] {
		NoPermissions(w,r,user)
		return
	}

	var group_active bool = true
	var group_name string = html.EscapeString(r.PostFormValue("group_name"))
	var group_desc string = html.EscapeString(r.PostFormValue("group_desc"))
	var gprivacy string = r.PostFormValue("group_privacy")

	var group_privacy int
	switch(gprivacy) {
		case "0": group_privacy = 0 // Public
		case "1": group_privacy = 1 // Protected
		case "2": group_privacy = 2 // private
		default: group_privacy = 0
	}

	// Create the backing forum
	fid, err := fstore.CreateForum(group_name,"",true,"")
	if err != nil {
		InternalError(err,w,r)
		return
	}

	res, err := socialgroups_create_group_stmt.Exec(group_name, group_desc, group_active, group_privacy, user.ID, fid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	// Add the main backing forum to the forum list
	err = socialgroups_attach_forum(int(lastId),fid)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	_, err = socialgroups_add_member_stmt.Exec(lastId,user.ID,2)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	http.Redirect(w,r,socialgroups_build_group_url(name_to_slug(group_name),int(lastId)), http.StatusSeeOther)
}

func socialgroups_member_list(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}

	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/group/members/"):],".")
	if len(halves) < 2 {
		halves = append(halves,halves[0])
	}
	sgid, err := strconv.Atoi(halves[1])
	if err != nil {
		PreError("Not a valid group ID",w,r)
		return
	}

	var sgItem SocialGroup = SocialGroup{ID:sgid}
	var mainForum int // Unused
	err = socialgroups_get_group_stmt.QueryRow(sgid).Scan(&sgItem.Name, &sgItem.Desc, &sgItem.Active, &sgItem.Privacy, &sgItem.Owner, &sgItem.MemberCount, &mainForum, &sgItem.Backdrop, &sgItem.CreatedAt, &sgItem.LastUpdateTime)
	if err != nil {
		LocalError("Bad group",w,r,user)
		return
	}
	sgItem.Link = socialgroups_build_group_url(name_to_slug(sgItem.Name),sgItem.ID)

	socialgroups_group_widgets(&headerVars, sgItem)

	rows, err := socialgroups_member_list_join_stmt.Query(sgid)
  if err != nil && err != ErrNoRows {
		InternalError(err,w,r)
		return
	}

	var sgMembers []SocialGroupMember
	for rows.Next() {
		sgMember := SocialGroupMember{PostCount:0}
		err := rows.Scan(&sgMember.User.ID,&sgMember.Rank,&sgMember.PostCount,&sgMember.JoinedAt,&sgMember.User.Name, &sgMember.User.Avatar)
    if err != nil {
      InternalError(err,w,r)
      return
    }
		sgMember.Link = build_profile_url(name_to_slug(sgMember.User.Name),sgMember.User.ID)
		if sgMember.User.Avatar != "" {
			if sgMember.User.Avatar[0] == '.' {
				sgMember.User.Avatar = "/uploads/avatar_" + strconv.Itoa(sgMember.User.ID) + sgMember.User.Avatar
			}
		} else {
			sgMember.User.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(sgMember.User.ID),1)
		}
		sgMember.JoinedAt, _ = relative_time(sgMember.JoinedAt)
		if sgItem.Owner == sgMember.User.ID {
			sgMember.RankString = "Owner"
		} else {
			switch(sgMember.Rank) {
				case 0: sgMember.RankString = "Member"
				case 1: sgMember.RankString = "Mod"
				case 2: sgMember.RankString = "Admin"
			}
		}
  	sgMembers = append(sgMembers,sgMember)
  }
  err = rows.Err()
  if err != nil {
    InternalError(err,w,r)
    return
  }
  rows.Close()

	pi := SocialGroupMemberListPage{"Group Member List",user,headerVars,sgMembers,sgItem,0,0,extData}
	// A plugin with plugins. Pluginception!
	if pre_render_hooks["pre_render_socialgroups_member_list"] != nil {
		if run_pre_render_hook("pre_render_socialgroups_member_list", w, r, &user, &pi) {
			return
		}
	}
	err = templates.ExecuteTemplate(w,"socialgroups_member_list.html", pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func socialgroups_attach_forum(sgid int, fid int) error {
	_, err := socialgroups_attach_forum_stmt.Exec(sgid,fid)
	return err
}

func socialgroups_unattach_forum(fid int) error {
	_, err := socialgroups_attach_forum_stmt.Exec(fid)
	return err
}

func socialgroups_build_group_url(slug string, id int) string {
	if slug == "" {
		return "/group/" + slug + "." + strconv.Itoa(id)
	}
	return "/group/" + strconv.Itoa(id)
}

/*
	Hooks
*/

func socialgroups_pre_render_view_forum(w http.ResponseWriter, r *http.Request, user *User, data interface{}) (halt bool) {
	pi := data.(*ForumPage)
	if pi.Header.ExtData.items != nil {
		 if sgData, ok := pi.Header.ExtData.items["socialgroups_current_group"]; ok {
			sgItem := sgData.(SocialGroup)

			sgpi := SocialGroupPage{pi.Title,pi.CurrentUser,pi.Header,pi.ItemList,pi.Forum,sgItem,pi.Page,pi.LastPage,pi.ExtData}
			err := templates.ExecuteTemplate(w,"socialgroups_view_group.html", sgpi)
			if err != nil {
				LogError(err)
				return false
			}
			return true
		 }
	}
	return false
}

func socialgroups_trow_assign(args ...interface{}) interface{} {
	var forum *Forum = args[1].(*Forum)
	if forum.ParentType == "socialgroup" {
		var topicItem *TopicsRow = args[0].(*TopicsRow)
		topicItem.ForumLink = "/group/" + strings.TrimPrefix(topicItem.ForumLink,get_forum_url_prefix())
	}
	return nil
}

// TO-DO: It would be nice, if you could select one of the boards in the group from that drop-down rather than just the one you got linked from
func socialgroups_topic_create_pre_loop(args ...interface{}) interface{} {
	var fid int = args[2].(int)
	if fstore.DirtyGet(fid).ParentType == "socialgroup" {
		var strictmode *bool = args[5].(*bool)
		*strictmode = true
	}
	return nil
}

// TO-DO: Permissions Override. It doesn't quite work yet.
func socialgroups_forum_check(args ...interface{}) (skip interface{}) {
	var r = args[1].(*http.Request)
	var fid *int = args[3].(*int)
	if fstore.DirtyGet(*fid).ParentType == "socialgroup" {
		sgItem, ok := r.Context().Value("socialgroups_current_group").(SocialGroup)
		if !ok {
			LogError(errors.New("Unable to find a parent group in the context data"))
			return false
		}

		//run_vhook("simple_forum_check_pre_perms", w, r, user, &fid, &success).(bool)
		var w = args[0].(http.ResponseWriter)
		var user *User = args[2].(*User)
		var success *bool = args[4].(*bool)
		var rank int
		var posts int
		var joinedAt string

		// TO-DO: Group privacy settings. For now, groups are all globally visible

		// Clear the default group permissions
		// TO-DO: Do this more efficiently, doing it quick and dirty for now to get this out quickly
		override_forum_perms(&user.Perms, false)
		user.Perms.ViewTopic = true

		err := socialgroups_get_member_stmt.QueryRow(sgItem.ID,user.ID).Scan(&rank,&posts,&joinedAt)
		if err != nil && err != ErrNoRows {
			*success = false
			InternalError(err,w,r)
			return false
		} else if err != nil {
			return false
		}

		// TO-DO: Implement bans properly by adding the Local Ban API in the next commit
		if rank < 0 {
			return false
		}

		// Basic permissions for members, more complicated permissions coming in the next commit!
		if sgItem.Owner == user.ID {
			override_forum_perms(&user.Perms,true)
		} else if rank == 0 {
			user.Perms.LikeItem = true
			user.Perms.CreateTopic = true
			user.Perms.CreateReply = true
		} else {
			override_forum_perms(&user.Perms,true)
		}
	}

	return false
}

// TO-DO: Override redirects? I don't think this is needed quite yet

func socialgroups_widgets(args ...interface{}) interface{} {
	var zone string = args[0].(string)
	var headerVars *HeaderVars = args[2].(*HeaderVars)
	var request *http.Request = args[3].(*http.Request)

	if zone != "view_forum" {
		return false
	}

	var forum *Forum = args[1].(*Forum)
	if forum.ParentType == "socialgroup" {
		// This is why I hate using contexts, all the daisy chains and interface casts x.x
		sgItem, ok := request.Context().Value("socialgroups_current_group").(SocialGroup)
		if !ok {
			LogError(errors.New("Unable to find a parent group in the context data"))
			return false
		}

		if headerVars.ExtData.items == nil {
			headerVars.ExtData.items = make(map[string]interface{})
		}
		headerVars.ExtData.items["socialgroups_current_group"] = sgItem

		return socialgroups_group_widgets(headerVars,sgItem)
	}
	return false
}
