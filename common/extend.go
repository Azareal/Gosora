/*
*
* Gosora Plugin System
* Copyright Azareal 2016 - 2020
*
 */
package common

// TODO: Break this file up into multiple files to make it easier to maintain
import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/Azareal/Gosora/query_gen"
)

var ErrPluginNotInstallable = errors.New("This plugin is not installable")

type PluginList map[string]*Plugin

// TODO: Have a proper store rather than a map?
var Plugins PluginList = make(map[string]*Plugin)

func (list PluginList) Add(plugin *Plugin) {
	buildPlugin(plugin)
	list[plugin.UName] = plugin
}

func buildPlugin(plugin *Plugin) {
	plugin.Installable = (plugin.Install != nil)
	/*
		The Active field should never be altered by a plugin. It's used internally by the software to determine whether an admin has enabled a plugin or not and whether to run it. This will be overwritten by the user's preference.
	*/
	plugin.Active = false
	plugin.Installed = false
	plugin.Hooks = make(map[string]int)
	plugin.Data = nil
}

var hookTableBox atomic.Value

// ! HookTable is a work in progress, do not use it yet
// TODO: Test how fast it is to indirect hooks off the hook table as opposed to using them normally or using an interface{} for the hooks
// TODO: Can we filter the HookTable for each request down to only hooks the request actually uses?
// TODO: Make the RunXHook functions methods on HookTable
// TODO: Have plugins update hooks on a mutex guarded map and create a copy of that map in a serial global goroutine which gets thrown in the atomic.Value
type HookTable struct {
	Hooks           map[string][]func(interface{}) interface{}
	Vhooks          map[string]func(...interface{}) interface{}
	VhookSkippable_ map[string]func(...interface{}) (bool, RouteError)
	Sshooks         map[string][]func(string) string
	PreRenderHooks  map[string][]func(http.ResponseWriter, *http.Request, *User, interface{}) bool

	// For future use:
	messageHooks map[string][]func(Message, PageInt, ...interface{}) interface{}
}

func init() {
	RebuildHookTable()
}

// For extend.go use only, access this via GetHookTable() elsewhere
var hookTable = &HookTable{
	map[string][]func(interface{}) interface{}{
		"forums_frow_assign":       nil,
		"topic_create_frow_assign": nil,
	},
	map[string]func(...interface{}) interface{}{
		"forum_trow_assign":       nil,
		"topics_topic_row_assign": nil,
		//"topics_user_row_assign": nil,
		"topic_reply_row_assign": nil,
		"create_group_preappend": nil, // What is this? Investigate!
		"topic_create_pre_loop":  nil,

		"router_end": nil,
	},
	map[string]func(...interface{}) (bool, RouteError){
		"simple_forum_check_pre_perms": nil,
		"forum_check_pre_perms":        nil,

		"route_topic_list_start": nil,
		"route_forum_list_start": nil,

		"action_end_create_topic":             nil,
		"action_end_edit_topic":nil,
		"action_end_delete_topic":nil,
		"action_end_lock_topic":nil,
		"action_end_unlock_topic": nil,
		"action_end_stick_topic": nil,
		"action_end_unstick_topic": nil,
		"action_end_move_topic": nil,
		"action_end_like_topic":nil,

		"action_end_create_reply":             nil,
		"action_end_edit_reply":               nil,
		"action_end_delete_reply":             nil,
		"action_end_add_attach_to_reply":      nil,
		"action_end_remove_attach_from_reply": nil,

		"action_end_like_reply":nil,

		"action_end_ban_user":nil,
		"action_end_unban_user":nil,
		"action_end_activate_user":nil,

		"router_after_filters": nil,
		"router_pre_route":     nil,

		"tasks_tick_topic_list": nil,
		"tasks_tick_widget_wol": nil,
	},
	map[string][]func(string) string{
		"preparse_preassign": nil,
		"parse_assign":       nil,
	},
	nil,
	nil,
}
var hookTableUpdateMutex sync.Mutex

func RebuildHookTable() {
	hookTableUpdateMutex.Lock()
	defer hookTableUpdateMutex.Unlock()
	unsafeRebuildHookTable()
}

func unsafeRebuildHookTable() {
	ihookTable := new(HookTable)
	*ihookTable = *hookTable
	hookTableBox.Store(ihookTable)
}

func GetHookTable() *HookTable {
	return hookTableBox.Load().(*HookTable)
}

// Hooks with a single argument. Is this redundant? Might be useful for inlining, as variadics aren't inlined? Are closures even inlined to begin with?
func (table *HookTable) Hook(name string, data interface{}) interface{} {
	hooks, ok := table.Hooks[name]
	if ok {
		for _, hook := range hooks {
			data = hook(data)
		}
	}
	return data
}

