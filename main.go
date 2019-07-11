/*
*
*	Gosora Main File
*	Copyright Azareal 2016 - 2020
*
 */
// Package main contains the main initialisation logic for Gosora
package main // import "github.com/Azareal/Gosora"

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
	"github.com/Azareal/Gosora/routes"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

var router *GenRouter

// TODO: Wrap the globals in here so we can pass pointers to them to subpackages
var globs *Globs

type Globs struct {
	stmts *Stmts
}

// Temporary alias for renderTemplate
func init() {
	c.RenderTemplateAlias = routes.RenderTemplate
}

func afterDBInit() (err error) {
	err = storeInit()
	if err != nil {
		return err
	}
	log.Print("Exitted storeInit")

	var uids []int
	tcache := c.Topics.GetCache()
	if tcache != nil {
		// Preload ten topics to get the wheels going
		var count = 10
		if tcache.GetCapacity() <= 10 {
			count = 2
			if tcache.GetCapacity() <= 2 {
				count = 0
			}
		}
		group, err := c.Groups.Get(c.GuestUser.Group)
		if err != nil {
			return err
		}

		// TODO: Use the same cached data for both the topic list and the topic fetches...
		tList, _, _, err := c.TopicList.GetListByCanSee(group.CanSee, 1, "", nil)
		if err != nil {
			return err
		}
		ctList := make([]*c.TopicsRow, len(tList))
		copy(ctList, tList)

		tList, _, _, err = c.TopicList.GetListByCanSee(group.CanSee, 2, "", nil)
		if err != nil {
			return err
		}
		for _, tItem := range tList {
			ctList = append(ctList, tItem)
		}

		tList, _, _, err = c.TopicList.GetListByCanSee(group.CanSee, 3, "", nil)
		if err != nil {
			return err
		}
		for _, tItem := range tList {
			ctList = append(ctList, tItem)
		}

		if count > len(ctList) {
			count = len(ctList)
		}
		for i := 0; i < count; i++ {
			_, _ = c.Topics.Get(ctList[i].ID)
		}
	}

	ucache := c.Users.GetCache()
	if ucache != nil {
		// Preload associated users too...
		for _, uid := range uids {
			_, _ = c.Users.Get(uid)
		}
	}

	log.Print("Exitted afterDBInit")
	return nil
}

