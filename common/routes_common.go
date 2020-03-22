package common

import (
	"crypto/subtle"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/uutils"
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
var UserCheckNano func(w http.ResponseWriter, r *http.Request, user *User, nano int64) (header *Header, err RouteError) = userCheck2

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
	header.CurrentUser = user // TODO: Use a pointer instead for CurrentUser, so we don't have to do this
	return rerr
}

// TODO: Put this on the user instance? Do we really want forum specific logic in there? Maybe, a method which spits a new pointer with the same contents as user?
func cascadeForumPerms(fp *ForumPerms, u *User) {
	if fp.Overrides && !u.IsSuperAdmin {
		u.Perms.ViewTopic = fp.ViewTopic
		u.Perms.LikeItem = fp.LikeItem
		u.Perms.CreateTopic = fp.CreateTopic
		u.Perms.EditTopic = fp.EditTopic
		u.Perms.DeleteTopic = fp.DeleteTopic
		u.Perms.CreateReply = fp.CreateReply
		u.Perms.EditReply = fp.EditReply
		u.Perms.DeleteReply = fp.DeleteReply
		u.Perms.PinTopic = fp.PinTopic
		u.Perms.CloseTopic = fp.CloseTopic
		u.Perms.MoveTopic = fp.MoveTopic

		if len(fp.ExtData) != 0 {
			for name, perm := range fp.ExtData {
				u.PluginPerms[name] = perm
			}
		}
	}
}

