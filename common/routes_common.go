package common

import (
	"html"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// nolint
var PreRoute func(http.ResponseWriter, *http.Request) (User, bool) = preRoute

// TODO: Come up with a better middleware solution
// nolint We need these types so people can tell what they are without scrolling to the bottom of the file
var PanelUserCheck func(http.ResponseWriter, *http.Request, *User) (*HeaderVars, PanelStats, RouteError) = panelUserCheck
var SimplePanelUserCheck func(http.ResponseWriter, *http.Request, *User) (*HeaderLite, RouteError) = simplePanelUserCheck
var SimpleForumUserCheck func(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerLite *HeaderLite, err RouteError) = simpleForumUserCheck
var ForumUserCheck func(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerVars *HeaderVars, err RouteError) = forumUserCheck
var MemberCheck func(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, err RouteError) = memberCheck
var SimpleUserCheck func(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, err RouteError) = simpleUserCheck
var UserCheck func(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, err RouteError) = userCheck

func simpleForumUserCheck(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerLite *HeaderLite, rerr RouteError) {
	if !Forums.Exists(fid) {
		return nil, PreError("The target forum doesn't exist.", w, r)
	}

	// Is there a better way of doing the skip AND the success flag on this hook like multiple returns?
	if VhookSkippable["simple_forum_check_pre_perms"] != nil {
		var skip bool
		skip, rerr = RunVhookSkippable("simple_forum_check_pre_perms", w, r, user, &fid, &headerLite)
		if skip || rerr != nil {
			return headerLite, rerr
		}
	}

	fperms, err := FPStore.Get(fid, user.Group)
	if err != nil {
		// TODO: Refactor this
		log.Printf("Unable to get the forum perms for Group #%d for User #%d", user.Group, user.ID)
		return nil, PreError("Something weird happened", w, r)
	}
	cascadeForumPerms(fperms, user)
	return headerLite, nil
}

func forumUserCheck(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerVars *HeaderVars, rerr RouteError) {
	headerVars, rerr = UserCheck(w, r, user)
	if rerr != nil {
		return headerVars, rerr
	}
	if !Forums.Exists(fid) {
		return headerVars, NotFound(w, r, headerVars)
	}

	if VhookSkippable["forum_check_pre_perms"] != nil {
		var skip bool
		skip, rerr = RunVhookSkippable("forum_check_pre_perms", w, r, user, &fid, &headerVars)
		if skip || rerr != nil {
			return headerVars, rerr
		}
	}

	fperms, err := FPStore.Get(fid, user.Group)
	if err != nil {
		// TODO: Refactor this
		log.Printf("Unable to get the forum perms for Group #%d for User #%d", user.Group, user.ID)
		return nil, PreError("Something weird happened", w, r)
	}
	cascadeForumPerms(fperms, user)
	return headerVars, rerr
}

// TODO: Put this on the user instance? Do we really want forum specific logic in there? Maybe, a method which spits a new pointer with the same contents as user?
func cascadeForumPerms(fperms *ForumPerms, user *User) {
	if fperms.Overrides && !user.IsSuperAdmin {
		user.Perms.ViewTopic = fperms.ViewTopic
		user.Perms.LikeItem = fperms.LikeItem
		user.Perms.CreateTopic = fperms.CreateTopic
		user.Perms.EditTopic = fperms.EditTopic
		user.Perms.DeleteTopic = fperms.DeleteTopic
		user.Perms.CreateReply = fperms.CreateReply
		user.Perms.EditReply = fperms.EditReply
		user.Perms.DeleteReply = fperms.DeleteReply
		user.Perms.PinTopic = fperms.PinTopic
		user.Perms.CloseTopic = fperms.CloseTopic
		user.Perms.MoveTopic = fperms.MoveTopic

		if len(fperms.ExtData) != 0 {
			for name, perm := range fperms.ExtData {
				user.PluginPerms[name] = perm
			}
		}
	}
}

// Even if they have the right permissions, the control panel is only open to supermods+. There are many areas without subpermissions which assume that the current user is a supermod+ and admins are extremely unlikely to give these permissions to someone who isn't at-least a supermod to begin with
// TODO: Do a panel specific theme?
func panelUserCheck(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, stats PanelStats, rerr RouteError) {
	var theme = &Theme{Name: ""}

	cookie, err := r.Cookie("current_theme")
	if err == nil {
		inTheme, ok := Themes[html.EscapeString(cookie.Value)]
		if ok && !theme.HideFromThemes {
			theme = inTheme
		}
	}
	if theme.Name == "" {
		theme = Themes[DefaultThemeBox.Load().(string)]
	}

	headerVars = &HeaderVars{
		Site:     Site,
		Settings: SettingBox.Load().(SettingMap),
		Themes:   Themes,
		Theme:    theme,
		Zone:     "panel",
		Writer:   w,
	}
	// TODO: We should probably initialise headerVars.ExtData

	headerVars.Stylesheets = append(headerVars.Stylesheets, theme.Name+"/panel.css")
	if len(theme.Resources) > 0 {
		rlist := theme.Resources
		for _, resource := range rlist {
			if resource.Location == "global" || resource.Location == "panel" {
				extarr := strings.Split(resource.Name, ".")
				ext := extarr[len(extarr)-1]
				if ext == "css" {
					headerVars.Stylesheets = append(headerVars.Stylesheets, resource.Name)
				} else if ext == "js" {
					headerVars.Scripts = append(headerVars.Scripts, resource.Name)
				}
			}
		}
	}

	stats.Users = Users.GlobalCount()
	stats.Groups = Groups.GlobalCount()
	stats.Forums = Forums.GlobalCount() // TODO: Stop it from showing the blanked forums
	stats.Settings = len(headerVars.Settings)
	stats.WordFilters = len(WordFilterBox.Load().(WordFilterMap))
	stats.Themes = len(Themes)
	stats.Reports = 0 // TODO: Do the report count. Only show open threads?

	pusher, ok := w.(http.Pusher)
	if ok {
		pusher.Push("/static/"+theme.Name+"/main.css", nil)
		pusher.Push("/static/"+theme.Name+"/panel.css", nil)
		pusher.Push("/static/global.js", nil)
		pusher.Push("/static/jquery-3.1.1.min.js", nil)
		// TODO: Test these
		for _, sheet := range headerVars.Stylesheets {
			pusher.Push("/static/"+sheet, nil)
		}
		for _, script := range headerVars.Scripts {
			pusher.Push("/static/"+script, nil)
		}
		// TODO: Push avatars?
	}

	return headerVars, stats, nil
}

func simplePanelUserCheck(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, rerr RouteError) {
	return &HeaderLite{
		Site:     Site,
		Settings: SettingBox.Load().(SettingMap),
	}, nil
}

// TODO: Add this to the member routes
func memberCheck(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, rerr RouteError) {
	headerVars, rerr = UserCheck(w, r, user)
	if !user.Loggedin {
		return headerVars, NoPermissions(w, r, *user)
	}
	return headerVars, rerr
}

// SimpleUserCheck is back from the grave, yay :D
func simpleUserCheck(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, rerr RouteError) {
	headerLite = &HeaderLite{
		Site:     Site,
		Settings: SettingBox.Load().(SettingMap),
	}
	return headerLite, nil
}

// TODO: Add the ability for admins to restrict certain themes to certain groups?
func userCheck(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, rerr RouteError) {
	var theme = &Theme{Name: ""}

	cookie, err := r.Cookie("current_theme")
	if err == nil {
		inTheme, ok := Themes[html.EscapeString(cookie.Value)]
		if ok && !theme.HideFromThemes {
			theme = inTheme
		}
	}
	if theme.Name == "" {
		theme = Themes[DefaultThemeBox.Load().(string)]
	}

	headerVars = &HeaderVars{
		Site:     Site,
		Settings: SettingBox.Load().(SettingMap),
		Themes:   Themes,
		Theme:    theme,
		Zone:     "frontend",
		Writer:   w,
	}

	if user.IsBanned {
		headerVars.NoticeList = append(headerVars.NoticeList, GetNoticePhrase("account_banned"))
	}
	if user.Loggedin && !user.Active {
		headerVars.NoticeList = append(headerVars.NoticeList, GetNoticePhrase("account_inactive"))
	}

	if len(theme.Resources) > 0 {
		rlist := theme.Resources
		for _, resource := range rlist {
			if resource.Location == "global" || resource.Location == "frontend" {
				extarr := strings.Split(resource.Name, ".")
				ext := extarr[len(extarr)-1]
				if ext == "css" {
					headerVars.Stylesheets = append(headerVars.Stylesheets, resource.Name)
				} else if ext == "js" {
					headerVars.Scripts = append(headerVars.Scripts, resource.Name)
				}
			}
		}
	}

	pusher, ok := w.(http.Pusher)
	if ok {
		pusher.Push("/static/"+theme.Name+"/main.css", nil)
		pusher.Push("/static/global.js", nil)
		pusher.Push("/static/jquery-3.1.1.min.js", nil)
		// TODO: Test these
		for _, sheet := range headerVars.Stylesheets {
			pusher.Push("/static/"+sheet, nil)
		}
		for _, script := range headerVars.Scripts {
			pusher.Push("/static/"+script, nil)
		}
		// TODO: Push avatars?
	}

	return headerVars, nil
}

func preRoute(w http.ResponseWriter, r *http.Request) (User, bool) {
	user, halt := Auth.SessionCheck(w, r)
	if halt {
		return *user, false
	}
	if user == &GuestUser {
		return *user, true
	}

	var usercpy *User = BlankUser()
	*usercpy = *user

	// TODO: WIP. Refactor this to eliminate the unnecessary query
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		PreError("Bad IP", w, r)
		return *usercpy, false
	}
	if host != usercpy.LastIP {
		err = usercpy.UpdateIP(host)
		if err != nil {
			InternalError(err, w, r)
			return *usercpy, false
		}
		usercpy.LastIP = host
	}

	h := w.Header()
	h.Set("X-Frame-Options", "deny")
	h.Set("X-XSS-Protection", "1; mode=block") // TODO: Remove when we add a CSP? CSP's are horrendously glitchy things, tread with caution before removing
	// TODO: Set the content policy header

	return *usercpy, true
}

