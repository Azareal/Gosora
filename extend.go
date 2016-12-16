/* Copyright Azareal 2016 - 2017 */
package main

var plugins map[string]Plugin = make(map[string]Plugin)
var hooks map[string]func(interface{})interface{} = make(map[string]func(interface{})interface{})
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
}

func add_hook(name string, handler func(interface{})interface{}) {
	hooks[name] = handler
}

func remove_hook(name string) {
	delete(hooks, name)
}

func run_hook(name string, data interface{}) interface{} {
	return hooks[name](data)
}

func add_vhook(name string, handler func(...interface{}) interface{}) {
	vhooks[name] = handler
}

func remove_vhook(name string) {
	delete(vhooks, name)
}

func run_vhook(name string, data ...interface{}) interface{} {
	return vhooks[name](data...)
}

func run_vhook_noreturn(name string, data ...interface{}) {
	_ = vhooks[name](data...)
}