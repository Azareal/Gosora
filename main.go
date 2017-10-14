/*
*
*	Gosora Main File
*	Copyright Azareal 2016 - 2018
*
 */
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
const petabyte int = terabyte * 1024
const saltLength int = 32
const sessionLength int = 80

var router *GenRouter
var startTime time.Time

// ? - Make this more customisable?
var externalSites = map[string]string{
	"YT": "https://www.youtube.com/",
}

type StringList []string

// ? - Should we allow users to upload .php or .go files? It could cause security issues. We could store them with a mangled extension to render them inert
// TODO: Let admins manage this from the Control Panel
var allowedFileExts = StringList{
	"png", "jpg", "jpeg", "svg", "bmp", "gif", "tif", "webp", "apng", // images

	"txt", "xml", "json", "yaml", "toml", "ini", "md", "html", "rtf", "js", "py", "rb", "css", "scss", "less", "eqcss", "java", "ts", "cs", "c", "cc", "cpp", "cxx", "C", "c++", "h", "hh", "hpp", "hxx", "h++", "rs", "rlib", "htaccess", "gitignore", // text

	"mp3", "mp4", "avi", "wmv", "webm", // video

	"otf", "woff2", "woff", "ttf", "eot", // fonts
}
var imageFileExts = StringList{
	"png", "jpg", "jpeg", "svg", "bmp", "gif", "tif", "webp", "apng",
}

// TODO: Write a test for this
func (slice StringList) Contains(needle string) bool {
	for _, item := range slice {
		if item == needle {
			return true
		}
	}
	return false
}

