package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
)

var pluginLangs = make(map[string]PluginLang)

// For non-native plugins to bind JSON files to. E.g. JS and Lua
type PluginMeta struct {
	UName    string
	Name     string
	Author   string
	URL      string
	Settings string
	Tag      string

	Main  string            // The main file
	Hooks map[string]string // Hooks mapped to functions
}

type PluginLang interface {
	GetName() string
	GetExts() []string

	Init() error
	AddPlugin(meta PluginMeta) (*Plugin, error)
	//AddHook(name string, handler interface{}) error
	//RemoveHook(name string, handler interface{})
	//RunHook(name string, data interface{}) interface{}
	//RunVHook(name string data ...interface{}) interface{}
}

/*
var ext = filepath.Ext(pluginFile.Name())
if ext == ".txt" || ext == ".go" {
	continue
}
*/

func InitPluginLangs() error {
	for _, pluginLang := range pluginLangs {
		pluginLang.Init()
	}

	pluginList, err := GetPluginFiles()
	if err != nil {
		return err
	}

	for _, pluginItem := range pluginList {
		pluginFile, err := ioutil.ReadFile("./extend/" + pluginItem + "/plugin.json")
		if err != nil {
			return err
		}

		var plugin PluginMeta
		err = json.Unmarshal(pluginFile, &plugin)
		if err != nil {
			return err
		}

		if plugin.UName == "" {
			return errors.New("The UName field must not be blank on plugin '" + pluginItem + "'")
		}
		if plugin.Name == "" {
			return errors.New("The Name field must not be blank on plugin '" + pluginItem + "'")
		}
		if plugin.Author == "" {
			return errors.New("The Author field must not be blank on plugin '" + pluginItem + "'")
		}
		if plugin.Main == "" {
			return errors.New("Couldn't find a main file for plugin '" + pluginItem + "'")
		}

		var ext = filepath.Ext(plugin.Main)
		pluginLang, err := ExtToPluginLang(ext)
		if err != nil {
			return err
		}
		pplugin, err := pluginLang.AddPlugin(plugin)
		if err != nil {
			return err
		}
		plugins[plugin.UName] = pplugin
	}
	return nil
}

func GetPluginFiles() (pluginList []string, err error) {
	pluginFiles, err := ioutil.ReadDir("./extend")
	if err != nil {
		return nil, err
	}
	for _, pluginFile := range pluginFiles {
		if !pluginFile.IsDir() {
			continue
		}
		pluginList = append(pluginList, pluginFile.Name())
	}
	return pluginList, nil
}

func ExtToPluginLang(ext string) (PluginLang, error) {
	for _, pluginLang := range pluginLangs {
		for _, registeredExt := range pluginLang.GetExts() {
			if registeredExt == ext {
				return pluginLang, nil
			}
		}
	}
	return nil, errors.New("No plugin lang handlers are capable of handling extension '" + ext + "'")
}
