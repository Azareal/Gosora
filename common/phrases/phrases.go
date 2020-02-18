/*
*
* Gosora Phrase System
* Copyright Azareal 2017 - 2020
*
 */
package phrases

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	Name    string
	IsoCode string
	//LastUpdated string

	// Should we use a sync map or a struct for these? It would be nice, if we could keep all the phrases consistent.
	Levels              LevelPhrases
	Perms               map[string]string
	SettingPhrases      map[string]string
	PermPresets         map[string]string
	Accounts            map[string]string // TODO: Apply these phrases in the software proper
	UserAgents          map[string]string
	OperatingSystems    map[string]string
	HumanLanguages      map[string]string
	Errors              map[string]string // Temp stand-in
	ErrorsBytes         map[string][]byte
	NoticePhrases       map[string]string
	PageTitles          map[string]string
	TmplPhrases         map[string]string
	TmplPhrasesPrefixes map[string]map[string]string // [prefix][name]phrase

	TmplIndicesToPhrases [][][]byte // [tmplID][index]phrase
}

// TODO: Add the ability to edit language JSON files from the Control Panel and automatically scan the files for changes
var langPacks sync.Map                // nolint it is used
var langTmplIndicesToNames [][]string // [tmplID][index]phraseName

func InitPhrases(lang string) error {
	log.Print("Loading the language packs")
	err := filepath.Walk("./langs", func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}

		ext := filepath.Ext("/langs/" + path)
		if ext != ".json" {
			log.Printf("Found a '%s' in /langs/", ext)
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var langPack LanguagePack
		err = json.Unmarshal(data, &langPack)
		if err != nil {
			return err
		}

		langPack.ErrorsBytes = make(map[string][]byte)
		for name, phrase := range langPack.Errors {
			langPack.ErrorsBytes[name] = []byte(phrase)
		}

		// [prefix][name]phrase
		langPack.TmplPhrasesPrefixes = make(map[string]map[string]string)
		conMap := make(map[string]string) // Cache phrase strings so we can de-dupe items to reduce memory use. There appear to be some minor improvements with this, although we would need a more thorough check to be sure.
		for name, phrase := range langPack.TmplPhrases {
			_, ok := conMap[phrase]
			if !ok {
				conMap[phrase] = phrase
			}
			cItem := conMap[phrase]
			prefix := strings.Split(name, ".")[0]
			_, ok = langPack.TmplPhrasesPrefixes[prefix]
			if !ok {
				langPack.TmplPhrasesPrefixes[prefix] = make(map[string]string)
			}
			langPack.TmplPhrasesPrefixes[prefix][name] = cItem
		}

		// [prefix][name]phrase
		/*langPack.TmplPhrasesPrefixes = make(map[string]map[string]string)
		for name, phrase := range langPack.TmplPhrases {
			prefix := strings.Split(name, ".")[0]
			_, ok := langPack.TmplPhrasesPrefixes[prefix]
			if !ok {
				langPack.TmplPhrasesPrefixes[prefix] = make(map[string]string)
			}
			langPack.TmplPhrasesPrefixes[prefix][name] = phrase
		}*/

		langPack.TmplIndicesToPhrases = make([][][]byte, len(langTmplIndicesToNames))
		for tmplID, phraseNames := range langTmplIndicesToNames {
			phraseSet := make([][]byte, len(phraseNames))
			for index, phraseName := range phraseNames {
				phrase, ok := langPack.TmplPhrases[phraseName]
				if !ok {
					log.Printf("langPack.TmplPhrases: %+v\n", langPack.TmplPhrases)
					panic("Couldn't find template phrase '" + phraseName + "'")
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

	langPack, ok := langPacks.Load(lang)
	if !ok {
		return errors.New("Couldn't find the " + lang + " language pack")
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

func GetLangPack() *LanguagePack {
	return currentLangPack.Load().(*LanguagePack)
}

func GetLevelPhrase(level int) string {
	levelPhrases := currentLangPack.Load().(*LanguagePack).Levels
	if len(levelPhrases.Levels) > 0 && level < len(levelPhrases.Levels) {
		return strings.Replace(levelPhrases.Levels[level], "{0}", strconv.Itoa(level), -1)
	}
	return strings.Replace(levelPhrases.Level, "{0}", strconv.Itoa(level), -1)
}

func GetPermPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).Perms[name]
	if !ok {
		return getPlaceholder("perms", name)
	}
	return res
}

func GetSettingPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).SettingPhrases[name]
	if !ok {
		return getPlaceholder("settings", name)
	}
	return res
}

func GetAllSettingPhrases() map[string]string {
	return currentLangPack.Load().(*LanguagePack).SettingPhrases
}

func GetAllPermPresets() map[string]string {
	return currentLangPack.Load().(*LanguagePack).PermPresets
}

func GetAccountPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).Accounts[name]
	if !ok {
		return getPlaceholder("account", name)
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
		return getPlaceholder("humanlang", name), false
	}
	return res, true
}

// TODO: Does comma ok work with multi-dimensional maps?
func GetErrorPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).Errors[name]
	if !ok {
		return getPlaceholder("error", name)
	}
	return res
}
func GetErrorPhraseBytes(name string) []byte {
	res, ok := currentLangPack.Load().(*LanguagePack).ErrorsBytes[name]
	if !ok {
		return getPlaceholderBytes("error", name)
	}
	return res
}

func GetNoticePhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).NoticePhrases[name]
	if !ok {
		return getPlaceholder("notices", name)
	}
	return res
}

func GetTitlePhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).PageTitles[name]
	if !ok {
		return getPlaceholder("title", name)
	}
	return res
}

func GetTitlePhrasef(name string, params ...interface{}) string {
	res, ok := currentLangPack.Load().(*LanguagePack).PageTitles[name]
	if !ok {
		return getPlaceholder("title", name)
	}
	return fmt.Sprintf(res, params...)
}

func GetTmplPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).TmplPhrases[name]
	if !ok {
		return getPlaceholder("tmpl", name)
	}
	return res
}

func GetTmplPhrasef(name string, params ...interface{}) string {
	res, ok := currentLangPack.Load().(*LanguagePack).TmplPhrases[name]
	if !ok {
		return getPlaceholder("tmpl", name)
	}
	return fmt.Sprintf(res, params...)
}

func GetTmplPhrases() map[string]string {
	return currentLangPack.Load().(*LanguagePack).TmplPhrases
}

func GetTmplPhrasesByPrefix(prefix string) (phrases map[string]string, ok bool) {
	res, ok := currentLangPack.Load().(*LanguagePack).TmplPhrasesPrefixes[prefix]
	return res, ok
}

func getPlaceholder(prefix, suffix string) string {
	return "{lang." + prefix + "[" + suffix + "]}"
}
func getPlaceholderBytes(prefix, suffix string) []byte {
	return []byte("{lang." + prefix + "[" + suffix + "]}")
}

// Please don't mutate *LanguagePack
func GetCurrentLangPack() *LanguagePack {
	return currentLangPack.Load().(*LanguagePack)
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
