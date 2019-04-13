package common

import (
	"html"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Azareal/Gosora/common/phrases"
)

// nolint
var PreRoute func(http.ResponseWriter, *http.Request) (User, bool) = preRoute

// TODO: Come up with a better middleware solution
// nolint We need these types so people can tell what they are without scrolling to the bottom of the file
var PanelUserCheck func(http.ResponseWriter, *http.Request, *User) (*Header, PanelStats, RouteError) = panelUserCheck
var SimplePanelUserCheck func(http.ResponseWriter, *http.Request, *User) (*HeaderLite, RouteError) = simplePanelUserCheck
var SimpleForumUserCheck func(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerLite *HeaderLite, err RouteError) = simpleForumUserCheck
var ForumUserCheck func(header *Header, w http.ResponseWriter, r *http.Request, user *User, fid int) (err RouteError) = forumUserCheck
var SimpleUserCheck func(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, err RouteError) = simpleUserCheck
var UserCheck func(w http.ResponseWriter, r *http.Request, user *User) (header *Header, err RouteError) = userCheck

func simpleForumUserCheck(w http.ResponseWriter, r *http.Request, user *User, fid int) (header *HeaderLite, rerr RouteError) {
	header, rerr = SimpleUserCheck(w, r, user)
	if rerr != nil {
		return header, rerr
	}
	if !Forums.Exists(fid) {
		return nil, PreError("The target forum doesn't exist.", w, r)
	}

	// Is there a better way of doing the skip AND the success flag on this hook like multiple returns?
	skip, rerr := header.Hooks.VhookSkippable("simple_forum_check_pre_perms", w, r, user, &fid, &header)
	if skip || rerr != nil {
		return header, rerr
	}

	fperms, err := FPStore.Get(fid, user.Group)
	if err == ErrNoRows {
		fperms = BlankForumPerms()
	} else if err != nil {
		return header, InternalError(err, w, r)
	}
	cascadeForumPerms(fperms, user)
	return header, nil
}

