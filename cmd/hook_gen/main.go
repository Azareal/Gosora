// +build hookgen

package main // import "github.com/Azareal/Gosora/hook_gen"

import (
	"fmt"
	"log"
	"strings"
	"runtime/debug"
	
	_ "github.com/Azareal/Gosora/extend"
	c "github.com/Azareal/Gosora/common"
	h "github.com/Azareal/Gosora/cmd/common_hook_gen"
)

// TODO: Make sure all the errors in this file propagate upwards properly
func main() {
	// Capture panics instead of closing the window at a superhuman speed before the user can read the message on Windows
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
		}
	}()
	
	hooks := make(map[string]int)
	for _, pl := range c.Plugins {
		if len(pl.Meta.Hooks) > 0 {
			for _, hook := range pl.Meta.Hooks {
				hooks[hook]++
			}
			continue
		}
		if pl.Init != nil {
			if e := pl.Init(pl); e != nil {
				log.Print("early plugin init err: ", e)
				return
			}
		}
		if pl.Hooks != nil {
			log.Print("Hooks not nil for ", pl.UName)
			for hook, _ := range pl.Hooks {
				hooks[hook] += 1
			}
		}
	}
	log.Printf("hooks: %+v\n", hooks)
	
	imports := []string{"net/http"}
	hookVars := h.HookVars{imports,nil}
	add := func(name, params, ret, htype string, multiHook, skip bool, defaultRet, pure string) {
		var params2 string
		first := true
		for _, param := range strings.Split(params,",") {
			if !first {
				params2 += ","
			}
			pspl := strings.Split(strings.ReplaceAll(strings.TrimSpace(param),"  "," ")," ")
			params2 += pspl[0]
			first = false
		}
		hookVars.Hooks = append(hookVars.Hooks, h.Hook{name, params, params2, ret, htype, hooks[name] > 0, multiHook, skip, defaultRet, pure})
	}
	
	h.AddHooks(add)
	h.Write(hookVars)
	log.Println("Successfully generated the hooks")
}