/*
*
* Gosora Plugin System
* Copyright Azareal 2016 - 2018
*
 */
package common

import (
	"database/sql"
	"log"
	"net/http"

	"../query_gen/lib"
)

type PluginList map[string]*Plugin

var Plugins PluginList = make(map[string]*Plugin)

// Hooks with a single argument. Is this redundant? Might be useful for inlining, as variadics aren't inlined? Are closures even inlined to begin with?
var Hooks = map[string][]func(interface{}) interface{}{
	"forums_frow_assign":       nil,
	"topic_create_frow_assign": nil,
}

// Hooks with a variable number of arguments
var Vhooks = map[string]func(...interface{}) interface{}{
	"forum_trow_assign":       nil,
	"topics_topic_row_assign": nil,
	//"topics_user_row_assign": nil,
	"topic_reply_row_assign": nil,
	"create_group_preappend": nil, // What is this? Investigate!
	"topic_create_pre_loop":  nil,
}

// Hooks with a variable number of arguments and return values for skipping the parent function and propagating an error upwards
var VhookSkippable = map[string]func(...interface{}) (bool, RouteError){
	"simple_forum_check_pre_perms": nil,
	"forum_check_pre_perms":        nil,
}

//var vhookErrorable = map[string]func(...interface{}) (interface{}, RouteError){}

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
	HeaderVars() *HeaderVars
	CurrentUser() *User
	GetExtData(name string) interface{}
	SetExtData(name string, contents interface{})
}

// Coming Soon:
var messageHooks = map[string][]func(Message, PageInt, ...interface{}) interface{}{
	"topic_reply_row_assign": nil,
}

// Hooks which take in and spit out a string. This is usually used for parser components
var Sshooks = map[string][]func(string) string{
	"preparse_preassign": nil,
	"parse_assign":       nil,
}

// The hooks which run before the template is rendered for a route
var PreRenderHooks = map[string][]func(http.ResponseWriter, *http.Request, *User, interface{}) bool{
	"pre_render": nil,

	"pre_render_forum_list":   nil,
	"pre_render_view_forum":   nil,
	"pre_render_topic_list":   nil,
	"pre_render_view_topic":   nil,
	"pre_render_profile":      nil,
	"pre_render_custom_page":  nil,
	"pre_render_overview":     nil,
	"pre_render_create_topic": nil,

	"pre_render_account_own_edit_critical": nil,
	"pre_render_account_own_edit_avatar":   nil,
	"pre_render_account_own_edit_username": nil,
	"pre_render_account_own_edit_email":    nil,
	"pre_render_login":                     nil,
	"pre_render_register":                  nil,
	"pre_render_ban":                       nil,
	"pre_render_ips":                       nil,

	"pre_render_panel_dashboard":             nil,
	"pre_render_panel_forums":                nil,
	"pre_render_panel_delete_forum":          nil,
	"pre_render_panel_edit_forum":            nil,
	"pre_render_panel_analytics_views":       nil,
	"pre_render_panel_analytics_routes":      nil,
	"pre_render_panel_analytics_agents":      nil,
	"pre_render_panel_analytics_route_views": nil,
	"pre_render_panel_analytics_agent_views": nil,
	"pre_render_panel_settings":              nil,
	"pre_render_panel_setting":               nil,
	"pre_render_panel_word_filters":          nil,
	"pre_render_panel_word_filters_edit":     nil,
	"pre_render_panel_plugins":               nil,
	"pre_render_panel_users":                 nil,
	"pre_render_panel_edit_user":             nil,
	"pre_render_panel_groups":                nil,
	"pre_render_panel_edit_group":            nil,
	"pre_render_panel_edit_group_perms":      nil,
	"pre_render_panel_themes":                nil,
	"pre_render_panel_modlogs":               nil,

	"pre_render_error":          nil, // Note: This hook isn't run for a few errors whose templates are computed at startup and reused, such as InternalError. This hook is also not available in JS mode.
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

	Init       func() error
	Activate   func() error
	Deactivate func()
	Install    func() error
	Uninstall  func() error

	Hooks map[string]int
	Data  interface{} // Usually used for hosting the VMs / reusable elements of non-native plugins
}

type ExtendStmts struct {
	getPlugins *sql.Stmt
}

