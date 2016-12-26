/* Copyright Azareal 2016 - 2017 */
package main
import "github.com/robertkrimen/otto"

var vm *Otto
var js_plugins map[string]*otto.Script = make(map[string]*otto.Script)
var js_vars map[string]*otto.Object = make(map[string]*otto.Object)

func init()
{
	var err error
	vm = otto.New()
	js_vars["current_page"], err = vm.Object(`current_page = {}`)
	if err != nil {
		log.Fatal(err)
	}
}

func js_add_plugin(plugin string) error
{
	script, err := otto.Compile("./extend/" + plugin + ".js")
	if err != nil {
		return err
	}
	vm.Run(script)
	return nil
}

func js_add_hook(hook string, plugin string)
{
	hooks[hook] = func(data interface{}) interface{} {
		switch d := data.(type) {
			case Page:
				current_page := js_vars["current_page"]
				current_page.Set("Title", d.Title)
			case TopicPage:
				
			case ProfilePage:
				
			case Reply:
				
			default:
				log.Print("Not a valid JS datatype")
		}
	}
}