// Even if they have the right permissions, the control panel is only open to supermods+. There are many areas without subpermissions which assume that the current user is a supermod+ and admins are extremely unlikely to give these permissions to someone who isn't at-least a supermod to begin with
// TODO: Do a panel specific theme?
func panelUserCheck(w http.ResponseWriter, r *http.Request, user *User) (h *Header, stats PanelStats, rerr RouteError) {
	theme := GetThemeByReq(r)
	h = &Header{
		Site:        Site,
		Settings:    SettingBox.Load().(SettingMap),
		Themes:      Themes,
		Theme:       theme,
		CurrentUser: user,
		Hooks:       GetHookTable(),
		Zone:        "panel",
		Writer:      w,
		IsoCode:     phrases.GetLangPack().IsoCode,
		//StartedAt:   time.Now(),
		StartedAt: uutils.Nanotime(),
	}
	// TODO: We should probably initialise header.ExtData
	// ? - Should we only show this in debug mode? It might be useful for detecting issues in production, if we show it there as-well
	//if user.IsAdmin {
	//h.StartedAt = time.Now()
	//}

	h.AddSheet(theme.Name + "/main.css")
	h.AddSheet(theme.Name + "/panel.css")
	if len(theme.Resources) > 0 {
		rlist := theme.Resources
		for _, res := range rlist {
			if res.Location == "global" || res.Location == "panel" {
				extarr := strings.Split(res.Name, ".")
				ext := extarr[len(extarr)-1]
				if ext == "css" {
					h.AddSheet(res.Name)
				} else if ext == "js" {
					if res.Async {
						h.AddScriptAsync(res.Name)
					} else {
						h.AddScript(res.Name)
					}
				}
			}
		}
	}

	//h := w.Header()
	//h.Set("Content-Security-Policy", "default-src 'self'")

	// TODO: GDPR. Add a global control panel notice warning the admins of staff members who don't have 2FA enabled
	stats.Users = Users.Count()
	stats.Groups = Groups.Count()
	stats.Forums = Forums.Count()
	stats.Pages = Pages.Count()
	stats.Settings = len(h.Settings)
	stats.WordFilters = WordFilters.EstCount()
	stats.Themes = len(Themes)
	stats.Reports = 0 // TODO: Do the report count. Only show open threads?

	addPreScript := func(name string) {
		// TODO: Optimise this by removing a superfluous string alloc
		var tname string
		if theme.OverridenMap != nil {
			_, ok := theme.OverridenMap[name]
			if ok {
				tname = "_" + theme.Name
			}
		}
		h.AddPreScriptAsync("template_" + name + tname + ".js")
	}
	addPreScript("alert")
	addPreScript("notice")

	return h, stats, nil
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

func GetThemeByReq(r *http.Request) *Theme {
	theme := &Theme{Name: ""}
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

	return theme
}

func userCheck(w http.ResponseWriter, r *http.Request, user *User) (header *Header, rerr RouteError) {
	return userCheck2(w, r, user, uutils.Nanotime())
}

// TODO: Add the ability for admins to restrict certain themes to certain groups?
// ! Be careful about firing errors off here as CustomError uses this
func userCheck2(w http.ResponseWriter, r *http.Request, user *User, nano int64) (h *Header, rerr RouteError) {
	theme := GetThemeByReq(r)
	h = &Header{
		Site:        Site,
		Settings:    SettingBox.Load().(SettingMap),
		Themes:      Themes,
		Theme:       theme,
		CurrentUser: user, // ! Some things rely on this being a pointer downstream from this function
		Hooks:       GetHookTable(),
		Zone:        "frontend",
		Writer:      w,
		IsoCode:     phrases.GetLangPack().IsoCode,
		StartedAt:   nano,
	}
	// TODO: Optimise this by avoiding accessing a map string index
	if !user.Loggedin {
		h.GoogSiteVerify = h.Settings["google_site_verify"].(string)
	}

	if user.IsBanned {
		h.AddNotice("account_banned")
	}
	if user.Loggedin && !user.Active {
		h.AddNotice("account_inactive")
	}

	// An optimisation so we don't populate StartedAt for users who shouldn't see the stat anyway
	// ? - Should we only show this in debug mode? It might be useful for detecting issues in production, if we show it there as-well
	//if user.IsAdmin {
	//h.StartedAt = time.Now()
	//}

	//PrepResources(user,h,theme)
	return h, nil
}

func PrepResources(u *User, h *Header, theme *Theme) {
	h.AddSheet(theme.Name + "/main.css")

	if len(theme.Resources) > 0 {
		rlist := theme.Resources
		for _, res := range rlist {
			if res.Loggedin && !u.Loggedin {
				continue
			}
			if res.Location == "global" || res.Location == "frontend" {
				extarr := strings.Split(res.Name, ".")
				ext := extarr[len(extarr)-1]
				if ext == "css" {
					h.AddSheet(res.Name)
				} else if ext == "js" {
					if res.Async {
						h.AddScriptAsync(res.Name)
					} else {
						h.AddScript(res.Name)
					}
				}
			}
		}
	}

	addPreScript := func(name string) {
		// TODO: Optimise this by removing a superfluous string alloc
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
		h.AddPreScriptAsync("template_" + name + tname + ".js")
	}
	addPreScript("topics_topic")
	addPreScript("paginator")
	addPreScript("alert")
	addPreScript("notice")
	if u.Loggedin {
		addPreScript("topic_c_edit_post")
		addPreScript("topic_c_attach_item")
		addPreScript("topic_c_poll_input")
	}
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
	if Config.SslSchema {
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

	if !Config.DisableLastIP && usercpy.Loggedin && host != usercpy.GetIP() {
		mon := time.Now().Month()
		err = usercpy.UpdateIP(strconv.Itoa(int(mon)) + "-" + host)
		if err != nil {
			InternalError(err, w, r)
			return *usercpy, false
		}
	}
	usercpy.LastIP = host

	return *usercpy, true
}

func UploadAvatar(w http.ResponseWriter, r *http.Request, user *User, tuid int) (ext string, ferr RouteError) {
	// We don't want multiple files
	// TODO: Are we doing this correctly?
	filenameMap := make(map[string]bool)
	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			if hdr.Filename == "" {
				continue
			}
			filenameMap[hdr.Filename] = true
		}
	}
	if len(filenameMap) > 1 {
		return "", LocalError("You may only upload one avatar", w, r, user)
	}

	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			if hdr.Filename == "" {
				continue
			}
			inFile, err := hdr.Open()
			if err != nil {
				return "", LocalError("Upload failed", w, r, user)
			}
			defer inFile.Close()

			if ext == "" {
				extarr := strings.Split(hdr.Filename, ".")
				if len(extarr) < 2 {
					return "", LocalError("Bad file", w, r, user)
				}
				ext = extarr[len(extarr)-1]

				// TODO: Can we do this without a regex?
				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					return "", LocalError("Bad file extension", w, r, user)
				}
				ext = reg.ReplaceAllString(ext, "")
				ext = strings.ToLower(ext)

				if !ImageFileExts.Contains(ext) {
					return "", LocalError("You can only use an image for your avatar", w, r, user)
				}
			}

			// TODO: Centralise this string, so we don't have to change it in two different places when it changes
			outFile, err := os.Create("./uploads/avatar_" + strconv.Itoa(tuid) + "." + ext)
			if err != nil {
				return "", LocalError("Upload failed [File Creation Failed]", w, r, user)
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return "", LocalError("Upload failed [Copy Failed]", w, r, user)
			}
		}
	}
	if ext == "" {
		return "", LocalError("No file", w, r, user)
	}
	return ext, nil
}

