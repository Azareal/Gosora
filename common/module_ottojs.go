/*
*
*	OttoJS Plugin Module
*	Copyright Azareal 2016 - 2018
*
 */
package common

import (
	"errors"

	"github.com/robertkrimen/otto"
)

type OttoPluginLang struct {
	vm      *otto.Otto
	plugins map[string]*otto.Script
	vars    map[string]*otto.Object
}

func init() {
	pluginLangs["ottojs"] = &OttoPluginLang{
		plugins: make(map[string]*otto.Script),
		vars:    make(map[string]*otto.Object),
	}
}

func (js *OttoPluginLang) Init() (err error) {
	js.vm = otto.New()
	js.vars["current_page"], err = js.vm.Object(`var current_page = {}`)
	return err
}

func (js *OttoPluginLang) GetName() string {
	return "ottojs"
}

func (js *OttoPluginLang) GetExts() []string {
	return []string{".js"}
}

func (js *OttoPluginLang) AddPlugin(meta PluginMeta) (plugin *Plugin, err error) {
	script, err := js.vm.Compile("./extend/"+meta.UName+"/"+meta.Main, nil)
	if err != nil {
		return nil, err
	}

	var pluginInit = func() error {
		retValue, err := js.vm.Run(script)
		if err != nil {
			return err
		}
		if retValue.IsString() {
			ret, err := retValue.ToString()
			if err != nil {
				return err
			}
			if ret != "" {
				return errors.New(ret)
			}
		}
		return nil
	}
	var pluginActivate func() error
	var pluginDeactivate func()
	var pluginInstall func() error
	var pluginUninstall func() error

	plugin = NewPlugin(meta.UName, meta.Name, meta.Author, meta.URL, meta.Settings, meta.Tag, "ottojs", pluginInit, pluginActivate, pluginDeactivate, pluginInstall, pluginUninstall)

	plugin.Data = script
	return plugin, nil
}

/*func (js *OttoPluginLang) addHook(hook string, plugin string) {
	hooks[hook] = func(data interface{}) interface{} {
		switch d := data.(type) {
		case Page:
			currentPage := js.vars["current_page"]
			currentPage.Set("Title", d.Title)
		case TopicPage:

		case ProfilePage:

		case Reply:

		default:
			log.Print("Not a valid JS datatype")
		}
	}
}*/