// To cover the case in routes/topic.go's CreateTopic route, we could probably obsolete this use and replace it
func (table *HookTable) HookSkippable(name string, data interface{}) (skip bool) {
	hooks, ok := table.Hooks[name]
	if ok {
		for _, hook := range hooks {
			skip = hook(data).(bool)
			if skip {
				break
			}
		}
	}
	return skip
}

// Hooks with a variable number of arguments
// TODO: Use RunHook semantics to allow multiple lined up plugins / modules their turn?
func (table *HookTable) Vhook(name string, data ...interface{}) interface{} {
	hook := table.Vhooks[name]
	if hook != nil {
		return hook(data...)
	}
	return nil
}

func (table *HookTable) VhookNoRet(name string, data ...interface{}) {
	hook := table.Vhooks[name]
	if hook != nil {
		_ = hook(data...)
	}
}

// TODO: Find a better way of doing this
func (table *HookTable) VhookNeedHook(name string, data ...interface{}) (ret interface{}, hasHook bool) {
	hook := table.Vhooks[name]
	if hook != nil {
		return hook(data...), true
	}
	return nil, false
}

// Hooks with a variable number of arguments and return values for skipping the parent function and propagating an error upwards
func (table *HookTable) VhookSkippable(name string, data ...interface{}) (bool, RouteError) {
	hook := table.VhookSkippable_[name]
	if hook != nil {
		return hook(data...)
	}
	return false, nil
}

// Hooks which take in and spit out a string. This is usually used for parser components
// Trying to get a teeny bit of type-safety where-ever possible, especially for such a critical set of hooks
func (table *HookTable) Sshook(name string, data string) string {
	ssHooks, ok := table.Sshooks[name]
	if ok {
		for _, hook := range ssHooks {
			data = hook(data)
		}
	}
	return data
}

//var vhookErrorable = map[string]func(...interface{}) (interface{}, RouteError){}

var taskHooks = map[string][]func() error{
	"before_half_second_tick":    nil,
	"after_half_second_tick":     nil,
	"before_second_tick":         nil,
	"after_second_tick":          nil,
	"before_fifteen_minute_tick": nil,
	"after_fifteen_minute_tick":  nil,
}

// Coming Soon:
type Message interface {
	ID() int
	Poster() int
	Contents() string
	ParsedContents() string
}

// While the idea is nice, this might result in too much code duplication, as we have seventy billion page structs, what else could we do to get static typing with these in plugins?
type PageInt interface {
	Title() string
	Header() *Header
	CurrentUser() *User
	GetExtData(name string) interface{}
	SetExtData(name string, contents interface{})
}

// Coming Soon:
var messageHooks = map[string][]func(Message, PageInt, ...interface{}) interface{}{
	"topic_reply_row_assign": nil,
}

// The hooks which run before the template is rendered for a route
var PreRenderHooks = map[string][]func(http.ResponseWriter, *http.Request, *User, interface{}) bool{
	"pre_render": nil,

	"pre_render_forums":       nil,
	"pre_render_forum":        nil,
	"pre_render_topics":       nil,
	"pre_render_topic":        nil,
	"pre_render_profile":      nil,
	"pre_render_custom_page":  nil,
	"pre_render_tmpl_page":    nil,
	"pre_render_overview":     nil,
	"pre_render_create_topic": nil,

	"pre_render_account_own_edit":           nil,
	"pre_render_account_own_edit_password":  nil,
	"pre_render_account_own_edit_mfa":       nil,
	"pre_render_account_own_edit_mfa_setup": nil,
	"pre_render_account_own_edit_email":     nil,
	"pre_render_level_list":                 nil,
	"pre_render_login":                      nil,
	"pre_render_login_mfa_verify":           nil,
	"pre_render_register":                   nil,
	"pre_render_ban":                        nil,
	"pre_render_ip_search":                  nil,

	"pre_render_panel_dashboard":        nil,
	"pre_render_panel_forums":           nil,
	"pre_render_panel_delete_forum":     nil,
	"pre_render_panel_forum_edit":       nil,
	"pre_render_panel_forum_edit_perms": nil,

	"pre_render_panel_analytics_views":          nil,
	"pre_render_panel_analytics_routes":         nil,
	"pre_render_panel_analytics_agents":         nil,
	"pre_render_panel_analytics_systems":        nil,
	"pre_render_panel_analytics_referrers":      nil,
	"pre_render_panel_analytics_route_views":    nil,
	"pre_render_panel_analytics_agent_views":    nil,
	"pre_render_panel_analytics_system_views":   nil,
	"pre_render_panel_analytics_referrer_views": nil,

	"pre_render_panel_settings":          nil,
	"pre_render_panel_setting":           nil,
	"pre_render_panel_word_filters":      nil,
	"pre_render_panel_word_filters_edit": nil,
	"pre_render_panel_plugins":           nil,
	"pre_render_panel_users":             nil,
	"pre_render_panel_user_edit":         nil,
	"pre_render_panel_groups":            nil,
	"pre_render_panel_group_edit":        nil,
	"pre_render_panel_group_edit_perms":  nil,
	"pre_render_panel_themes":            nil,
	"pre_render_panel_modlogs":           nil,

	"pre_render_error": nil, // Note: This hook isn't run for a few errors whose templates are computed at startup and reused, such as InternalError. This hook is also not available in JS mode.
	// ^-- I don't know if it's run for InternalError, but it isn't computed at startup anymore
	"pre_render_security_error": nil,
}

