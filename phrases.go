/*
*
* Gosora Phrase System
* Copyright Azareal 2017 - 2018
*
 */
package main

import (
	"sync"
	"sync/atomic"
)

// TO-DO: Let the admin edit phrases from inside the Control Panel? How should we persist these? Should we create a copy of the langpack or edit the primaries? Use the changeLangpack mutex for this?
var changeLangpackMutex sync.Mutex
var currentLanguage = "english"
var currentLangPack atomic.Value

type LevelPhrases struct {
	Level    string
	LevelMax string

	// Override the phrase for individual levels, if the phrases exist
	Levels []string // index = level
}

type LanguagePack struct {
	Name              string
	Phrases           map[string]string // Should we use a sync map or a struct for these? It would be nice, if we could keep all the phrases consistent.
	LevelPhrases      LevelPhrases
	GlobalPermPhrases map[string]string
	LocalPermPhrases  map[string]string
	SettingLabels     map[string]string
}

// TO-DO: Move the english language pack into it's own file and just keep the common logic here
var langpacks = map[string]*LanguagePack{
	"english": &LanguagePack{
		Name: "english",

		// We'll be implementing the level phrases in the software proper very very soon!
		LevelPhrases: LevelPhrases{
			Level:    "Level {0}",
			LevelMax: "", // Add a max level setting?
		},

		GlobalPermPhrases: map[string]string{
			"BanUsers":              "Can ban users",
			"ActivateUsers":         "Can activate users",
			"EditUser":              "Can edit users",
			"EditUserEmail":         "Can change a user's email",
			"EditUserPassword":      "Can change a user's password",
			"EditUserGroup":         "Can change a user's group",
			"EditUserGroupSuperMod": "Can edit super-mods",
			"EditUserGroupAdmin":    "Can edit admins",
			"EditGroup":             "Can edit groups",
			"EditGroupLocalPerms":   "Can edit a group's minor perms",
			"EditGroupGlobalPerms":  "Can edit a group's global perms",
			"EditGroupSuperMod":     "Can edit super-mod groups",
			"EditGroupAdmin":        "Can edit admin groups",
			"ManageForums":          "Can manage forums",
			"EditSettings":          "Can edit settings",
			"ManageThemes":          "Can manage themes",
			"ManagePlugins":         "Can manage plugins",
			"ViewAdminLogs":         "Can view the administrator action logs",
			"ViewIPs":               "Can view IP addresses",
		},

		LocalPermPhrases: map[string]string{
			"ViewTopic":   "Can view topics",
			"LikeItem":    "Can like items",
			"CreateTopic": "Can create topics",
			"EditTopic":   "Can edit topics",
			"DeleteTopic": "Can delete topics",
			"CreateReply": "Can create replies",
			"EditReply":   "Can edit replies",
			"DeleteReply": "Can delete replies",
			"PinTopic":    "Can pin topics",
			"CloseTopic":  "Can lock topics",
		},

		SettingLabels: map[string]string{
			"activation_type": "Activate All,Email Activation,Admin Approval",
		},
	},
}

func init() {
	currentLangPack.Store(langpacks[currentLanguage])
}

// We might not need to use a mutex for this, we shouldn't need to change the phrases after start-up, and when we do we could overwrite the entire map
func GetPhrase(name string) (string, bool) {
	res, ok := currentLangPack.Load().(*LanguagePack).Phrases[name]
	return res, ok
}

func GetGlobalPermPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).GlobalPermPhrases[name]
	if !ok {
		return "{name}"
	}
	return res
}

func GetLocalPermPhrase(name string) string {
	res, ok := currentLangPack.Load().(*LanguagePack).LocalPermPhrases[name]
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

// TO-DO: Use atomics to store the pointer of the current active langpack?
func ChangeLanguagePack(name string) (exists bool) {
	changeLangpackMutex.Lock()
	pack, ok := langpacks[name]
	if !ok {
		changeLangpackMutex.Unlock()
		return false
	}
	currentLangPack.Store(pack)
	changeLangpackMutex.Unlock()
	return true
}
