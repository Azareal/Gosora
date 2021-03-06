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
	"mime"
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
	co "github.com/Azareal/Gosora/common/counters"
	meta "github.com/Azareal/Gosora/common/meta"
	p "github.com/Azareal/Gosora/common/phrases"
	_ "github.com/Azareal/Gosora/extend"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/Azareal/Gosora/routes"
	"github.com/fsnotify/fsnotify"

	//"github.com/lucas-clemente/quic-go/http3"
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
	if err := storeInit(); err != nil {
		return err
	}
	log.Print("Exitted storeInit")

	c.GzipStartEtag = "\"" + strconv.FormatInt(c.StartTime.Unix(), 10) + "-ng\""
	c.StartEtag = "\"" + strconv.FormatInt(c.StartTime.Unix(), 10) + "-n\""

	var uids []int
	tc := c.Topics.GetCache()
	if tc != nil {
		// Preload ten topics to get the wheels going
		var count = 10
		if tc.GetCapacity() <= 10 {
			count = 2
			if tc.GetCapacity() <= 2 {
				count = 0
			}
		}
		group, err := c.Groups.Get(c.GuestUser.Group)
		if err != nil {
			return err
		}

		// TODO: Use the same cached data for both the topic list and the topic fetches...
		tList, _, _, err := c.TopicList.GetListByCanSee(group.CanSee, 1, 0, nil)
		if err != nil {
			return err
		}
		ctList := make([]*c.TopicsRow, len(tList))
		copy(ctList, tList)

		tList, _, _, err = c.TopicList.GetListByCanSee(group.CanSee, 2, 0, nil)
		if err != nil {
			return err
		}
		for _, tItem := range tList {
			ctList = append(ctList, tItem)
		}

		tList, _, _, err = c.TopicList.GetListByCanSee(group.CanSee, 3, 0, nil)
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

	uc := c.Users.GetCache()
	if uc != nil {
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
	c.Convos, err = c.NewDefaultConversationStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.UserBlocks, err = c.NewDefaultBlockStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.GroupPromotions, err = c.NewDefaultGroupPromotionStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}

	if err = p.InitPhrases(c.Site.Language); err != nil {
		return errors.WithStack(err)
	}
	if err = c.InitEmoji(); err != nil {
		return errors.WithStack(err)
	}
	if err = c.InitWeakPasswords(); err != nil {
		return errors.WithStack(err)
	}

	log.Print("Loading the static files.")
	if err = c.Themes.LoadStaticFiles(); err != nil {
		return errors.WithStack(err)
	}
	if err = c.StaticFiles.Init(); err != nil {
		return errors.WithStack(err)
	}
	if err = c.StaticFiles.JSTmplInit(); err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the widgets")
	c.Widgets = c.NewDefaultWidgetStore()
	if err = c.InitWidgets(); err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the menu item list")
	c.Menus = c.NewDefaultMenuStore()
	if err = c.Menus.Load(1); err != nil { // 1 = the default menu
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
	c.TopicList, err = c.NewDefaultTopicList(acc)
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
	c.Recalc, err = c.NewDefaultRecalc(acc)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the meta store")
	c.Meta, err = meta.NewDefaultMetaStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the view counters")
	if !c.Config.DisableAnalytics {
		co.GlobalViewCounter, err = co.NewGlobalViewCounter(acc)
		if err != nil {
			return errors.WithStack(err)
		}
		co.AgentViewCounter, err = co.NewDefaultAgentViewCounter(acc)
		if err != nil {
			return errors.WithStack(err)
		}
		co.OSViewCounter, err = co.NewDefaultOSViewCounter(acc)
		if err != nil {
			return errors.WithStack(err)
		}
		co.LangViewCounter, err = co.NewDefaultLangViewCounter(acc)
		if err != nil {
			return errors.WithStack(err)
		}
		if !c.Config.RefNoTrack {
			co.ReferrerTracker, err = co.NewDefaultReferrerTracker()
			if err != nil {
				return errors.WithStack(err)
			}
		}
		co.MemoryCounter, err = co.NewMemoryCounter(acc)
		if err != nil {
			return errors.WithStack(err)
		}
		co.PerfCounter, err = co.NewDefaultPerfCounter(acc)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	co.RouteViewCounter, err = co.NewDefaultRouteViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	co.PostCounter, err = co.NewPostCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	co.TopicCounter, err = co.NewTopicCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	co.TopicViewCounter, err = co.NewDefaultTopicViewCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	co.ForumViewCounter, err = co.NewDefaultForumViewCounter()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// TODO: Split this function up
func main() {
	// TODO: Recover from panics
	/*defer func() {
		if r := recover(); r != nil {
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

	err = mime.AddExtensionType(".avif", "image/avif")
	if err != nil {
		log.Fatal(err)
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
			var ErrFileSkip = errors.New("skip mod file")
			modifiedFileEvent := func(path string) error {
				pathBits := strings.Split(path, "\\")
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
				return ErrFileSkip
			}

			// TODO: Expand this to more types of files
			var err error
			for {
				select {
				case ev := <-watcher.Events:
					// TODO: Handle file deletes (and renames more graciously by removing the old version of it)
					if ev.Op&fsnotify.Write == fsnotify.Write {
						err = modifiedFileEvent(ev.Name)
						if err != ErrFileSkip {
							log.Println("modified file:", ev.Name)
						} else {
							err = nil
						}
					} else if ev.Op&fsnotify.Create == fsnotify.Create {
						log.Println("new file:", ev.Name)
						err = modifiedFileEvent(ev.Name)
					} else {
						log.Println("unknown event:", ev)
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

	/*if err = c.StaticFiles.GenJS(); err != nil {
		c.LogError(err)
	}*/

	log.Print("Checking for init tasks")
	if err = sched(); err != nil {
		c.LogError(err)
	}

	log.Print("Initialising the task system")

	// Thumbnailer goroutine, we only want one image being thumbnailed at a time, otherwise they might wind up consuming all the CPU time and leave no resources left to service the actual requests
	// TODO: Could we expand this to attachments and other things too?
	thumbChan := make(chan bool)
	go c.ThumbTask(thumbChan)
	go tickLoop(thumbChan)

	// Resource Management Goroutine
	go func() {
		uc := c.Users.GetCache()
		tc := c.Topics.GetCache()
		if uc == nil && tc == nil {
			return
		}

		var lastEvictedCount int
		var couldNotDealloc bool
		secondTicker := time.NewTicker(time.Second)
		for {
			select {
			case <-secondTicker.C:
				// TODO: Add a LastRequested field to cached User structs to avoid evicting the same things which wind up getting loaded again anyway?
				if uc != nil {
					ucap := uc.GetCapacity()
					if uc.Length() <= ucap || c.Users.Count() <= ucap {
						couldNotDealloc = false
						continue
					}
					lastEvictedCount = uc.DeallocOverflow(couldNotDealloc)
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
	newServer := func(addr string, handler http.Handler) *http.Server {
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
	/*if c.Dev.QuicPort != 0 {
		sQuicPort := strconv.Itoa(c.Dev.QuicPort)
		log.Print("Listening on quic port " + sQuicPort)
		go func() {
			c.StoppedServer(http3.ListenAndServeQUIC(":"+sQuicPort, c.Config.SslFullchain, c.Config.SslPrivkey, router))
		}()
	}*/

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