func forumUserCheck(header *Header, w http.ResponseWriter, r *http.Request, user *User, fid int) (rerr RouteError) {
	if !Forums.Exists(fid) {
		return NotFound(w, r, header)
	}

	skip, rerr := header.Hooks.VhookSkippable("forum_check_pre_perms", w, r, user, &fid, &header)
	if skip || rerr != nil {
		return rerr
	}

	fperms, err := FPStore.Get(fid, user.Group)
	if err == ErrNoRows {
		fperms = BlankForumPerms()
	} else if err != nil {
		return InternalError(err, w, r)
	}
	cascadeForumPerms(fperms, user)
	header.CurrentUser = *user // TODO: Use a pointer instead for CurrentUser, so we don't have to do this
	return rerr
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
func panelUserCheck(w http.ResponseWriter, r *http.Request, user *User) (header *Header, stats PanelStats, rerr RouteError) {
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

	header = &Header{
		Site:        Site,
		Settings:    SettingBox.Load().(SettingMap),
		Themes:      Themes,
		Theme:       theme,
		CurrentUser: *user,
		Hooks:       GetHookTable(),
		Zone:        "panel",
		Writer:      w,
		IsoCode:     phrases.GetLangPack().IsoCode,
	}
	// TODO: We should probably initialise header.ExtData
	// ? - Should we only show this in debug mode? It might be useful for detecting issues in production, if we show it there as-well
	if user.IsAdmin {
		header.StartedAt = time.Now()
	}

	header.AddSheet(theme.Name + "/main.css")
	header.AddSheet(theme.Name + "/panel.css")
	if len(theme.Resources) > 0 {
		rlist := theme.Resources
		for _, resource := range rlist {
			if resource.Location == "global" || resource.Location == "panel" {
				extarr := strings.Split(resource.Name, ".")
				ext := extarr[len(extarr)-1]
				if ext == "css" {
					header.AddSheet(resource.Name)
				} else if ext == "js" {
					if resource.Async {
						header.AddScriptAsync(resource.Name)
					} else {
						header.AddScript(resource.Name)
					}
				}
			}
		}
	}

	//h := w.Header()
	//h.Set("Content-Security-Policy", "default-src 'self'")

	// TODO: GDPR. Add a global control panel notice warning the admins of staff members who don't have 2FA enabled
	stats.Users = Users.GlobalCount()
	stats.Groups = Groups.GlobalCount()
	stats.Forums = Forums.GlobalCount()
	stats.Pages = Pages.GlobalCount()
	stats.Settings = len(header.Settings)
	stats.WordFilters = WordFilters.EstCount()
	stats.Themes = len(Themes)
	stats.Reports = 0 // TODO: Do the report count. Only show open threads?

	var addPreScript = func(name string) {
		var tname string
		if theme.OverridenMap != nil {
			_, ok := theme.OverridenMap[name]
			if ok {
				tname = "_" + theme.Name
			}
		}
		header.AddPreScriptAsync("template_" + name + tname + ".js")
	}
	addPreScript("alert")

	return header, stats, nil
}

func simplePanelUserCheck(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, rerr RouteError) {
	return SimpleUserCheck(w, r, user)
}

// SimpleUserCheck is back from the grave, yay :D
func simpleUserCheck(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, rerr RouteError) {
	return &HeaderLite{
		Site:     Site,
		Settings: SettingBox.Load().(SettingMap),
		Hooks:    GetHookTable(),
	}, nil
}

// TODO: Add the ability for admins to restrict certain themes to certain groups?
func userCheck(w http.ResponseWriter, r *http.Request, user *User) (header *Header, rerr RouteError) {
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

	header = &Header{
		Site:        Site,
		Settings:    SettingBox.Load().(SettingMap),
		Themes:      Themes,
		Theme:       theme,
		CurrentUser: *user, // ! Some things rely on this being a pointer downstream from this function
		Hooks:       GetHookTable(),
		Zone:        "frontend",
		Writer:      w,
		IsoCode:     phrases.GetLangPack().IsoCode,
	}
	header.GoogSiteVerify = header.Settings["google_site_verify"].(string)

	if user.IsBanned {
		header.AddNotice("account_banned")
	}
	if user.Loggedin && !user.Active {
		header.AddNotice("account_inactive")
	}
	// An optimisation so we don't populate StartedAt for users who shouldn't see the stat anyway
	// ? - Should we only show this in debug mode? It might be useful for detecting issues in production, if we show it there as-well
	if user.IsAdmin {
		header.StartedAt = time.Now()
	}

	header.AddSheet(theme.Name + "/main.css")
	if len(theme.Resources) > 0 {
		rlist := theme.Resources
		for _, resource := range rlist {
			if resource.Loggedin && !user.Loggedin {
				continue
			}
			if resource.Location == "global" || resource.Location == "frontend" {
				extarr := strings.Split(resource.Name, ".")
				ext := extarr[len(extarr)-1]
				if ext == "css" {
					header.AddSheet(resource.Name)
				} else if ext == "js" {
					if resource.Async {
						header.AddScriptAsync(resource.Name)
					} else {
						header.AddScript(resource.Name)
					}
				}
			}
		}
	}

	var addPreScript = func(name string) {
		var tname string
		if theme.OverridenMap != nil {
			//fmt.Printf("name %+v\n", name)
			//fmt.Printf("theme.OverridenMap %+v\n", theme.OverridenMap)
			_, ok := theme.OverridenMap[name]
			if ok {
				tname = "_" + theme.Name
			}
		}
		//fmt.Printf("tname %+v\n", tname)
		header.AddPreScriptAsync("template_" + name + tname + ".js")
	}
	addPreScript("topics_topic")
	addPreScript("paginator")
	addPreScript("alert")
	addPreScript("topic_c_edit_post")

	return header, nil
}

func preRoute(w http.ResponseWriter, r *http.Request) (User, bool) {
	userptr, halt := Auth.SessionCheck(w, r)
	if halt {
		return *userptr, false
	}
	var usercpy *User = BlankUser()
	*usercpy = *userptr
	usercpy.Init() // TODO: Can we reduce the amount of work we do here?

	// TODO: Add a config setting to disable this header
	// TODO: Have this header cover more things
	if Site.EnableSsl {
		w.Header().Set("Content-Security-Policy", "upgrade-insecure-requests")
	}

	// TODO: WIP. Refactor this to eliminate the unnecessary query
	// TODO: Better take proxies into consideration
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		PreError("Bad IP", w, r)
		return *usercpy, false
	}

	// TODO: Prefer Cf-Connecting-Ip header, fewer shenanigans
	if Site.HasProxy {
		// TODO: Check the right-most IP, might get tricky with multiple proxies, maybe have a setting for the number of hops we jump through
		xForwardedFor := r.Header.Get("X-Forwarded-For")
		if xForwardedFor != "" {
			forwardedFor := strings.Split(xForwardedFor, ",")
			// TODO: Check if this is a valid IP Address, reject if not
			host = forwardedFor[len(forwardedFor)-1]
		}
	}

	usercpy.LastIP = host

	if usercpy.Loggedin && host != usercpy.LastIP {
		err = usercpy.UpdateIP(host)
		if err != nil {
			InternalError(err, w, r)
			return *usercpy, false
		}
	}

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
	r.Body = http.MaxBytesReader(w, r.Body, r.ContentLength)

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