// ? - Should we make this an interface which plugins implement instead?
// Plugin is a struct holding the metadata for a plugin, along with a few of it's primary handlers.
type Plugin struct {
	UName       string
	Name        string
	Author      string
	URL         string
	Settings    string
	Active      bool
	Tag         string
	Type        string
	Installable bool
	Installed   bool

	Init       func(plugin *Plugin) error
	Activate   func(plugin *Plugin) error
	Deactivate func(plugin *Plugin) // TODO: We might want to let this return an error?
	Install    func(plugin *Plugin) error
	Uninstall  func(plugin *Plugin) error // TODO: I'm not sure uninstall is implemented

	Hooks map[string]int
	Data  interface{} // Usually used for hosting the VMs / reusable elements of non-native plugins
}

func (plugin *Plugin) BypassActive() (active bool, err error) {
	err = extendStmts.isActive.QueryRow(plugin.UName).Scan(&active)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return active, nil
}

func (plugin *Plugin) InDatabase() (exists bool, err error) {
	var sink bool
	err = extendStmts.isActive.QueryRow(plugin.UName).Scan(&sink)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return err == nil, nil
}

// TODO: Silently add to the database, if it doesn't exist there rather than forcing users to call AddToDatabase instead?
func (plugin *Plugin) SetActive(active bool) (err error) {
	_, err = extendStmts.setActive.Exec(active, plugin.UName)
	if err == nil {
		plugin.Active = active
	}
	return err
}

// TODO: Silently add to the database, if it doesn't exist there rather than forcing users to call AddToDatabase instead?
func (plugin *Plugin) SetInstalled(installed bool) (err error) {
	if !plugin.Installable {
		return ErrPluginNotInstallable
	}
	_, err = extendStmts.setInstalled.Exec(installed, plugin.UName)
	if err == nil {
		plugin.Installed = installed
	}
	return err
}

func (plugin *Plugin) AddToDatabase(active bool, installed bool) (err error) {
	_, err = extendStmts.add.Exec(plugin.UName, active, installed)
	if err == nil {
		plugin.Active = active
		plugin.Installed = installed
	}
	return err
}

type ExtendStmts struct {
	getPlugins *sql.Stmt

	isActive     *sql.Stmt
	setActive    *sql.Stmt
	setInstalled *sql.Stmt
	add          *sql.Stmt
}

var extendStmts ExtendStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		extendStmts = ExtendStmts{
			getPlugins: acc.Select("plugins").Columns("uname, active, installed").Prepare(),

			isActive:     acc.Select("plugins").Columns("active").Where("uname = ?").Prepare(),
			setActive:    acc.Update("plugins").Set("active = ?").Where("uname = ?").Prepare(),
			setInstalled: acc.Update("plugins").Set("installed = ?").Where("uname = ?").Prepare(),
			add:          acc.Insert("plugins").Columns("uname, active, installed").Fields("?,?,?").Prepare(),
		}
		return acc.FirstError()
	})
}

func InitExtend() (err error) {
	err = InitPluginLangs()
	if err != nil {
		return err
	}
	return Plugins.Load()
}

// Load polls the database to see which plugins have been activated and which have been installed
func (plugins PluginList) Load() error {
	rows, err := extendStmts.getPlugins.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var uname string
	var active, installed bool
	for rows.Next() {
		err = rows.Scan(&uname, &active, &installed)
		if err != nil {
			return err
		}

		// Was the plugin deleted at some point?
		plugin, ok := plugins[uname]
		if !ok {
			continue
		}
		plugin.Active = active
		plugin.Installed = installed
		plugins[uname] = plugin
	}
	return rows.Err()
}

