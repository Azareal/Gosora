package main

import (
	//"log"
	//"fmt"
	"html"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var guestUser = User{ID: 0, Link: "#", Group: 6, Perms: GuestPerms}

// nolint
var PreRoute func(http.ResponseWriter, *http.Request) (User, bool) = preRoute

// TODO: Are these even session checks anymore? We might need to rethink these names
// nolint We need these types so people can tell what they are without scrolling to the bottom of the file
var PanelSessionCheck func(http.ResponseWriter, *http.Request, *User) (*HeaderVars, PanelStats, bool) = _panel_session_check
var SimplePanelSessionCheck func(http.ResponseWriter, *http.Request, *User) (*HeaderLite, bool) = _simple_panel_session_check
var SimpleForumSessionCheck func(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerLite *HeaderLite, success bool) = _simple_forum_session_check
var ForumSessionCheck func(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerVars *HeaderVars, success bool) = _forum_session_check
var SimpleSessionCheck func(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, success bool) = _simple_session_check
var SessionCheck func(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, success bool) = _session_check

//func(real_password string, password string, salt string) (err error)
var CheckPassword = BcryptCheckPassword

//func(password string) (hashed_password string, salt string, err error)
var GeneratePassword = BcryptGeneratePassword

type User struct {
	ID           int
	Link         string
	Name         string
	Email        string
	Group        int
	Active       bool
	IsMod        bool
	IsSuperMod   bool
	IsAdmin      bool
	IsSuperAdmin bool
	IsBanned     bool
	Perms        Perms
	PluginPerms  map[string]bool
	Session      string
	Loggedin     bool
	Avatar       string
	Message      string
	URLPrefix    string // Move this to another table? Create a user lite?
	URLName      string
	Tag          string
	Level        int
	Score        int
	LastIP       string
	TempGroup    int
}

type Email struct {
	UserID    int
	Email     string
	Validated bool
	Primary   bool
	Token     string
}

// duration in seconds
func (user *User) Ban(duration time.Duration, issuedBy int) error {
	return user.ScheduleGroupUpdate(4, issuedBy, duration)
}

func (user *User) Unban() error {
	err := user.RevertGroupUpdate()
	if err != nil {
		return err
	}
	return users.Load(user.ID)
}

// TODO: Use a transaction to avoid race conditions
// Make this more stateless?
func (user *User) ScheduleGroupUpdate(gid int, issuedBy int, duration time.Duration) error {
	var temporary bool
	if duration.Nanoseconds() != 0 {
		temporary = true
	}

	revertAt := time.Now().Add(duration)
	_, err := replace_schedule_group_stmt.Exec(user.ID, gid, issuedBy, revertAt, temporary)
	if err != nil {
		return err
	}
	_, err = set_temp_group_stmt.Exec(gid, user.ID)
	if err != nil {
		return err
	}
	return users.Load(user.ID)
}

// TODO: Use a transaction to avoid race conditions
func (user *User) RevertGroupUpdate() error {
	_, err := replace_schedule_group_stmt.Exec(user.ID, 0, 0, time.Now(), false)
	if err != nil {
		return err
	}
	_, err = set_temp_group_stmt.Exec(0, user.ID)
	if err != nil {
		return err
	}
	return users.Load(user.ID)
}

func BcryptCheckPassword(realPassword string, password string, salt string) (err error) {
	return bcrypt.CompareHashAndPassword([]byte(realPassword), []byte(password+salt))
}

// Investigate. Do we need the extra salt?
func BcryptGeneratePassword(password string) (hashedPassword string, salt string, err error) {
	salt, err = GenerateSafeString(saltLength)
	if err != nil {
		return "", "", err
	}

	password = password + salt
	hashedPassword, err = BcryptGeneratePasswordNoSalt(password)
	if err != nil {
		return "", "", err
	}
	return hashedPassword, salt, nil
}

func BcryptGeneratePasswordNoSalt(password string) (hash string, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func SetPassword(uid int, password string) error {
	hashedPassword, salt, err := GeneratePassword(password)
	if err != nil {
		return err
	}
	_, err = set_password_stmt.Exec(hashedPassword, salt, uid)
	return err
}

func SendValidationEmail(username string, email string, token string) bool {
	var schema = "http"
	if site.EnableSsl {
		schema += "s"
	}

	// TODO: Move these to the phrase system
	subject := "Validate Your Email @ " + site.Name
	msg := "Dear " + username + ", following your registration on our forums, we ask you to validate your email, so that we can confirm that this email actually belongs to you.\n\nClick on the following link to do so. " + schema + "://" + site.Url + "/user/edit/token/" + token + "\n\nIf you haven't created an account here, then please feel free to ignore this email.\nWe're sorry for the inconvenience this may have caused."
	return SendEmail(email, subject, msg)
}

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

func _simple_forum_session_check(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerLite *HeaderLite, success bool) {
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

	fperms := groups[user.Group].Forums[fid]
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

func _forum_session_check(w http.ResponseWriter, r *http.Request, user *User, fid int) (headerVars *HeaderVars, success bool) {
	headerVars, success = SessionCheck(w, r, user)
	if !fstore.Exists(fid) {
		NotFound(w, r)
		return headerVars, false
	}

	if vhooks["forum_check_pre_perms"] != nil {
		if runVhook("forum_check_pre_perms", w, r, user, &fid, &success, &headerVars).(bool) {
			return headerVars, success
		}
	}

	fperms := groups[user.Group].Forums[fid]
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
func _panel_session_check(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, stats PanelStats, success bool) {
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

	err = group_count_stmt.QueryRow().Scan(&stats.Groups)
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

func _simple_panel_session_check(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, success bool) {
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

// SimpleSessionCheck is back from the grave, yay :D
func _simple_session_check(w http.ResponseWriter, r *http.Request, user *User) (headerLite *HeaderLite, success bool) {
	headerLite = &HeaderLite{
		Site:     site,
		Settings: settingBox.Load().(SettingBox),
	}
	return headerLite, true
}

// TODO: Add the ability for admins to restrict certain themes to certain groups?
func _session_check(w http.ResponseWriter, r *http.Request, user *User) (headerVars *HeaderVars, success bool) {
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
		_, err = update_last_ip_stmt.Exec(host, user.ID)
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

func wordsToScore(wcount int, topic bool) (score int) {
	if topic {
		score = 2
	} else {
		score = 1
	}

	settings := settingBox.Load().(SettingBox)
	if wcount >= settings["megapost_min_words"].(int) {
		score += 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		score++
	}
	return score
}

// TODO: Move this to where the other User methods are
func (user *User) increasePostStats(wcount int, topic bool) error {
	var mod int
	baseScore := 1
	if topic {
		_, err := increment_user_topics_stmt.Exec(1, user.ID)
		if err != nil {
			return err
		}
		baseScore = 2
	}

	settings := settingBox.Load().(SettingBox)
	if wcount >= settings["megapost_min_words"].(int) {
		_, err := increment_user_megaposts_stmt.Exec(1, 1, 1, user.ID)
		if err != nil {
			return err
		}
		mod = 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		_, err := increment_user_bigposts_stmt.Exec(1, 1, user.ID)
		if err != nil {
			return err
		}
		mod = 1
	} else {
		_, err := increment_user_posts_stmt.Exec(1, user.ID)
		if err != nil {
			return err
		}
	}
	_, err := increment_user_score_stmt.Exec(baseScore+mod, user.ID)
	if err != nil {
		return err
	}
	//log.Print(user.Score + base_score + mod)
	//log.Print(getLevel(user.Score + base_score + mod))
	// TODO: Use a transaction to prevent level desyncs?
	_, err = update_user_level_stmt.Exec(getLevel(user.Score+baseScore+mod), user.ID)
	return err
}

// TODO: Move this to where the other User methods are
func (user *User) decreasePostStats(wcount int, topic bool) error {
	var mod int
	baseScore := -1
	if topic {
		_, err := increment_user_topics_stmt.Exec(-1, user.ID)
		if err != nil {
			return err
		}
		baseScore = -2
	}

	settings := settingBox.Load().(SettingBox)
	if wcount >= settings["megapost_min_words"].(int) {
		_, err := increment_user_megaposts_stmt.Exec(-1, -1, -1, user.ID)
		if err != nil {
			return err
		}
		mod = 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		_, err := increment_user_bigposts_stmt.Exec(-1, -1, user.ID)
		if err != nil {
			return err
		}
		mod = 1
	} else {
		_, err := increment_user_posts_stmt.Exec(-1, user.ID)
		if err != nil {
			return err
		}
	}
	_, err := increment_user_score_stmt.Exec(baseScore-mod, user.ID)
	if err != nil {
		return err
	}
	// TODO: Use a transaction to prevent level desyncs?
	_, err = update_user_level_stmt.Exec(getLevel(user.Score-baseScore-mod), user.ID)
	return err
}

func initUserPerms(user *User) {
	if user.IsSuperAdmin {
		user.Perms = AllPerms
		user.PluginPerms = AllPluginPerms
	} else {
		user.Perms = groups[user.Group].Perms
		user.PluginPerms = groups[user.Group].PluginPerms
	}

	if user.TempGroup != 0 {
		user.Group = user.TempGroup
	}

	user.IsAdmin = user.IsSuperAdmin || groups[user.Group].IsAdmin
	user.IsSuperMod = user.IsAdmin || groups[user.Group].IsMod
	user.IsMod = user.IsSuperMod
	user.IsBanned = groups[user.Group].IsBanned
	if user.IsBanned && user.IsSuperMod {
		user.IsBanned = false
	}
}

func buildProfileURL(slug string, uid int) string {
	if slug == "" {
		return "/user/" + strconv.Itoa(uid)
	}
	return "/user/" + slug + "." + strconv.Itoa(uid)
}
