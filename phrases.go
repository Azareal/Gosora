/*
*
* Gosora Phrase System
* Copyright Azareal 2017 - 2018
*
 */
package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// TODO: Let the admin edit phrases from inside the Control Panel? How should we persist these? Should we create a copy of the langpack or edit the primaries? Use the changeLangpack mutex for this?
// nolint Be quiet megacheck, this *is* used
var currentLanguage = "english"
var currentLangPack atomic.Value
var langpackCount int // TODO: Use atomics for this

// We'll be implementing the level phrases in the software proper very very soon!
type LevelPhrases struct {
	Level    string
	LevelMax string // ? Add a max level setting?

	// Override the phrase for individual levels, if the phrases exist
	Levels []string // index = level
}

type LanguagePack struct {
	Name          string
	Phrases       map[string]string // Should we use a sync map or a struct for these? It would be nice, if we could keep all the phrases consistent.
	Levels        LevelPhrases
	GlobalPerms   map[string]string
	LocalPerms    map[string]string
	SettingLabels map[string]string
}

// TODO: Add the ability to edit language JSON files from the Control Panel and automatically scan the files for changes
// TODO: Move the english language pack into a JSON file and load that on start-up
////var langpacks = map[string]*LanguagePack
var langpacks sync.Map // nolint it is used

func initPhrases() error {
	log.Print("Loading the language packs")
	err := filepath.Walk("./langs", func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var ext = filepath.Ext("/langs/" + path)
		if ext != ".json" {
			if dev.DebugMode {
				log.Print("Found a " + ext + "in /langs/")
			}
			return nil
		}

		var langPack LanguagePack
		err = json.Unmarshal(data, &langPack)
		if err != nil {
			return err
		}

		log.Print("Adding the '" + langPack.Name + "' language pack")
		langpacks.Store(langPack.Name, &langPack)
		langpackCount++

		return nil
	})

	if err != nil {
		return err
	}
	if langpackCount == 0 {
		return errors.New("You don't have any language packs")
	}

	langPack, ok := langpacks.Load(currentLanguage)
	if !ok {
		return errors.New("Couldn't find the " + currentLanguage + " language pack")
	}
	currentLangPack.Store(langPack)
	return nil
}

func LoadLangPack(name string) error {
	_ = name
	return nil
}

// We might not need to use a mutex for this, we shouldn't need to change the phrases after start-up, and when we do we could overwrite the entire map
func GetPhrase(name string) (string, bool) {
	res, ok := currentLangPack.Load().(*LanguagePack).Phrases[name]
	return res, ok
}

func GetGlobalPermPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).GlobalPerms[name]
	if !ok {
		return "{name}"
	}
	return res
}

func GetLocalPermPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).LocalPerms[name]
	if !ok {
		return "{name}"
	}
	return res
}

func GetSettingLabel(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).SettingLabels[name]
	if !ok {
		return "{name}"
	}
	return res
}

// Is this a copy of the map or a pointer to it? We don't want to accidentally create a race condition
func GetAllSettingLabels() map[string]string {
	return currentLangPack.Load().(*LanguagePack).SettingLabels
}

// Use runtime reflection for updating phrases?
func AddPhrase() {

}
func UpdatePhrase() {

}
func DeletePhrase() {

}

// TODO: Use atomics to store the pointer of the current active langpack?
// nolint
func ChangeLanguagePack(name string) (exists bool) {
	pack, ok := langpacks.Load(name)
	if !ok {
		return false
	}
	currentLangPack.Store(pack)
	return true
}
