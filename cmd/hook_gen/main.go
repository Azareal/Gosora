// +build hookgen

package main // import "github.com/Azareal/Gosora/hook_gen"

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	h "github.com/Azareal/Gosora/cmd/common_hook_gen"
	c "github.com/Azareal/Gosora/common"
	_ "github.com/Azareal/Gosora/extend"
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
	hookVars := h.HookVars{imports, nil}
	var params2sb strings.Builder
	add := func(name, params, ret, htype string, multiHook, skip bool, defaultRet, pure string) {
		first := true
		for _, param := range strings.Split(params, ",") {
			if !first {
				params2sb.WriteRune(',')
			}
			pspl := strings.Split(strings.ReplaceAll(strings.TrimSpace(param), "  ", " "), " ")
			params2sb.WriteString(pspl[0])
			first = false
		}
		hookVars.Hooks = append(hookVars.Hooks, h.Hook{name, params, params2sb.String(), ret, htype, hooks[name] > 0, multiHook, skip, defaultRet, pure})
		params2sb.Reset()
	}

	h.AddHooks(add)
	h.Write(hookVars)
	log.Println("Successfully generated the hooks")
}
