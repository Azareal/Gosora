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
	"os/signal"
	"strings"
	"syscall"
	"time"
	//"runtime/pprof"
	"./common"
	"github.com/fsnotify/fsnotify"
)

var version = common.Version{Major: 0, Minor: 1, Patch: 0, Tag: "dev"}
var router *GenRouter
var startTime time.Time
var logWriter = io.MultiWriter(os.Stderr)

// TODO: Wrap the globals in here so we can pass pointers to them to subpackages
var globs *Globs

type Globs struct {
	stmts *Stmts
}

func afterDBInit() (err error) {
	common.Rstore, err = common.NewSQLReplyStore()
	if err != nil {
		return err
	}
	common.Prstore, err = common.NewSQLProfileReplyStore()
	if err != nil {
		return err
	}

	err = common.InitTemplates()
	if err != nil {
		return err
	}
	err = common.InitPhrases()
	if err != nil {
		return err
	}

	log.Print("Loading the static files.")
	err = common.StaticFiles.Init()
	if err != nil {
		return err
	}
	log.Print("Initialising the widgets")
	err = common.InitWidgets()
	if err != nil {
		return err
	}
	log.Print("Initialising the authentication system")
	common.Auth, err = common.NewDefaultAuth()
	if err != nil {
		return err
	}

	err = common.LoadWordFilters()
	if err != nil {
		return err
	}
	common.ModLogs, err = common.NewModLogStore()
	if err != nil {
		return err
	}
	common.AdminLogs, err = common.NewAdminLogStore()
	if err != nil {
		return err
	}
	common.GlobalViewCounter, err = common.NewGlobalViewCounter()
	if err != nil {
		return err
	}

	return nil
}

// TODO: Split this function up
func main() {
	// TODO: Recover from panics
	/*defer func() {
		r := recover()
		if r != nil {
			log.Print(r)
			debug.PrintStack()
			return
		}
	}()*/

	// WIP: Mango Test
	/*res, err := ioutil.ReadFile("./templates/topic.html")
	if err != nil {
		log.Fatal(err)
	}
	tagIndices, err := mangoParse(string(res))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("tagIndices: %+v\n", tagIndices)
	log.Fatal("")*/

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
	err = common.ProcessConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = common.InitThemes()
	if err != nil {
		log.Fatal(err)
	}

	err = InitDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = afterDBInit()
	if err != nil {
		log.Fatal(err)
	}

	err = common.VerifyConfig()
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		var modifiedFileEvent = func(path string) error {
			var pathBits = strings.Split(path, "\\")
			if len(pathBits) == 0 {
				return nil
			}
			if pathBits[0] == "themes" {
				var themeName string
				if len(pathBits) >= 2 {
					themeName = pathBits[1]
				}
				if len(pathBits) >= 3 && pathBits[2] == "public" {
					// TODO: Handle new themes freshly plopped into the folder?
					theme, ok := common.Themes[themeName]
					if ok {
						return theme.LoadStaticFiles()
					}
				}
			}
			return nil
		}

		var err error
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				// TODO: Handle file deletes (and renames more graciously by removing the old version of it)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					err = modifiedFileEvent(event.Name)
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("new file:", event.Name)
					err = modifiedFileEvent(event.Name)
				}
				if err != nil {
					common.LogError(err)
				}
			case err = <-watcher.Errors:
				common.LogError(err)
			}
		}
	}()

	// TODO: Keep tabs on the (non-resource) theme stuff, and the langpacks
	err = watcher.Add("./public")
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add("./templates")
	if err != nil {
		log.Fatal(err)
	}
	for _, theme := range common.Themes {
		err = watcher.Add("./themes/" + theme.Name + "/public")
		if err != nil {
			log.Fatal(err)
		}
	}

	// Run this goroutine once a second
	secondTicker := time.NewTicker(1 * time.Second)
	fifteenMinuteTicker := time.NewTicker(15 * time.Minute)
	//hourTicker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-secondTicker.C:
				//log.Print("Running the second ticker")
				// TODO: Add a plugin hook here

				for _, task := range common.ScheduledSecondTasks {
					if task() != nil {
						common.LogError(err)
					}
				}

				err := common.HandleExpiredScheduledGroups()
				if err != nil {
					common.LogError(err)
				}

				// TODO: Handle delayed moderation tasks
				// TODO: Handle the daily clean-up. Move this to a 24 hour task?

				// Sync with the database, if there are any changes
				err = common.HandleServerSync()
				if err != nil {
					common.LogError(err)
				}

				// TODO: Manage the TopicStore, UserStore, and ForumStore
				// TODO: Alert the admin, if CPU usage, RAM usage, or the number of posts in the past second are too high
				// TODO: Clean-up alerts with no unread matches which are over two weeks old. Move this to a 24 hour task?
				// TODO: Rescan the static files for changes

				// TODO: Add a plugin hook here
			case <-fifteenMinuteTicker.C:
				// TODO: Add a plugin hook here

				for _, task := range common.ScheduledFifteenMinuteTasks {
					if task() != nil {
						common.LogError(err)
					}
				}

				// TODO: Automatically lock topics, if they're really old, and the associated setting is enabled.
				// TODO: Publish scheduled posts.

				// TODO: Add a plugin hook here
			}
		}
	}()

	log.Print("Initialising the router")
	router = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	router.HandleFunc("/topic/create/submit/", routeTopicCreateSubmit)
	router.HandleFunc("/topic/", routeTopicID)
	router.HandleFunc("/reply/create/", routeCreateReply)
	//router.HandleFunc("/reply/edit/", routeReplyEdit)
	//router.HandleFunc("/reply/delete/", routeReplyDelete)
	router.HandleFunc("/reply/edit/submit/", routeReplyEditSubmit)
	router.HandleFunc("/reply/delete/submit/", routeReplyDeleteSubmit)
	router.HandleFunc("/reply/like/submit/", routeReplyLikeSubmit)
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

	// TODO: Move these into /user/?
	router.HandleFunc("/profile/reply/create/", routeProfileReplyCreate)
	router.HandleFunc("/profile/reply/edit/submit/", routeProfileReplyEditSubmit)
	router.HandleFunc("/profile/reply/delete/submit/", routeProfileReplyDeleteSubmit)
	//router.HandleFunc("/user/edit/submit/", routeLogout) // routeLogout? what on earth? o.o
	router.HandleFunc("/ws/", routeWebsockets)

	log.Print("Initialising the plugins")
	common.InitPlugins()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		// TODO: Gracefully shutdown the HTTP server
		log.Fatal("Received a signal to shutdown: ", sig)
	}()

	//if profiling {
	//	pprof.StopCPUProfile()
	//}

	// TODO: Let users run *both* HTTP and HTTPS
	log.Print("Initialising the HTTP server")
	if !common.Site.EnableSsl {
		if common.Site.Port == "" {
			common.Site.Port = "80"
		}
		log.Print("Listening on port " + common.Site.Port)
		err = http.ListenAndServe(":"+common.Site.Port, router)
	} else {
		if common.Site.Port == "" {
			common.Site.Port = "443"
		}
		if common.Site.Port == "80" || common.Site.Port == "443" {
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
		log.Printf("Listening on port %s", common.Site.Port)
		err = http.ListenAndServeTLS(":"+common.Site.Port, common.Config.SslFullchain, common.Config.SslPrivkey, router)
	}

	// Why did the server stop?
	if err != nil {
		log.Fatal(err)
	}
}
