package guilds

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"html"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"../../../common"
)

var ListStmt *sql.Stmt
var MemberListStmt *sql.Stmt
var MemberListJoinStmt *sql.Stmt
var GetMemberStmt *sql.Stmt
var GetGuildStmt *sql.Stmt
var CreateGuildStmt *sql.Stmt
var AttachForumStmt *sql.Stmt
var UnattachForumStmt *sql.Stmt
var AddMemberStmt *sql.Stmt

// Guild is a struct representing a guild
type Guild struct {
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
	MainForum   *common.Forum
	Forums      []*common.Forum
	ExtData     ExtData
}

type Page struct {
	Title       string
	CurrentUser User
	Header      *common.HeaderVars
	ItemList    []*TopicsRow
	Forum       *commmon.Forum
	Guild       *Guild
	Page        int
	LastPage    int
}

// ListPage is a page struct for constructing a list of every guild
type ListPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	GuildList   []*Guild
}

type MemberListPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	ItemList    []GuildMember
	Guild       *Guild
	Page        int
	LastPage    int
}

// Member is a struct representing a specific member of a guild, not to be confused with the global User struct.
type Member struct {
	Link       string
	Rank       int    /* 0: Member. 1: Mod. 2: Admin. */
	RankString string /* Member, Mod, Admin, Owner */
	PostCount  int
	JoinedAt   string
	Offline    bool // TODO: Need to track the online states of members when WebSockets are enabled

	User User
}

func PrebuildTmplList(user *User, headerVars *HeaderVars) CTmpl {
	var guildList = []*Guild{
		&Guild{
			ID:             1,
			Name:           "lol",
			Link:           guildsBuildGuildURL(nameToSlug("lol"), 1),
			Desc:           "A group for people who like to laugh",
			Active:         true,
			MemberCount:    1,
			Owner:          1,
			CreatedAt:      "date",
			LastUpdateTime: "date",
			MainForumID:    1,
			MainForum:      fstore.DirtyGet(1),
			Forums:         []*Forum{fstore.DirtyGet(1)},
		},
	}
	listPage := ListPage{"Guild List", user, headerVars, guildList}
	return CTmpl{"guilds-guild-list", "guilds_guild_list", "templates/", "guilds.ListPage", listPage}
}

