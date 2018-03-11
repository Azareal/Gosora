/*
*
* Gosora Phrase System
* Copyright Azareal 2017 - 2018
*
 */
package common

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

// TODO: Add a phrase store?
// TODO: Let the admin edit phrases from inside the Control Panel? How should we persist these? Should we create a copy of the langpack or edit the primaries? Use the changeLangpack mutex for this?
// nolint Be quiet megacheck, this *is* used
var currentLangPack atomic.Value
var langPackCount int // TODO: Use atomics for this

// TODO: We'll be implementing the level phrases in the software proper very very soon!
type LevelPhrases struct {
	Level    string
	LevelMax string // ? Add a max level setting?

	// Override the phrase for individual levels, if the phrases exist
	Levels []string // index = level
}

// ! For the sake of thread safety, you must never modify a *LanguagePack directly, but to create a copy of it and overwrite the entry in the sync.Map
type LanguagePack struct {
	Name             string
	Phrases          map[string]string // Should we use a sync map or a struct for these? It would be nice, if we could keep all the phrases consistent.
	Levels           LevelPhrases
	GlobalPerms      map[string]string
	LocalPerms       map[string]string
	SettingLabels    map[string]string
	PermPresets      map[string]string
	Accounts         map[string]string // TODO: Apply these phrases in the software proper
	UserAgents       map[string]string
	OperatingSystems map[string]string
	HumanLanguages   map[string]string
	Errors           map[string]map[string]string // map[category]map[name]value
	PageTitles       map[string]string
	TmplPhrases      map[string]string
	CSSPhrases       map[string]string

	TmplIndicesToPhrases [][][]byte // [tmplID][index]phrase
}

// TODO: Add the ability to edit language JSON files from the Control Panel and automatically scan the files for changes
var langPacks sync.Map                // nolint it is used
var langTmplIndicesToNames [][]string // [tmplID][index]phraseName

func InitPhrases() error {
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
			log.Printf("Found a '%s' in /langs/", ext)
			return nil
		}

		var langPack LanguagePack
		err = json.Unmarshal(data, &langPack)
		if err != nil {
			return err
		}

		langPack.TmplIndicesToPhrases = make([][][]byte, len(langTmplIndicesToNames))
		for tmplID, phraseNames := range langTmplIndicesToNames {
			var phraseSet = make([][]byte, len(phraseNames))
			for index, phraseName := range phraseNames {
				phrase, ok := langPack.TmplPhrases[phraseName]
				if !ok {
					log.Print("Couldn't find template phrase '" + phraseName + "'")
				}
				phraseSet[index] = []byte(phrase)
			}
			langPack.TmplIndicesToPhrases[tmplID] = phraseSet
		}

		log.Print("Adding the '" + langPack.Name + "' language pack")
		langPacks.Store(langPack.Name, &langPack)
		langPackCount++

		return nil
	})

	if err != nil {
		return err
	}
	if langPackCount == 0 {
		return errors.New("You don't have any language packs")
	}

	langPack, ok := langPacks.Load(Site.Language)
	if !ok {
		return errors.New("Couldn't find the " + Site.Language + " language pack")
	}
	currentLangPack.Store(langPack)
	return nil
}

// TODO: Implement this
func LoadLangPack(name string) error {
	_ = name
	return nil
}

// TODO: Implement this
func SaveLangPack(langPack *LanguagePack) error {
	_ = langPack
	return nil
}

func GetPhrase(name string) (string, bool) {
	res, ok := currentLangPack.Load().(*LanguagePack).Phrases[name]
	return res, ok
}

func GetGlobalPermPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).GlobalPerms[name]
	if !ok {
		return getPhrasePlaceholder("perms", name)
	}
	return res
}

func GetLocalPermPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).LocalPerms[name]
	if !ok {
		return getPhrasePlaceholder("perms", name)
	}
	return res
}

func GetSettingLabel(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).SettingLabels[name]
	if !ok {
		return getPhrasePlaceholder("settings", name)
	}
	return res
}

func GetAllSettingLabels() map[string]string {
	return currentLangPack.Load().(*LanguagePack).SettingLabels
}

func GetAllPermPresets() map[string]string {
	return currentLangPack.Load().(*LanguagePack).PermPresets
}

func GetAccountPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).Accounts[name]
	if !ok {
		return getPhrasePlaceholder("account", name)
	}
	return res
}

func GetUserAgentPhrase(name string) (string, bool) {
	res, ok := currentLangPack.Load().(*LanguagePack).UserAgents[name]
	if !ok {
		return "", false
	}
	return res, true
}

func GetOSPhrase(name string) (string, bool) {
	res, ok := currentLangPack.Load().(*LanguagePack).OperatingSystems[name]
	if !ok {
		return "", false
	}
	return res, true
}

func GetHumanLangPhrase(name string) (string, bool) {
	res, ok := currentLangPack.Load().(*LanguagePack).HumanLanguages[name]
	if !ok {
		return getPhrasePlaceholder("humanlang", name), false
	}
	return res, true
}

// TODO: Does comma ok work with multi-dimensional maps?
func GetErrorPhrase(category string, name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).Errors[category][name]
	if !ok {
		return getPhrasePlaceholder("error", name)
	}
	return res
}

func GetTitlePhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).PageTitles[name]
	if !ok {
		return getPhrasePlaceholder("title", name)
	}
	return res
}

func GetTmplPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).TmplPhrases[name]
	if !ok {
		return getPhrasePlaceholder("tmpl", name)
	}
	return res
}

func GetCSSPhrases() map[string]string {
	return currentLangPack.Load().(*LanguagePack).CSSPhrases
}

func getPhrasePlaceholder(prefix string, suffix string) string {
	return "{lang." + prefix + "[" + suffix + "]}"
}

// ? - Use runtime reflection for updating phrases?
// TODO: Implement these
func AddPhrase() {

}
func UpdatePhrase() {

}
func DeletePhrase() {

}

// TODO: Use atomics to store the pointer of the current active langpack?
// nolint
func ChangeLanguagePack(name string) (exists bool) {
	pack, ok := langPacks.Load(name)
	if !ok {
		return false
	}
	currentLangPack.Store(pack)
	return true
}

func CurrentLanguagePackName() (name string) {
	return currentLangPack.Load().(*LanguagePack).Name
}

func GetLanguagePackByName(name string) (pack *LanguagePack, ok bool) {
	packInt, ok := langPacks.Load(name)
	if !ok {
		return nil, false
	}
	return packInt.(*LanguagePack), true
}

// Template Transpiler Stuff

func RegisterTmplPhraseNames(phraseNames []string) (tmplID int) {
	langTmplIndicesToNames = append(langTmplIndicesToNames, phraseNames)
	return len(langTmplIndicesToNames) - 1
}

func GetTmplPhrasesBytes(tmplID int) [][]byte {
	return currentLangPack.Load().(*LanguagePack).TmplIndicesToPhrases[tmplID]
}
