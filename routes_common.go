package main

import (
	"html"
	"html/template"
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
var PanelUserCheck func(http.ResponseWriter, *http.Request, *User) (*HeaderVars, PanelStats, bool) = panelUserCheck
var SimplePanelUserCheck func(http.ResponseWriter, *http.Request, *User) (*HeaderLite, bool) = simplePanelUserCheck
var SimpleForumUserCheck func(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerLite *HeaderLite, success bool) = simpleForumUserCheck
var ForumUserCheck func(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerVars *HeaderVars, success bool) = forumUserCheck
var MemberCheck func(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, success bool) = memberCheck
var SimpleUserCheck func(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, success bool) = simpleUserCheck
var UserCheck func(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, success bool) = userCheck

// TODO: Support for left sidebars and sidebars on both sides
// http.Request is for context.Context middleware. Mostly for plugin_socialgroups right now
func BuildWidgets(zone string, data interface{}, headerVars *HeaderVars, r *http.Request) {
	if vhooks["intercept_build_widgets"] != nil {
		if runVhook("intercept_build_widgets", zone, data, headerVars, r).(bool) {
			return
		}
	}

	//log.Print("themes[headerVars.ThemeName].Sidebars",themes[headerVars.ThemeName].Sidebars)
	if themes[headerVars.ThemeName].Sidebars == "right" {
		if len(docks.RightSidebar) != 0 {
			var sbody string
			for _, widget := range docks.RightSidebar {
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

func simpleForumUserCheck(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerLite *HeaderLite, success bool) {
	if !fstore.Exists(fid) {
		PreError("The target forum doesn't exist.", w, r)
		return nil, false
	}
	success = true

	// Is there a better way of doing the skip AND the success flag on this hook like multiple returns?
	if vhooks["simple_forum_check_pre_perms"] != nil {
		if runVhook("simple_forum_check_pre_perms", w, r, user, &fid, &success, &headerLite).(bool) {
			return headerLite, success
		}
	}

	group, err := gstore.Get(user.Group)
	if err != nil {
		PreError("Something weird happened", w, r)
		log.Print("Group #" + strconv.Itoa(user.Group) + " doesn't exist despite being used by User #" + strconv.Itoa(user.ID))
		return
	}

	fperms := group.Forums[fid]
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
	return headerLite, true
}

func forumUserCheck(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerVars *HeaderVars, success bool) {
	headerVars, success = UserCheck(w, r, user)
	if !fstore.Exists(fid) {
		NotFound(w, r)
		return headerVars, false
	}

	if vhooks["forum_check_pre_perms"] != nil {
		if runVhook("forum_check_pre_perms", w, r, user, &fid, &success, &headerVars).(bool) {
			return headerVars, success
		}
	}

	group, err := gstore.Get(user.Group)
	if err != nil {
		PreError("Something weird happened", w, r)
		log.Print("Group #" + strconv.Itoa(user.Group) + " doesn't exist despite being used by User #" + strconv.Itoa(user.ID))
		return
	}

	fperms := group.Forums[fid]
	//log.Printf("user.Perms: %+v\n", user.Perms)
	//log.Printf("fperms: %+v\n", fperms)
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
	return headerVars, success
}

// Even if they have the right permissions, the control panel is only open to supermods+. There are many areas without subpermissions which assume that the current user is a supermod+ and admins are extremely unlikely to give these permissions to someone who isn't at-least a supermod to begin with
// TODO: Do a panel specific theme?
func panelUserCheck(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, stats PanelStats, success bool) {
	var themeName = defaultThemeBox.Load().(string)

	cookie, err := r.Cookie("current_theme")
	if err == nil {
		cookie := html.EscapeString(cookie.Value)
		theme, ok := themes[cookie]
		if ok && !theme.HideFromThemes {
			themeName = cookie
		}
	}

	headerVars = &HeaderVars{
		Site:      site,
		Settings:  settingBox.Load().(SettingBox),
		Themes:    themes,
		ThemeName: themeName,
	}
	// TODO: We should probably initialise headerVars.ExtData

	if !user.IsSuperMod {
		NoPermissions(w, r, *user)
		return headerVars, stats, false
	}

	headerVars.Stylesheets = append(headerVars.Stylesheets, headerVars.ThemeName+"/panel.css")
	if len(themes[headerVars.ThemeName].Resources) != 0 {
		rlist := themes[headerVars.ThemeName].Resources
		for _, resource := range rlist {
			if resource.Location == "global" || resource.Location == "panel" {
				halves := strings.Split(resource.Name, ".")
				if len(halves) != 2 {
					continue
				}
				if halves[1] == "css" {
					headerVars.Stylesheets = append(headerVars.Stylesheets, resource.Name)
				} else if halves[1] == "js" {
					headerVars.Scripts = append(headerVars.Scripts, resource.Name)
				}
			}
		}
	}

	err = groupCountStmt.QueryRow().Scan(&stats.Groups)
	if err != nil {
		InternalError(err, w)
		return headerVars, stats, false
	}

	stats.Users = users.GetGlobalCount()
	stats.Forums = fstore.GetGlobalCount() // TODO: Stop it from showing the blanked forums
	stats.Settings = len(headerVars.Settings)
	stats.WordFilters = len(wordFilterBox.Load().(WordFilterBox))
	stats.Themes = len(themes)
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

	return headerVars, stats, true
}

func simplePanelUserCheck(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, success bool) {
	if !user.IsSuperMod {
		NoPermissions(w, r, *user)
		return headerLite, false
	}
	headerLite = &HeaderLite{
		Site:     site,
		Settings: settingBox.Load().(SettingBox),
	}
	return headerLite, true
}

// TODO: Add this to the member routes
func memberCheck(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, success bool) {
	headerVars, success = UserCheck(w, r, user)
	if !user.Loggedin {
		NoPermissions(w, r, *user)
		return headerVars, false
	}
	return headerVars, success
}

// SimpleUserCheck is back from the grave, yay :D
func simpleUserCheck(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, success bool) {
	headerLite = &HeaderLite{
		Site:     site,
		Settings: settingBox.Load().(SettingBox),
	}
	return headerLite, true
}

// TODO: Add the ability for admins to restrict certain themes to certain groups?
func userCheck(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, success bool) {
	var themeName = defaultThemeBox.Load().(string)

	cookie, err := r.Cookie("current_theme")
	if err == nil {
		cookie := html.EscapeString(cookie.Value)
		theme, ok := themes[cookie]
		if ok && !theme.HideFromThemes {
			themeName = cookie
		}
	}

	headerVars = &HeaderVars{
		Site:      site,
		Settings:  settingBox.Load().(SettingBox),
		Themes:    themes,
		ThemeName: themeName,
	}

	if user.IsBanned {
		headerVars.NoticeList = append(headerVars.NoticeList, "Your account has been suspended. Some of your permissions may have been revoked.")
	}

	if len(themes[headerVars.ThemeName].Resources) != 0 {
		rlist := themes[headerVars.ThemeName].Resources
		for _, resource := range rlist {
			if resource.Location == "global" || resource.Location == "frontend" {
				halves := strings.Split(resource.Name, ".")
				if len(halves) != 2 {
					continue
				}
				if halves[1] == "css" {
					headerVars.Stylesheets = append(headerVars.Stylesheets, resource.Name)
				} else if halves[1] == "js" {
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

	return headerVars, true
}

func preRoute(w http.ResponseWriter, r *http.Request) (User, bool) {
	user, halt := auth.SessionCheck(w, r)
	if halt {
		return *user, false
	}
	if user == &guestUser {
		return *user, true
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		PreError("Bad IP", w, r)
		return *user, false
	}
	if host != user.LastIP {
		_, err = updateLastIPStmt.Exec(host, user.ID)
		if err != nil {
			InternalError(err, w)
			return *user, false
		}
		user.LastIP = host
	}

	h := w.Header()
	h.Set("X-Frame-Options", "deny")
	//h.Set("X-XSS-Protection", "1")
	// TODO: Set the content policy header
	return *user, true
}