var extendStmts ExtendStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		extendStmts = ExtendStmts{
			getPlugins: acc.Select("plugins").Columns("uname, active, installed").Prepare(),
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

func NewPlugin(uname string, name string, author string, url string, settings string, tag string, ptype string, init func() error, activate func() error, deactivate func(), install func() error, uninstall func() error) *Plugin {
	return &Plugin{
		UName:       uname,
		Name:        name,
		Author:      author,
		URL:         url,
		Settings:    settings,
		Tag:         tag,
		Type:        ptype,
		Installable: (install != nil),
		Init:        init,
		Activate:    activate,
		Deactivate:  deactivate,
		Install:     install,
		//Uninstall: uninstall,

		/*
			The Active field should never be altered by a plugin. It's used internally by the software to determine whether an admin has enabled a plugin or not and whether to run it. This will be overwritten by the user's preference.
		*/
		Active:    false,
		Installed: false,
		Hooks:     make(map[string]int),
	}
}

// ? - Is this racey?
func (plugin *Plugin) AddHook(name string, handler interface{}) {
	switch h := handler.(type) {
	case func(interface{}) interface{}:
		if len(Hooks[name]) == 0 {
			var hookSlice []func(interface{}) interface{}
			hookSlice = append(hookSlice, h)
			Hooks[name] = hookSlice
		} else {
			Hooks[name] = append(Hooks[name], h)
		}
		plugin.Hooks[name] = len(Hooks[name])
	case func(string) string:
		if len(Sshooks[name]) == 0 {
			var hookSlice []func(string) string
			hookSlice = append(hookSlice, h)
			Sshooks[name] = hookSlice
		} else {
			Sshooks[name] = append(Sshooks[name], h)
		}
		plugin.Hooks[name] = len(Sshooks[name])
	case func(http.ResponseWriter, *http.Request, *User, interface{}) bool:
		if len(PreRenderHooks[name]) == 0 {
			var hookSlice []func(http.ResponseWriter, *http.Request, *User, interface{}) bool
			hookSlice = append(hookSlice, h)
			PreRenderHooks[name] = hookSlice
		} else {
			PreRenderHooks[name] = append(PreRenderHooks[name], h)
		}
		plugin.Hooks[name] = len(PreRenderHooks[name])
	case func(...interface{}) interface{}:
		Vhooks[name] = h
		plugin.Hooks[name] = 0
	case func(...interface{}) (bool, RouteError):
		VhookSkippable[name] = h
		plugin.Hooks[name] = 0
	default:
		panic("I don't recognise this kind of handler!") // Should this be an error for the plugin instead of a panic()?
	}
}

// ? - Is this racey?
func (plugin *Plugin) RemoveHook(name string, handler interface{}) {
	switch handler.(type) {
	case func(interface{}) interface{}:
		key := plugin.Hooks[name]
		hook := Hooks[name]
		if len(hook) == 1 {
			hook = []func(interface{}) interface{}{}
		} else {
			hook = append(hook[:key], hook[key+1:]...)
		}
		Hooks[name] = hook
	case func(string) string:
		key := plugin.Hooks[name]
		hook := Sshooks[name]
		if len(hook) == 1 {
			hook = []func(string) string{}
		} else {
			hook = append(hook[:key], hook[key+1:]...)
		}
		Sshooks[name] = hook
	case func(http.ResponseWriter, *http.Request, *User, interface{}) bool:
		key := plugin.Hooks[name]
		hook := PreRenderHooks[name]
		if len(hook) == 1 {
			hook = []func(http.ResponseWriter, *http.Request, *User, interface{}) bool{}
		} else {
			hook = append(hook[:key], hook[key+1:]...)
		}
		PreRenderHooks[name] = hook
	case func(...interface{}) interface{}:
		delete(Vhooks, name)
	case func(...interface{}) (bool, RouteError):
		delete(VhookSkippable, name)
	default:
		panic("I don't recognise this kind of handler!") // Should this be an error for the plugin instead of a panic()?
	}
	delete(plugin.Hooks, name)
}

var PluginsInited = false

func InitPlugins() {
	for name, body := range Plugins {
		log.Printf("Added plugin '%s'", name)
		if body.Active {
			log.Printf("Initialised plugin '%s'", name)
			if Plugins[name].Init != nil {
				err := Plugins[name].Init()
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
func RunHook(name string, data interface{}) interface{} {
	for _, hook := range Hooks[name] {
		data = hook(data)
	}
	return data
}

func RunHookNoreturn(name string, data interface{}) {
	for _, hook := range Hooks[name] {
		_ = hook(data)
	}
}

func RunVhook(name string, data ...interface{}) interface{} {
	return Vhooks[name](data...)
}

func RunVhookSkippable(name string, data ...interface{}) (bool, RouteError) {
	return VhookSkippable[name](data...)
}

func RunVhookNoreturn(name string, data ...interface{}) {
	_ = Vhooks[name](data...)
}

// Trying to get a teeny bit of type-safety where-ever possible, especially for such a critical set of hooks
func RunSshook(name string, data string) string {
	for _, hook := range Sshooks[name] {
		data = hook(data)
	}
	return data
}

func RunPreRenderHook(name string, w http.ResponseWriter, r *http.Request, user *User, data interface{}) (halt bool) {
	// This hook runs on ALL pre_render hooks
	for _, hook := range PreRenderHooks["pre_render"] {
		if hook(w, r, user, data) {
			return true
		}
	}

	// The actual pre_render hook
	for _, hook := range PreRenderHooks[name] {
		if hook(w, r, user, data) {
			return true
		}
	}
	return false
}