// SuperAdminOnly makes sure that only super admin can access certain critical panel routes
func SuperAdminOnly(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if !user.IsSuperAdmin {
		return NoPermissions(w, r, user)
	}
	return nil
}

// AdminOnly makes sure that only admins can access certain panel routes
func AdminOnly(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if !user.IsAdmin {
		return NoPermissions(w, r, user)
	}
	return nil
}

// SuperModeOnly makes sure that only super mods or higher can access the panel routes
func SuperModOnly(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if !user.IsSuperMod {
		return NoPermissions(w, r, user)
	}
	return nil
}

// MemberOnly makes sure that only logged in users can access this route
func MemberOnly(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if !user.Loggedin {
		return LoginRequired(w, r, user)
	}
	return nil
}

// NoBanned stops any banned users from accessing this route
func NoBanned(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if user.IsBanned {
		return Banned(w, r, user)
	}
	return nil
}

func ParseForm(w http.ResponseWriter, r *http.Request, user User) RouteError {
	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	return nil
}

func NoSessionMismatch(w http.ResponseWriter, r *http.Request, user User) RouteError {
	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}
	return nil
}

func ReqIsJson(r *http.Request) bool {
	return r.Header.Get("Content-type") == "application/json"
}

func HandleUploadRoute(w http.ResponseWriter, r *http.Request, user User, maxFileSize int) RouteError {
	// TODO: Reuse this code more
	if r.ContentLength > int64(maxFileSize) {
		size, unit := ConvertByteUnit(float64(maxFileSize))
		return CustomError("Your upload is too big. Your files need to be smaller than "+strconv.Itoa(int(size))+unit+".", http.StatusExpectationFailed, "Error", w, r, nil, user)
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxFileSize))

	err := r.ParseMultipartForm(int64(Megabyte))
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	return nil
}

func NoUploadSessionMismatch(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}
	return nil
}
