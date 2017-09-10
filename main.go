/* Copyright Azareal 2016 - 2018 */
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
	//"runtime/pprof"
)

var version = Version{Major: 0, Minor: 1, Patch: 0, Tag: "dev"}

const hour int = 60 * 60
const day int = hour * 24
const week int = day * 7
const month int = day * 30
const year int = day * 365
const kilobyte int = 1024
const megabyte int = kilobyte * 1024
const gigabyte int = megabyte * 1024
const terabyte int = gigabyte * 1024
const saltLength int = 32
const sessionLength int = 80

var enableWebsockets = false // Don't change this, the value is overwritten by an initialiser

var router *GenRouter
var startTime time.Time
var externalSites = map[string]string{
	"YT": "https://www.youtube.com/",
}
var groups []Group
var groupCapCount int
var staticFiles = make(map[string]SFile)
var logWriter = io.MultiWriter(os.Stderr)

type WordFilter struct {
	ID          int
	Find        string
	Replacement string
}
type WordFilterBox map[int]WordFilter

var wordFilterBox atomic.Value // An atomic value holding a WordFilterBox

func init() {
	wordFilterBox.Store(WordFilterBox(make(map[int]WordFilter)))
}

func LoadWordFilters() error {
	rows, err := get_word_filters_stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var wordFilters = WordFilterBox(make(map[int]WordFilter))
	var wfid int
	var find string
	var replacement string

	for rows.Next() {
		err := rows.Scan(&wfid, &find, &replacement)
		if err != nil {
			return err
		}
		wordFilters[wfid] = WordFilter{ID: wfid, Find: find, Replacement: replacement}
	}
	wordFilterBox.Store(wordFilters)
	return rows.Err()
}

func addWordFilter(id int, find string, replacement string) {
	wordFilters := wordFilterBox.Load().(WordFilterBox)
	wordFilters[id] = WordFilter{ID: id, Find: find, Replacement: replacement}
	wordFilterBox.Store(wordFilters)
}

func processConfig() {
	config.Noavatar = strings.Replace(config.Noavatar, "{site_url}", site.Url, -1)
	if site.Port != "80" && site.Port != "443" {
		site.Url = strings.TrimSuffix(site.Url, "/")
		site.Url = strings.TrimSuffix(site.Url, "\\")
		site.Url = strings.TrimSuffix(site.Url, ":")
		site.Url = site.Url + ":" + site.Port
	}
}

