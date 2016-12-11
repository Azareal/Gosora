/* Copyright Azareal 2016 - 2017 */
package main

var plugins map[string]Plugin = make(map[string]Plugin)
var hooks map[string]func(interface{})interface{} = make(map[string]func(interface{})interface{})

type Plugin struct
{
	UName string
	Name string
	Author string
	URL string
	Settings string
	Active bool
	Type string
	Init func()
}

func add_hook(name string, handler func(interface{})interface{}) {
	hooks[name] = handler
}

func run_hook(name string, data interface{}) interface{} {
	return hooks[name](data)
}