/* Copyright Azareal 2016 - 2017 */
package main
import "log"

var plugins map[string]*Plugin = make(map[string]*Plugin)
var hooks map[string][]func(interface{})interface{} = make(map[string][]func(interface{})interface{})
var vhooks map[string]func(...interface{})interface{} = make(map[string]func(...interface{})interface{})

type Plugin struct
{
	UName string
	Name string
	Author string
	URL string
	Settings string
	Active bool
	Tag string
	Type string
	Init func()
	Activate func()error
	Deactivate func()
	
	Hooks map[string]int
}

func NewPlugin(uname string, name string, author string, url string, settings string, tag string, ptype string, init func(), activate func()error, deactivate func()) *Plugin {
	return &Plugin{
		UName: uname,
		Name: name,
		Author: author,
		URL: url,
		Settings: settings,
		Tag: tag,
		Type: ptype,
		Init: init,
		Activate: activate,
		Deactivate: deactivate,
		
		/*
		The Active field should never be altered by a plugin. It's used internally by the software to determine whether an admin has enabled a plugin or not and whether to run it. This will be overwritten by the user's preference.
		*/
		Active: false,
		Hooks: make(map[string]int),
	}
}

func (plugin *Plugin) AddHook(name string, handler interface{}) {
	switch h := handler.(type) {
		case func(interface{})interface{}:
			if len(hooks[name]) == 0 {
				var hookSlice []func(interface{})interface{}
				hookSlice = append(hookSlice, h)
				hooks[name] = hookSlice
			} else {
				hooks[name] = append(hooks[name], h)
			}
			plugin.Hooks[name] = len(hooks[name])
		case func(...interface{}) interface{}:
			vhooks[name] = h
			plugin.Hooks[name] = 0
		default:
			panic("I don't recognise this kind of handler!") // Should this be an error for the plugin instead of a panic()?
	}
}

func (plugin *Plugin) RemoveHook(name string, handler interface{}) {
	switch handler.(type) {
		case func(interface{})interface{}:
			key := plugin.Hooks[name]
			hook := hooks[name]
			if len(hook) == 1 {
				hook = []func(interface{})interface{}{}
			} else {
				hook = append(hook[:key], hook[key + 1:]...)
			}
			hooks[name] = hook
		case func(...interface{}) interface{}:
			delete(vhooks, name)
		default:
			panic("I don't recognise this kind of handler!") // Should this be an error for the plugin instead of a panic()?
	}
	delete(plugin.Hooks, name)
}

var plugins_inited bool = false
func init_plugins() {
	for name, body := range plugins {
		log.Print("Added plugin " + name)
		if body.Active {
			log.Print("Initialised plugin " + name)
			plugins[name].Init()
		}
	}
	plugins_inited = true
}

func run_hook(name string, data interface{}) interface{} {
	for _, hook := range hooks[name] {
		data = hook(data)
	}
	return data
}

func run_vhook(name string, data ...interface{}) interface{} {
	return vhooks[name](data...)
}

func run_vhook_noreturn(name string, data ...interface{}) {
	_ = vhooks[name](data...)
}