func main() {
	// TODO: Have a file for each run with the time/date the server started as the file name?
	// TODO: Log panics with recover()
	f, err := os.OpenFile("./operations.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	logWriter = io.MultiWriter(os.Stderr, f)
	log.SetOutput(logWriter)

	//if profiling {
	//	f, err := os.Create("startup_cpu.prof")
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	pprof.StartCPUProfile(f)
	//}

	log.Print("Running Gosora v" + version.String())
	fmt.Println("")
	startTime = time.Now()
	//timeLocation = startTime.Location()

	log.Print("Processing configuration data")
	processConfig()

	err = initThemes()
	if err != nil {
		log.Fatal(err)
	}

	err = initDatabase()
	if err != nil {
		log.Fatal(err)
	}

	initTemplates()
	err = initErrors()
	if err != nil {
		log.Fatal(err)
	}

	if config.CacheTopicUser == CACHE_STATIC {
		users = NewMemoryUserStore(config.UserCacheCapacity)
		topics = NewMemoryTopicStore(config.TopicCacheCapacity)
	} else {
		users = NewSQLUserStore()
		topics = NewSQLTopicStore()
	}

	log.Print("Loading the static files.")
	err = initStaticFiles()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Initialising the widgets")
	err = initWidgets()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Initialising the authentication system")
	auth = NewDefaultAuth()

	err = LoadWordFilters()
	if err != nil {
		log.Fatal(err)
	}

	// Run this goroutine once a second
	secondTicker := time.NewTicker(1 * time.Second)
	fifteenMinuteTicker := time.NewTicker(15 * time.Minute)
	//hour_ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-secondTicker.C:
				//log.Print("Running the second ticker")
				err := handleExpiredScheduledGroups()
				if err != nil {
					LogError(err)
				}

				// TODO: Handle delayed moderation tasks
				// TODO: Handle the daily clean-up. Move this to a 24 hour task?

				// Sync with the database, if there are any changes
				err = handleServerSync()
				if err != nil {
					LogError(err)
				}

				// TODO: Manage the TopicStore, UserStore, and ForumStore
				// TODO: Alert the admin, if CPU usage, RAM usage, or the number of posts in the past second are too high
				// TODO: Clean-up alerts with no unread matches which are over two weeks old. Move this to a 24 hour task?
			case <-fifteenMinuteTicker.C:
				// TODO: Automatically lock topics, if they're really old, and the associated setting is enabled.
				// TODO: Publish scheduled posts.
				// TODO: Delete the empty users_groups_scheduler entries
			}
		}
	}()

	log.Print("Initialising the router")
	router = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	///router.HandleFunc("/static/", route_static)
	///router.HandleFunc("/overview/", route_overview)
	///router.HandleFunc("/topics/create/", route_topic_create)
	///router.HandleFunc("/topics/", route_topics)
	///router.HandleFunc("/forums/", route_forums)
	///router.HandleFunc("/forum/", route_forum)
	router.HandleFunc("/topic/create/submit/", route_topic_create_submit)
	router.HandleFunc("/topic/", route_topic_id)
	router.HandleFunc("/reply/create/", route_create_reply)
	//router.HandleFunc("/reply/edit/", route_reply_edit)
	//router.HandleFunc("/reply/delete/", route_reply_delete)
	router.HandleFunc("/reply/edit/submit/", route_reply_edit_submit)
	router.HandleFunc("/reply/delete/submit/", route_reply_delete_submit)
	router.HandleFunc("/reply/like/submit/", route_reply_like_submit)
	///router.HandleFunc("/report/submit/", route_report_submit)
	router.HandleFunc("/topic/edit/submit/", route_edit_topic)
	router.HandleFunc("/topic/delete/submit/", route_delete_topic)
	router.HandleFunc("/topic/stick/submit/", route_stick_topic)
	router.HandleFunc("/topic/unstick/submit/", route_unstick_topic)
	router.HandleFunc("/topic/like/submit/", route_like_topic)

	// Custom Pages
	router.HandleFunc("/pages/", route_custom_page)

	// Accounts
	router.HandleFunc("/accounts/login/", route_login)
	router.HandleFunc("/accounts/create/", route_register)
	router.HandleFunc("/accounts/logout/", route_logout)
	router.HandleFunc("/accounts/login/submit/", route_login_submit)
	router.HandleFunc("/accounts/create/submit/", route_register_submit)

	//router.HandleFunc("/accounts/list/", route_login) // Redirect /accounts/ and /user/ to here.. // Get a list of all of the accounts on the forum
	//router.HandleFunc("/accounts/create/full/", route_logout) // Advanced account creator for admins?
	//router.HandleFunc("/user/edit/", route_logout)
	router.HandleFunc("/user/edit/critical/", route_account_own_edit_critical) // Password & Email
	router.HandleFunc("/user/edit/critical/submit/", route_account_own_edit_critical_submit)
	router.HandleFunc("/user/edit/avatar/", route_account_own_edit_avatar)
	router.HandleFunc("/user/edit/avatar/submit/", route_account_own_edit_avatar_submit)
	router.HandleFunc("/user/edit/username/", route_account_own_edit_username)
	router.HandleFunc("/user/edit/username/submit/", route_account_own_edit_username_submit)
	router.HandleFunc("/user/edit/email/", route_account_own_edit_email)
	router.HandleFunc("/user/edit/token/", route_account_own_edit_email_token_submit)
	router.HandleFunc("/user/", route_profile)
	router.HandleFunc("/profile/reply/create/", route_profile_reply_create)
	router.HandleFunc("/profile/reply/edit/submit/", route_profile_reply_edit_submit)
	router.HandleFunc("/profile/reply/delete/submit/", route_profile_reply_delete_submit)
	//router.HandleFunc("/user/edit/submit/", route_logout) // route_logout? what on earth? o.o
	//router.HandleFunc("/users/ban/", route_ban)
	router.HandleFunc("/users/ban/submit/", route_ban_submit)
	router.HandleFunc("/users/unban/", route_unban)
	router.HandleFunc("/users/activate/", route_activate)
	router.HandleFunc("/users/ips/", route_ips)

	// The Control Panel
	///router.HandleFunc("/panel/", route_panel)
	///router.HandleFunc("/panel/forums/", route_panel_forums)
	///router.HandleFunc("/panel/forums/create/", route_panel_forums_create_submit)
	///router.HandleFunc("/panel/forums/delete/", route_panel_forums_delete)
	///router.HandleFunc("/panel/forums/delete/submit/", route_panel_forums_delete_submit)
	///router.HandleFunc("/panel/forums/edit/", route_panel_forums_edit)
	///router.HandleFunc("/panel/forums/edit/submit/", route_panel_forums_edit_submit)
	///router.HandleFunc("/panel/forums/edit/perms/submit/", route_panel_forums_edit_perms_submit)
	///router.HandleFunc("/panel/settings/", route_panel_settings)
	///router.HandleFunc("/panel/settings/edit/", route_panel_setting)
	///router.HandleFunc("/panel/settings/edit/submit/", route_panel_setting_edit)
	///router.HandleFunc("/panel/themes/", route_panel_themes)
	///router.HandleFunc("/panel/themes/default/", route_panel_themes_default)
	///router.HandleFunc("/panel/plugins/", route_panel_plugins)
	///router.HandleFunc("/panel/plugins/activate/", route_panel_plugins_activate)
	///router.HandleFunc("/panel/plugins/deactivate/", route_panel_plugins_deactivate)
	///router.HandleFunc("/panel/users/", route_panel_users)
	///router.HandleFunc("/panel/users/edit/", route_panel_users_edit)
	///router.HandleFunc("/panel/users/edit/submit/", route_panel_users_edit_submit)
	///router.HandleFunc("/panel/groups/", route_panel_groups)
	///router.HandleFunc("/panel/groups/edit/", route_panel_groups_edit)
	///router.HandleFunc("/panel/groups/edit/perms/", route_panel_groups_edit_perms)
	///router.HandleFunc("/panel/groups/edit/submit/", route_panel_groups_edit_submit)
	///router.HandleFunc("/panel/groups/edit/perms/submit/", route_panel_groups_edit_perms_submit)
	///router.HandleFunc("/panel/groups/create/", route_panel_groups_create_submit)
	///router.HandleFunc("/panel/logs/mod/", route_panel_logs_mod)
	///router.HandleFunc("/panel/debug/", route_panel_debug)

	///router.HandleFunc("/api/", route_api)
	//router.HandleFunc("/exit/", route_exit)
	///router.HandleFunc("/", default_route)
	router.HandleFunc("/ws/", route_websockets)

	log.Print("Initialising the plugins")
	initPlugins()

	defer db.Close()

	//if profiling {
	//	pprof.StopCPUProfile()
	//}

	// TODO: Let users run *both* HTTP and HTTPS
	log.Print("Initialising the HTTP server")
	if !site.EnableSsl {
		if site.Port == "" {
			site.Port = "80"
		}
		log.Print("Listening on port " + site.Port)
		err = http.ListenAndServe(":"+site.Port, router)
	} else {
		if site.Port == "" {
			site.Port = "443"
		}
		if site.Port == "80" || site.Port == "443" {
			// We should also run the server on port 80
			// TODO: Redirect to port 443
			go func() {
				log.Print("Listening on port 80")
				err = http.ListenAndServe(":80", &HTTPSRedirect{})
				if err != nil {
					log.Fatal(err)
				}
			}()
		}
		log.Print("Listening on port " + site.Port)
		err = http.ListenAndServeTLS(":"+site.Port, config.SslFullchain, config.SslPrivkey, router)
	}

	// Why did the server stop?
	if err != nil {
		log.Fatal(err)
	}
}