// Experimenting with a new error package here to try to reduce the amount of debugging we have to do
// TODO: Dynamically register these items to avoid maintaining as much code here?
func storeInit() (err error) {
	acc := qgen.NewAcc()
	var rcache c.ReplyCache
	if c.Config.ReplyCache == "static" {
		rcache = c.NewMemoryReplyCache(c.Config.ReplyCacheCapacity)
	}
	c.Rstore, err = c.NewSQLReplyStore(acc, rcache)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Prstore, err = c.NewSQLProfileReplyStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Likes, err = c.NewDefaultLikeStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	/*c.Convos, err = c.NewDefaultConversationStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}*/

	err = phrases.InitPhrases(c.Site.Language)
	if err != nil {
		return errors.WithStack(err)
	}
	err = c.InitEmoji()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Loading the static files.")
	err = c.Themes.LoadStaticFiles()
	if err != nil {
		return errors.WithStack(err)
	}
	err = c.StaticFiles.Init()
	if err != nil {
		return errors.WithStack(err)
	}
	err = c.StaticFiles.JSTmplInit()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the widgets")
	c.Widgets = c.NewDefaultWidgetStore()
	err = c.InitWidgets()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the menu item list")
	c.Menus = c.NewDefaultMenuStore()
	err = c.Menus.Load(1) // 1 = the default menu
	if err != nil {
		return errors.WithStack(err)
	}
	menuHold, err := c.Menus.Get(1)
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Printf("menuHold: %+v\n", menuHold)
	var b bytes.Buffer
	menuHold.Build(&b, &c.GuestUser, "/")
	fmt.Println("menuHold output: ", string(b.Bytes()))

	log.Print("Initialising the authentication system")
	c.Auth, err = c.NewDefaultAuth()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the stores")
	c.WordFilters, err = c.NewDefaultWordFilterStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.MFAstore, err = c.NewSQLMFAStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Pages, err = c.NewDefaultPageStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Reports, err = c.NewDefaultReportStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Emails, err = c.NewDefaultEmailStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.LoginLogs, err = c.NewLoginLogStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.RegLogs, err = c.NewRegLogStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.ModLogs, err = c.NewModLogStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.AdminLogs, err = c.NewAdminLogStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.IPSearch, err = c.NewDefaultIPSearcher()
	if err != nil {
		return errors.WithStack(err)
	}
	if c.Config.Search == "" || c.Config.Search == "sql" {
		c.RepliesSearch, err = c.NewSQLSearcher(acc)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	c.Subscriptions, err = c.NewDefaultSubscriptionStore()
	if err != nil {
		return errors.WithStack(err)
	}
	c.Attachments, err = c.NewDefaultAttachmentStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Polls, err = c.NewDefaultPollStore(c.NewMemoryPollCache(100)) // TODO: Max number of polls held in cache, make this a config item
	if err != nil {
		return errors.WithStack(err)
	}
	c.TopicList, err = c.NewDefaultTopicList()
	if err != nil {
		return errors.WithStack(err)
	}
	c.PasswordResetter, err = c.NewDefaultPasswordResetter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Activity, err = c.NewDefaultActivityStream(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	// TODO: Let the admin choose other thumbnailers, maybe ones defined in plugins
	c.Thumbnailer = c.NewCaireThumbnailer()

	log.Print("Initialising the view counters")
	counters.GlobalViewCounter, err = counters.NewGlobalViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	counters.AgentViewCounter, err = counters.NewDefaultAgentViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	counters.OSViewCounter, err = counters.NewDefaultOSViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	counters.LangViewCounter, err = counters.NewDefaultLangViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	counters.RouteViewCounter, err = counters.NewDefaultRouteViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	counters.PostCounter, err = counters.NewPostCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	counters.TopicCounter, err = counters.NewTopicCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	counters.TopicViewCounter, err = counters.NewDefaultTopicViewCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	counters.ForumViewCounter, err = counters.NewDefaultForumViewCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	counters.ReferrerTracker, err = counters.NewDefaultReferrerTracker()
	if err != nil {
		return errors.WithStack(err)
	}
	counters.MemoryCounter, err = counters.NewMemoryCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the meta store")
	c.Meta, err = c.NewDefaultMetaStore(acc)
	if err != nil {
		return errors.WithStack(err)
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
	c.StartTime = time.Now()

	// TODO: Have a file for each run with the time/date the server started as the file name?
	// TODO: Log panics with recover()
	f, err := os.OpenFile("./logs/ops-"+strconv.FormatInt(c.StartTime.Unix(), 10)+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	c.LogWriter = io.MultiWriter(os.Stderr, f)
	log.SetOutput(c.LogWriter)
	log.Print("Running Gosora v" + c.SoftwareVersion.String())
	fmt.Println("")

	// TODO: Add a flag for enabling the profiler
	if false {
		f, err := os.Create("./logs/cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	jsToken, err := c.GenerateSafeString(80)
	if err != nil {
		log.Fatal(err)
	}
	c.JSTokenBox.Store(jsToken)

	log.Print("Loading the configuration data")
	err = c.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Processing configuration data")
	err = c.ProcessConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = c.InitTemplates()
	if err != nil {
		log.Fatal(err)
	}
	c.Themes, err = c.NewThemeList()
	if err != nil {
		log.Fatal(err)
	}
	c.TopicListThaw = c.NewSingleServerThaw()

	err = InitDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	buildTemplates := flag.Bool("build-templates", false, "build the templates")
	flag.Parse()
	if *buildTemplates {
		err = c.CompileTemplates()
		if err != nil {
			log.Fatal(err)
		}
		err = c.CompileJSTemplates()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err = afterDBInit()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	err = c.VerifyConfig()
	if err != nil {
		log.Fatal(err)
	}

	if !c.Dev.NoFsnotify {
		log.Print("Initialising the file watcher")
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
						theme, ok := c.Themes[themeName]
						if ok {
							return theme.LoadStaticFiles()
						}
					}
				}
				return nil
			}

			// TODO: Expand this to more types of files
			var err error
			for {
				select {
				case event := <-watcher.Events:
					// TODO: Handle file deletes (and renames more graciously by removing the old version of it)
					if event.Op&fsnotify.Write == fsnotify.Write {
						log.Println("modified file:", event.Name)
						err = modifiedFileEvent(event.Name)
					} else if event.Op&fsnotify.Create == fsnotify.Create {
						log.Println("new file:", event.Name)
						err = modifiedFileEvent(event.Name)
					} else {
						log.Println("unknown event:", event)
						err = nil
					}
					if err != nil {
						c.LogError(err)
					}
				case err = <-watcher.Errors:
					c.LogWarning(err)
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
		for _, theme := range c.Themes {
			err = watcher.Add("./themes/" + theme.Name + "/public")
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Print("Initialising the task system")

	// Thumbnailer goroutine, we only want one image being thumbnailed at a time, otherwise they might wind up consuming all the CPU time and leave no resources left to service the actual requests
	// TODO: Could we expand this to attachments and other things too?
	thumbChan := make(chan bool)
	go c.ThumbTask(thumbChan)
	go tickLoop(thumbChan)

	// Resource Management Goroutine
	go func() {
		ucache := c.Users.GetCache()
		tcache := c.Topics.GetCache()
		if ucache == nil && tcache == nil {
			return
		}

		var lastEvictedCount int
		var couldNotDealloc bool
		var secondTicker = time.NewTicker(time.Second)
		for {
			select {
			case <-secondTicker.C:
				// TODO: Add a LastRequested field to cached User structs to avoid evicting the same things which wind up getting loaded again anyway?
				if ucache != nil {
					ucap := ucache.GetCapacity()
					if ucache.Length() <= ucap || c.Users.Count() <= ucap {
						couldNotDealloc = false
						continue
					}
					lastEvictedCount = ucache.DeallocOverflow(couldNotDealloc)
					couldNotDealloc = (lastEvictedCount == 0)
				}
			}
		}
	}()

	log.Print("Initialising the router")
	router, err = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Initialising the plugins")
	c.InitPlugins()

	log.Print("Setting up the signal handler")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		// TODO: Gracefully shutdown the HTTP server
		runTasks(c.ShutdownTasks)
		c.StoppedServer("Received a signal to shutdown: ", sig)
	}()

	// Start up the WebSocket ticks
	c.WsHub.Start()

	if false {
		f, err := os.Create("./logs/cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	//if profiling {
	//	pprof.StopCPUProfile()
	//}
	startServer()
	args := <-c.StopServerChan
	if false {
		pprof.StopCPUProfile()
		f, err := os.Create("./logs/mem.prof")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		runtime.GC()
		err = pprof.WriteHeapProfile(f)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Why did the server stop?
	log.Fatal(args...)
}

func startServer() {
	// We might not need the timeouts, if we're behind a reverse-proxy like Nginx
	var newServer = func(addr string, handler http.Handler) *http.Server {
		rtime := c.Config.ReadTimeout
		if rtime == 0 {
			rtime = 8
		} else if rtime == -1 {
			rtime = 0
		}
		wtime := c.Config.WriteTimeout
		if wtime == 0 {
			wtime = 10
		} else if wtime == -1 {
			wtime = 0
		}
		itime := c.Config.IdleTimeout
		if itime == 0 {
			itime = 120
		} else if itime == -1 {
			itime = 0
		}
		return &http.Server{
			Addr:    addr,
			Handler: handler,

			ReadTimeout:  time.Duration(rtime) * time.Second,
			WriteTimeout: time.Duration(wtime) * time.Second,
			IdleTimeout:  time.Duration(itime) * time.Second,

			TLSConfig: &tls.Config{
				PreferServerCipherSuites: true,
				CurvePreferences: []tls.CurveID{
					tls.CurveP256,
					tls.X25519,
				},
			},
		}
	}

	// TODO: Let users run *both* HTTP and HTTPS
	log.Print("Initialising the HTTP server")
	if !c.Site.EnableSsl {
		if c.Site.Port == "" {
			c.Site.Port = "80"
		}
		log.Print("Listening on port " + c.Site.Port)
		go func() {
			c.StoppedServer(newServer(":"+c.Site.Port, router).ListenAndServe())
		}()
		return
	}

	if c.Site.Port == "" {
		c.Site.Port = "443"
	}
	if c.Site.Port == "80" || c.Site.Port == "443" {
		// We should also run the server on port 80
		// TODO: Redirect to port 443
		go func() {
			log.Print("Listening on port 80")
			c.StoppedServer(newServer(":80", &HTTPSRedirect{}).ListenAndServe())
		}()
	}
	log.Printf("Listening on port %s", c.Site.Port)
	go func() {
		c.StoppedServer(newServer(":"+c.Site.Port, router).ListenAndServeTLS(c.Config.SslFullchain, c.Config.SslPrivkey))
	}()
}