func ChangeAvatar(path string, w http.ResponseWriter, r *http.Request, user *User) RouteError {
	err := user.ChangeAvatar(path)
	if err != nil {
		return InternalError(err, w, r)
	}

	// Clean up the old avatar data, so we don't end up with too many dead files in /uploads/
	if len(user.RawAvatar) > 2 {
		if user.RawAvatar[0] == '.' && user.RawAvatar[1] == '.' {
			err := os.Remove("./uploads/avatar_" + strconv.Itoa(user.ID) + "_tmp" + user.RawAvatar[1:])
			if err != nil && !os.IsNotExist(err) {
				LogWarning(err)
				return LocalError("Something went wrong", w, r, user)
			}
			err = os.Remove("./uploads/avatar_" + strconv.Itoa(user.ID) + "_w48" + user.RawAvatar[1:])
			if err != nil && !os.IsNotExist(err) {
				LogWarning(err)
				return LocalError("Something went wrong", w, r, user)
			}
		}
	}

	return nil
}

// SuperAdminOnly makes sure that only super admin can access certain critical panel routes
func SuperAdminOnly(w http.ResponseWriter, r *http.Request, user *User) RouteError {
	if !user.IsSuperAdmin {
		return NoPermissions(w, r, user)
	}
	return nil
}

// AdminOnly makes sure that only admins can access certain panel routes
func AdminOnly(w http.ResponseWriter, r *http.Request, user *User) RouteError {
	if !user.IsAdmin {
		return NoPermissions(w, r, user)
	}
	return nil
}

// SuperModeOnly makes sure that only super mods or higher can access the panel routes
func SuperModOnly(w http.ResponseWriter, r *http.Request, user *User) RouteError {
	if !user.IsSuperMod {
		return NoPermissions(w, r, user)
	}
	return nil
}

// MemberOnly makes sure that only logged in users can access this route
func MemberOnly(w http.ResponseWriter, r *http.Request, user *User) RouteError {
	if !user.Loggedin {
		return LoginRequired(w, r, user)
	}
	return nil
}

// NoBanned stops any banned users from accessing this route
func NoBanned(w http.ResponseWriter, r *http.Request, user *User) RouteError {
	if user.IsBanned {
		return Banned(w, r, user)
	}
	return nil
}

func ParseForm(w http.ResponseWriter, r *http.Request, user *User) RouteError {
	if err := r.ParseForm(); err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	return nil
}

func NoSessionMismatch(w http.ResponseWriter, r *http.Request, user *User) RouteError {
	if err := r.ParseForm(); err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	// TODO: Try to eliminate some of these allocations
	sess := []byte(user.Session)
	if len(sess) == 0 {
		return SecurityError(w, r, user)
	}
	if subtle.ConstantTimeCompare([]byte(r.FormValue("session")), sess) != 1 && subtle.ConstantTimeCompare([]byte(r.FormValue("s")), sess) != 1 {
		return SecurityError(w, r, user)
	}
	return nil
}

func ReqIsJson(r *http.Request) bool {
	return r.Header.Get("Content-type") == "application/json"
}

func HandleUploadRoute(w http.ResponseWriter, r *http.Request, user *User, maxFileSize int) RouteError {
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

func NoUploadSessionMismatch(w http.ResponseWriter, r *http.Request, user *User) RouteError {
	// TODO: Try to eliminate some of these allocations
	sess := []byte(user.Session)
	if len(sess) == 0 {
		return SecurityError(w, r, user)
	}
	if subtle.ConstantTimeCompare([]byte(r.FormValue("session")), sess) != 1 && subtle.ConstantTimeCompare([]byte(r.FormValue("s")), sess) != 1 {
		return SecurityError(w, r, user)
	}
	return nil
}
