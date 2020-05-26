package main // import "github.com/Azareal/Gosora/hook_stub_gen"

import (
	"fmt"
	"log"
	"strings"
	"runtime/debug"
	
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
		hookVars.Hooks = append(hookVars.Hooks, h.Hook{name, params, params2, ret, htype, true, multiHook, skip, defaultRet,pure})
	}

	h.AddHooks(add)
	h.Write(hookVars)
	log.Println("Successfully generated the hooks")
}