// ? - Is this racey?
// TODO: Generate the cases in this switch
func (plugin *Plugin) AddHook(name string, handler interface{}) {
	hookTableUpdateMutex.Lock()
	defer hookTableUpdateMutex.Unlock()

	switch h := handler.(type) {
	case func(interface{}) interface{}:
		if len(hookTable.Hooks[name]) == 0 {
			hookTable.Hooks[name] = []func(interface{}) interface{}{}
		}
		hookTable.Hooks[name] = append(hookTable.Hooks[name], h)
		plugin.Hooks[name] = len(hookTable.Hooks[name]) - 1
	case func(string) string:
		if len(hookTable.Sshooks[name]) == 0 {
			hookTable.Sshooks[name] = []func(string) string{}
		}
		hookTable.Sshooks[name] = append(hookTable.Sshooks[name], h)
		plugin.Hooks[name] = len(hookTable.Sshooks[name]) - 1
	case func(http.ResponseWriter, *http.Request, *User, interface{}) bool:
		if len(PreRenderHooks[name]) == 0 {
			PreRenderHooks[name] = []func(http.ResponseWriter, *http.Request, *User, interface{}) bool{}
		}
		PreRenderHooks[name] = append(PreRenderHooks[name], h)
		plugin.Hooks[name] = len(PreRenderHooks[name]) - 1
	case func() error: // ! We might want a more generic name, as we might use this signature for things other than tasks hooks
		if len(taskHooks[name]) == 0 {
			taskHooks[name] = []func() error{}
		}
		taskHooks[name] = append(taskHooks[name], h)
		plugin.Hooks[name] = len(taskHooks[name]) - 1
	case func(...interface{}) interface{}:
		hookTable.Vhooks[name] = h
		plugin.Hooks[name] = 0
	case func(...interface{}) (bool, RouteError):
		hookTable.VhookSkippable_[name] = h
		plugin.Hooks[name] = 0
	default:
		panic("I don't recognise this kind of handler!") // Should this be an error for the plugin instead of a panic()?
	}
	// TODO: Do this once during plugin activation / deactivation rather than doing it for each hook
	unsafeRebuildHookTable()
}

// ? - Is this racey?
// TODO: Generate the cases in this switch
func (plugin *Plugin) RemoveHook(name string, handler interface{}) {
	hookTableUpdateMutex.Lock()
	defer hookTableUpdateMutex.Unlock()

	key, ok := plugin.Hooks[name]
	if !ok {
		panic("handler not registered as hook")
	}

	switch handler.(type) {
	case func(interface{}) interface{}:
		hook := hookTable.Hooks[name]
		if len(hook) == 1 {
			hook = []func(interface{}) interface{}{}
		} else {
			hook = append(hook[:key], hook[key+1:]...)
		}
		hookTable.Hooks[name] = hook
	case func(string) string:
		hook := hookTable.Sshooks[name]
		if len(hook) == 1 {
			hook = []func(string) string{}
		} else {
			hook = append(hook[:key], hook[key+1:]...)
		}
		hookTable.Sshooks[name] = hook
	case func(http.ResponseWriter, *http.Request, *User, interface{}) bool:
		hook := PreRenderHooks[name]
		if len(hook) == 1 {
			hook = []func(http.ResponseWriter, *http.Request, *User, interface{}) bool{}
		} else {
			hook = append(hook[:key], hook[key+1:]...)
		}
		PreRenderHooks[name] = hook
	case func() error:
		hook := taskHooks[name]
		if len(hook) == 1 {
			hook = []func() error{}
		} else {
			hook = append(hook[:key], hook[key+1:]...)
		}
		taskHooks[name] = hook
	case func(...interface{}) interface{}:
		delete(hookTable.Vhooks, name)
	case func(...interface{}) (bool, RouteError):
		delete(hookTable.VhookSkippable_, name)
	default:
		panic("I don't recognise this kind of handler!") // Should this be an error for the plugin instead of a panic()?
	}
	delete(plugin.Hooks, name)
	// TODO: Do this once during plugin activation / deactivation rather than doing it for each hook
	unsafeRebuildHookTable()
}

// TODO: Add a HasHook method to complete the AddHook, RemoveHook, etc. set?

var PluginsInited = false

func InitPlugins() {
	for name, body := range Plugins {
		log.Printf("Added plugin '%s'", name)
		if body.Active {
			log.Printf("Initialised plugin '%s'", name)
			if body.Init != nil {
				err := body.Init(body)
				if err != nil {
					log.Print(err)
				}
			} else {
				log.Printf("Plugin '%s' doesn't have an initialiser.", name)
			}
		}
	}
	PluginsInited = true
}

// ? - Are the following functions racey?
func RunTaskHook(name string) error {
	for _, hook := range taskHooks[name] {
		err := hook()
		if err != nil {
			return err
		}
	}
	return nil
}

func RunPreRenderHook(name string, w http.ResponseWriter, r *http.Request, user *User, data interface{}) (halt bool) {
	// This hook runs on ALL PreRender hooks
	preRenderHooks, ok := PreRenderHooks["pre_render"]
	if ok {
		for _, hook := range preRenderHooks {
			if hook(w, r, user, data) {
				return true
			}
		}
	}

	// The actual PreRender hook
	preRenderHooks, ok = PreRenderHooks[name]
	if ok {
		for _, hook := range preRenderHooks {
			if hook(w, r, user, data) {
				return true
			}
		}
	}
	return false
}
