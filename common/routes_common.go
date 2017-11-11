package common

import (
	"html"
	"html/template"
	"log"
	"net"
	"net/http"
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

// TODO: Support for left sidebars and sidebars on both sides
// http.Request is for context.Context middleware. Mostly for plugin_guilds right now
func BuildWidgets(zone string, data interface{}, headerVars *HeaderVars, r *http.Request) {
	if Vhooks["intercept_build_widgets"] != nil {
		if RunVhook("intercept_build_widgets", zone, data, headerVars, r).(bool) {
			return
		}
	}

	//log.Print("Themes[headerVars.ThemeName].Sidebars", Themes[headerVars.ThemeName].Sidebars)
	if Themes[headerVars.ThemeName].Sidebars == "right" {
		if len(Docks.RightSidebar) != 0 {
			var sbody string
			for _, widget := range Docks.RightSidebar {
				if widget.Enabled {
					if widget.Location == "global" || widget.Location == zone {
						sbody += widget.Body
					}
				}
			}
			headerVars.Widgets.RightSidebar = template.HTML(sbody)
		}
	}
}

func simpleForumUserCheck(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerLite *HeaderLite, rerr RouteError) {
	if !Fstore.Exists(fid) {
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

	fperms, err := Fpstore.Get(fid, user.Group)
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
	if !Fstore.Exists(fid) {
		return headerVars, NotFound(w, r)
	}

	if VhookSkippable["forum_check_pre_perms"] != nil {
		var skip bool
		skip, rerr = RunVhookSkippable("forum_check_pre_perms", w, r, user, &fid, &headerVars)
		if skip || rerr != nil {
			return headerVars, rerr
		}
	}

	fperms, err := Fpstore.Get(fid, user.Group)
	if err != nil {
		// TODO: Refactor this
		log.Printf("Unable to get the forum perms for Group #%d for User #%d", user.Group, user.ID)
		return nil, PreError("Something weird happened", w, r)
	}
	//log.Printf("user.Perms: %+v\n", user.Perms)
	//log.Printf("fperms: %+v\n", fperms)
	cascadeForumPerms(fperms, user)
	return headerVars, rerr
}

// TODO: Put this on the user instance? Do we really want forum specific logic in there? Maybe, a method which spits a new pointer with the same contents as user?
func cascadeForumPerms(fperms ForumPerms, user *User) {
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
	var themeName = DefaultThemeBox.Load().(string)

	cookie, err := r.Cookie("current_theme")
	if err == nil {
		cookie := html.EscapeString(cookie.Value)
		theme, ok := Themes[cookie]
		if ok && !theme.HideFromThemes {
			themeName = cookie
		}
	}

	headerVars = &HeaderVars{
		Site:      Site,
		Settings:  SettingBox.Load().(SettingMap),
		Themes:    Themes,
		ThemeName: themeName,
	}
	// TODO: We should probably initialise headerVars.ExtData

	headerVars.Stylesheets = append(headerVars.Stylesheets, headerVars.ThemeName+"/panel.css")
	if len(Themes[headerVars.ThemeName].Resources) > 0 {
		rlist := Themes[headerVars.ThemeName].Resources
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
	stats.Groups = Gstore.GlobalCount()
	stats.Forums = Fstore.GlobalCount() // TODO: Stop it from showing the blanked forums
	stats.Settings = len(headerVars.Settings)
	stats.WordFilters = len(WordFilterBox.Load().(WordFilterMap))
	stats.Themes = len(Themes)
	stats.Reports = 0 // TODO: Do the report count. Only show open threads?

	pusher, ok := w.(http.Pusher)
	if ok {
		pusher.Push("/static/"+headerVars.ThemeName+"/main.css", nil)
		pusher.Push("/static/"+headerVars.ThemeName+"/panel.css", nil)
		pusher.Push("/static/global.js", nil)
		pusher.Push("/static/jquery-3.1.1.min.js", nil)
		// TODO: Push the theme CSS files
		// TODO: Push the theme scripts
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
	var themeName = DefaultThemeBox.Load().(string)

	cookie, err := r.Cookie("current_theme")
	if err == nil {
		cookie := html.EscapeString(cookie.Value)
		theme, ok := Themes[cookie]
		if ok && !theme.HideFromThemes {
			themeName = cookie
		}
	}

	headerVars = &HeaderVars{
		Site:      Site,
		Settings:  SettingBox.Load().(SettingMap),
		Themes:    Themes,
		ThemeName: themeName,
	}

	if user.IsBanned {
		headerVars.NoticeList = append(headerVars.NoticeList, "Your account has been suspended. Some of your permissions may have been revoked.")
	}

	if len(Themes[headerVars.ThemeName].Resources) > 0 {
		rlist := Themes[headerVars.ThemeName].Resources
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
		pusher.Push("/static/"+headerVars.ThemeName+"/main.css", nil)
		pusher.Push("/static/global.js", nil)
		pusher.Push("/static/jquery-3.1.1.min.js", nil)
		// TODO: Push the theme CSS files
		// TODO: Push the theme scripts
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

	var usercpy *User
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
	//h.Set("X-XSS-Protection", "1")
	// TODO: Set the content policy header

	return *usercpy, true
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