var staticFiles = make(map[string]SFile)
var logWriter = io.MultiWriter(os.Stderr)

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

	log.Print("Processing configuration data")
	err = processConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = initThemes()
	if err != nil {
		log.Fatal(err)
	}

	err = initDatabase()
	if err != nil {
		log.Fatal(err)
	}

	rstore = NewSQLReplyStore()
	prstore = NewSQLProfileReplyStore()

	initTemplates()

	err = initPhrases()
	if err != nil {
		log.Fatal(err)
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

	err = verifyConfig()
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
				// TODO: Add a plugin hook here

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
				// TODO: Rescan the static files for changes

				// TODO: Add a plugin hook here
			case <-fifteenMinuteTicker.C:
				// TODO: Add a plugin hook here

				// TODO: Automatically lock topics, if they're really old, and the associated setting is enabled.
				// TODO: Publish scheduled posts.
				// TODO: Delete the empty users_groups_scheduler entries

				// TODO: Add a plugin hook here
			}
		}
	}()

	log.Print("Initialising the router")
	router = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	////router.HandleFunc("/static/", routeStatic)
	////router.HandleFunc("/overview/", routeOverview)
	////router.HandleFunc("/topics/create/", routeTopicCreate)
	////router.HandleFunc("/topics/", routeTopics)
	////router.HandleFunc("/forums/", routeForums)
	////router.HandleFunc("/forum/", routeForum)
	router.HandleFunc("/topic/create/submit/", routeTopicCreateSubmit)
	router.HandleFunc("/topic/", routeTopicID)
	router.HandleFunc("/reply/create/", routeCreateReply)
	//router.HandleFunc("/reply/edit/", routeReplyEdit)
	//router.HandleFunc("/reply/delete/", routeReplyDelete)
	router.HandleFunc("/reply/edit/submit/", routeReplyEditSubmit)
	router.HandleFunc("/reply/delete/submit/", routeReplyDeleteSubmit)
	router.HandleFunc("/reply/like/submit/", routeReplyLikeSubmit)
	///router.HandleFunc("/report/submit/", route_report_submit)
	router.HandleFunc("/topic/edit/submit/", routeEditTopic)
	router.HandleFunc("/topic/delete/submit/", routeDeleteTopic)
	router.HandleFunc("/topic/stick/submit/", routeStickTopic)
	router.HandleFunc("/topic/unstick/submit/", routeUnstickTopic)
	router.HandleFunc("/topic/lock/submit/", routeLockTopic)
	router.HandleFunc("/topic/unlock/submit/", routeUnlockTopic)
	router.HandleFunc("/topic/like/submit/", routeLikeTopic)

	// Custom Pages
	router.HandleFunc("/pages/", routeCustomPage)

	// Accounts
	router.HandleFunc("/accounts/login/", routeLogin)
	router.HandleFunc("/accounts/create/", routeRegister)
	router.HandleFunc("/accounts/logout/", routeLogout)
	router.HandleFunc("/accounts/login/submit/", routeLoginSubmit)
	router.HandleFunc("/accounts/create/submit/", routeRegisterSubmit)

	//router.HandleFunc("/accounts/list/", routeLogin) // Redirect /accounts/ and /user/ to here.. // Get a list of all of the accounts on the forum
	//router.HandleFunc("/accounts/create/full/", routeLogout) // Advanced account creator for admins?
	//router.HandleFunc("/user/edit/", routeLogout)
	router.HandleFunc("/user/edit/critical/", routeAccountOwnEditCritical) // Password & Email
	router.HandleFunc("/user/edit/critical/submit/", routeAccountOwnEditCriticalSubmit)
	router.HandleFunc("/user/edit/avatar/", routeAccountOwnEditAvatar)
	router.HandleFunc("/user/edit/avatar/submit/", routeAccountOwnEditAvatarSubmit)
	router.HandleFunc("/user/edit/username/", routeAccountOwnEditUsername)
	router.HandleFunc("/user/edit/username/submit/", routeAccountOwnEditUsernameSubmit)
	router.HandleFunc("/user/edit/email/", routeAccountOwnEditEmail)
	router.HandleFunc("/user/edit/token/", routeAccountOwnEditEmailTokenSubmit)
	router.HandleFunc("/user/", routeProfile)
	router.HandleFunc("/profile/reply/create/", routeProfileReplyCreate)
	router.HandleFunc("/profile/reply/edit/submit/", routeProfileReplyEditSubmit)
	router.HandleFunc("/profile/reply/delete/submit/", routeProfileReplyDeleteSubmit)
	//router.HandleFunc("/user/edit/submit/", routeLogout) // routeLogout? what on earth? o.o
	//router.HandleFunc("/users/ban/", routeBan)
	router.HandleFunc("/users/ban/submit/", routeBanSubmit)
	router.HandleFunc("/users/unban/", routeUnban)
	router.HandleFunc("/users/activate/", routeActivate)
	router.HandleFunc("/users/ips/", routeIps)

	// The Control Panel
	// TODO: Rename the commented route handlers to the new camelCase format :'(
	////router.HandleFunc("/panel/", routePanel)
	////router.HandleFunc("/panel/forums/", routePanelForums)
	////router.HandleFunc("/panel/forums/create/", routePanelForumsCreateSubmit)
	////router.HandleFunc("/panel/forums/delete/", routePanelForumsDelete)
	////router.HandleFunc("/panel/forums/delete/submit/", routePanelForumsDeleteSubmit)
	////router.HandleFunc("/panel/forums/edit/", routePanelForumsEdit)
	////router.HandleFunc("/panel/forums/edit/submit/", routePanelForumsEditSubmit)
	////router.HandleFunc("/panel/forums/edit/perms/submit/", routePanelForumsEditPermsSubmit)
	////router.HandleFunc("/panel/settings/", routePanelSettings)
	////router.HandleFunc("/panel/settings/edit/", routePanelSetting)
	////router.HandleFunc("/panel/settings/edit/submit/", routePanelSettingEdit)
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

	////router.HandleFunc("/api/", routeAPI)
	//router.HandleFunc("/exit/", routeExit)
	////router.HandleFunc("/", config.DefaultRoute)
	router.HandleFunc("/ws/", routeWebsockets)

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