// TODO: Do this properly via the widget system
func CommonAreaWidgets(headerVars *HeaderVars) {
	// TODO: Hot Groups? Featured Groups? Official Groups?
	var b bytes.Buffer
	var menu = WidgetMenu{"Guilds", []WidgetMenuItem{
		WidgetMenuItem{"Create Guild", "/guild/create/", false},
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
func GuildWidgets(headerVars *HeaderVars, guildItem *Guild) (success bool) {
	return false // Disabled until the next commit

	/*var b bytes.Buffer
	var menu WidgetMenu = WidgetMenu{"Guild Options", []WidgetMenuItem{
		WidgetMenuItem{"Join", "/guild/join/" + strconv.Itoa(guildItem.ID), false},
		WidgetMenuItem{"Members", "/guild/members/" + strconv.Itoa(guildItem.ID), false},
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

func routeGuildList(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	guildsCommonAreaWidgets(headerVars)

	rows, err := guildsListStmt.Query()
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	var guildList []*Guild
	for rows.Next() {
		guildItem := &Guild{ID: 0}
		err := rows.Scan(&guildItem.ID, &guildItem.Name, &guildItem.Desc, &guildItem.Active, &guildItem.Privacy, &guildItem.Joinable, &guildItem.Owner, &guildItem.MemberCount, &guildItem.CreatedAt, &guildItem.LastUpdateTime)
		if err != nil {
			return InternalError(err, w, r)
		}
		guildItem.Link = guildsBuildGuildURL(nameToSlug(guildItem.Name), guildItem.ID)
		guildList = append(guildList, guildItem)
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}

	pi := GuildListPage{"Guild List", user, headerVars, guildList}
	err = RunThemeTemplate(headerVars.ThemeName, "guilds_guild_list", pi, w)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func GetGuild(guildID int) (guildItem *Guild, err error) {
	guildItem = &Guild{ID: guildID}
	err = guildsGetGuildStmt.QueryRow(guildID).Scan(&guildItem.Name, &guildItem.Desc, &guildItem.Active, &guildItem.Privacy, &guildItem.Joinable, &guildItem.Owner, &guildItem.MemberCount, &guildItem.MainForumID, &guildItem.Backdrop, &guildItem.CreatedAt, &guildItem.LastUpdateTime)
	return guildItem, err
}

func middleViewGuild(w http.ResponseWriter, r *http.Request, user User) RouteError {
	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/guild/"):], ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	guildID, err := strconv.Atoi(halves[1])
	if err != nil {
		return PreError("Not a valid guild ID", w, r)
	}

	guildItem, err := guildsGetGuild(guildID)
	if err != nil {
		return LocalError("Bad guild", w, r, user)
	}
	if !guildItem.Active {
		return NotFound(w, r)
	}

	// Re-route the request to routeForums
	var ctx = context.WithValue(r.Context(), "guilds_current_guild", guildItem)
	return routeForum(w, r.WithContext(ctx), user, strconv.Itoa(guildItem.MainForumID))
}

func CreateGuild(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	// TODO: Add an approval queue mode for group creation
	if !user.Loggedin || !user.PluginPerms["CreateGuild"] {
		return NoPermissions(w, r, user)
	}
	guildsCommonAreaWidgets(headerVars)

	pi := Page{"Create Guild", user, headerVars, tList, nil}
	err := templates.ExecuteTemplate(w, "guilds_create_guild.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func CreateGuildSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	// TODO: Add an approval queue mode for group creation
	if !user.Loggedin || !user.PluginPerms["CreateGuild"] {
		return NoPermissions(w, r, user)
	}

	var guildActive = true
	var guildName = html.EscapeString(r.PostFormValue("group_name"))
	var guildDesc = html.EscapeString(r.PostFormValue("group_desc"))
	var gprivacy = r.PostFormValue("group_privacy")

	var guildPrivacy int
	switch gprivacy {
	case "0":
		guildPrivacy = 0 // Public
	case "1":
		guildPrivacy = 1 // Protected
	case "2":
		guildPrivacy = 2 // private
	default:
		guildPrivacy = 0
	}

	// Create the backing forum
	fid, err := fstore.Create(guildName, "", true, "")
	if err != nil {
		return InternalError(err, w, r)
	}

	res, err := guildsCreateGuildStmt.Exec(guildName, guildDesc, guildActive, guildPrivacy, user.ID, fid)
	if err != nil {
		return InternalError(err, w, r)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return InternalError(err, w, r)
	}

	// Add the main backing forum to the forum list
	err = guildsAttachForum(int(lastID), fid)
	if err != nil {
		return InternalError(err, w, r)
	}

	_, err = guildsAddMemberStmt.Exec(lastID, user.ID, 2)
	if err != nil {
		return InternalError(err, w, r)
	}

	http.Redirect(w, r, guildsBuildGuildURL(nameToSlug(guildName), int(lastID)), http.StatusSeeOther)
	return nil
}

func MemberList(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/guild/members/"):], ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	guildID, err := strconv.Atoi(halves[1])
	if err != nil {
		return PreError("Not a valid group ID", w, r)
	}

	var guildItem = &Guild{ID: guildID}
	var mainForum int // Unused
	err = guildsGetGuildStmt.QueryRow(guildID).Scan(&guildItem.Name, &guildItem.Desc, &guildItem.Active, &guildItem.Privacy, &guildItem.Joinable, &guildItem.Owner, &guildItem.MemberCount, &mainForum, &guildItem.Backdrop, &guildItem.CreatedAt, &guildItem.LastUpdateTime)
	if err != nil {
		return LocalError("Bad group", w, r, user)
	}
	guildItem.Link = guildsBuildGuildURL(nameToSlug(guildItem.Name), guildItem.ID)

	guildsGuildWidgets(headerVars, guildItem)

	rows, err := guildsMemberListJoinStmt.Query(guildID)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}

	var guildMembers []GuildMember
	for rows.Next() {
		guildMember := GuildMember{PostCount: 0}
		err := rows.Scan(&guildMember.User.ID, &guildMember.Rank, &guildMember.PostCount, &guildMember.JoinedAt, &guildMember.User.Name, &guildMember.User.Avatar)
		if err != nil {
			return InternalError(err, w, r)
		}
		guildMember.Link = buildProfileURL(nameToSlug(guildMember.User.Name), guildMember.User.ID)
		if guildMember.User.Avatar != "" {
			if guildMember.User.Avatar[0] == '.' {
				guildMember.User.Avatar = "/uploads/avatar_" + strconv.Itoa(guildMember.User.ID) + guildMember.User.Avatar
			}
		} else {
			guildMember.User.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(guildMember.User.ID), 1)
		}
		guildMember.JoinedAt, _ = relativeTimeFromString(guildMember.JoinedAt)
		if guildItem.Owner == guildMember.User.ID {
			guildMember.RankString = "Owner"
		} else {
			switch guildMember.Rank {
			case 0:
				guildMember.RankString = "Member"
			case 1:
				guildMember.RankString = "Mod"
			case 2:
				guildMember.RankString = "Admin"
			}
		}
		guildMembers = append(guildMembers, guildMember)
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}
	rows.Close()

	pi := GuildMemberListPage{"Guild Member List", user, headerVars, guildMembers, guildItem, 0, 0}
	// A plugin with plugins. Pluginception!
	if preRenderHooks["pre_render_guilds_member_list"] != nil {
		if runPreRenderHook("pre_render_guilds_member_list", w, r, &user, &pi) {
			return nil
		}
	}
	err = RunThemeTemplate(headerVars.ThemeName, "guilds_member_list", pi, w)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func AttachForum(guildID int, fid int) error {
	_, err := guildsAttachForumStmt.Exec(guildID, fid)
	return err
}

func UnattachForum(fid int) error {
	_, err := guildsAttachForumStmt.Exec(fid)
	return err
}

func BuildGuildURL(slug string, id int) string {
	if slug == "" {
		return "/guild/" + slug + "." + strconv.Itoa(id)
	}
	return "/guild/" + strconv.Itoa(id)
}

/*
	Hooks
*/

// TODO: Prebuild this template
func PreRenderViewForum(w http.ResponseWriter, r *http.Request, user *User, data interface{}) (halt bool) {
	pi := data.(*ForumPage)
	if pi.Header.ExtData.items != nil {
		if guildData, ok := pi.Header.ExtData.items["guilds_current_group"]; ok {
			guildItem := guildData.(*Guild)

			guildpi := GuildPage{pi.Title, pi.CurrentUser, pi.Header, pi.ItemList, pi.Forum, guildItem, pi.Page, pi.LastPage}
			err := templates.ExecuteTemplate(w, "guilds_view_guild.html", guildpi)
			if err != nil {
				LogError(err)
				return false
			}
			return true
		}
	}
	return false
}

func TrowAssign(args ...interface{}) interface{} {
	var forum = args[1].(*Forum)
	if forum.ParentType == "guild" {
		var topicItem = args[0].(*TopicsRow)
		topicItem.ForumLink = "/guild/" + strings.TrimPrefix(topicItem.ForumLink, getForumURLPrefix())
	}
	return nil
}

// TODO: It would be nice, if you could select one of the boards in the group from that drop-down rather than just the one you got linked from
func TopicCreatePreLoop(args ...interface{}) interface{} {
	var fid = args[2].(int)
	if fstore.DirtyGet(fid).ParentType == "guild" {
		var strictmode = args[5].(*bool)
		*strictmode = true
	}
	return nil
}

// TODO: Add privacy options
// TODO: Add support for multiple boards and add per-board simplified permissions
// TODO: Take isJs into account for routes which expect JSON responses
func ForumCheck(args ...interface{}) (skip bool, rerr RouteError) {
	var r = args[1].(*http.Request)
	var fid = args[3].(*int)
	var forum = fstore.DirtyGet(*fid)

	if forum.ParentType == "guild" {
		var err error
		var w = args[0].(http.ResponseWriter)
		guildItem, ok := r.Context().Value("guilds_current_group").(*Guild)
		if !ok {
			guildItem, err = guildsGetGuild(forum.ParentID)
			if err != nil {
				return true, InternalError(errors.New("Unable to find the parent group for a forum"), w, r)
			}
			if !guildItem.Active {
				return true, NotFound(w, r)
			}
			r = r.WithContext(context.WithValue(r.Context(), "guilds_current_group", guildItem))
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

		err = guildsGetMemberStmt.QueryRow(guildItem.ID, user.ID).Scan(&rank, &posts, &joinedAt)
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
		if guildItem.Owner == user.ID {
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

func Widgets(args ...interface{}) interface{} {
	var zone = args[0].(string)
	var headerVars = args[2].(*HeaderVars)
	var request = args[3].(*http.Request)

	if zone != "view_forum" {
		return false
	}

	var forum = args[1].(*Forum)
	if forum.ParentType == "guild" {
		// This is why I hate using contexts, all the daisy chains and interface casts x.x
		guildItem, ok := request.Context().Value("guilds_current_group").(*Guild)
		if !ok {
			LogError(errors.New("Unable to find a parent group in the context data"))
			return false
		}

		if headerVars.ExtData.items == nil {
			headerVars.ExtData.items = make(map[string]interface{})
		}
		headerVars.ExtData.items["guilds_current_group"] = guildItem

		return guildsGuildWidgets(headerVars, guildItem)
	}
	return false
}
