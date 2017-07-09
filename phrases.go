package main

import "sync"

// I wish we had constant maps x.x
var phrase_mutex sync.RWMutex
var perm_phrase_mutex sync.RWMutex
var change_langpack_mutex sync.Mutex
var currentLanguage string = "english"
var currentLangPack *LanguagePack

type LevelPhrases struct
{
		Level string
		LevelMax string

		// Override the phrase for individual levels, if the phrases exist
		Levels []string // index = level
}

type LanguagePack struct
{
	Name string
	Phrases map[string]string // Should we use a sync map or a struct for these? It would be nice, if we could keep all the phrases consistent.
	LevelPhrases LevelPhrases
	GlobalPermPhrases map[string]string
	LocalPermPhrases map[string]string
}

// TO-DO: Move the english language pack into it's own file and just keep the common logic here
var langpacks map[string]*LanguagePack = map[string]*LanguagePack{
	"english": &LanguagePack{
		Name: "english",

		// We'll be implementing the level phrases in the software proper very very soon!
		LevelPhrases: LevelPhrases{
			Level: "Level {0}",
			LevelMax: "", // Add a max level setting?
		},

		GlobalPermPhrases: map[string]string{
			"BanUsers": "Can ban users",
			"ActivateUsers": "Can activate users",
			"EditUser": "Can edit users",
			"EditUserEmail": "Can change a user's email",
			"EditUserPassword": "Can change a user's password",
			"EditUserGroup": "Can change a user's group",
			"EditUserGroupSuperMod": "Can edit super-mods",
			"EditUserGroupAdmin": "Can edit admins",
			"EditGroup": "Can edit groups",
			"EditGroupLocalPerms": "Can edit a group's minor perms",
			"EditGroupGlobalPerms": "Can edit a group's global perms",
			"EditGroupSuperMod": "Can edit super-mod groups",
			"EditGroupAdmin": "Can edit admin groups",
			"ManageForums": "Can manage forums",
			"EditSettings": "Can edit settings",
			"ManageThemes": "Can manage themes",
			"ManagePlugins": "Can manage plugins",
			"ViewAdminLogs": "Can view the administrator action logs",
			"ViewIPs": "Can view IP addresses",
		},

		LocalPermPhrases: map[string]string{
			"ViewTopic": "Can view topics",
			"LikeItem": "Can like items",
			"CreateTopic": "Can create topics",
			"EditTopic": "Can edit topics",
			"DeleteTopic": "Can delete topics",
			"CreateReply": "Can create replies",
			"EditReply": "Can edit replies",
			"DeleteReply": "Can delete replies",
			"PinTopic": "Can pin topics",
			"CloseTopic": "Can lock topics",
		},
	},
}

func init() {
	currentLangPack = langpacks[currentLanguage]
}

// We might not need to use a mutex for this, we shouldn't need to change the phrases after start-up, and when we do we could overwrite the entire map
func GetPhrase(name string) (string,bool) {
	phrase_mutex.RLock()
	defer perm_phrase_mutex.RUnlock()
	res, ok := currentLangPack.Phrases[name]
	return res, ok
}

func GetPhraseUnsafe(name string) (string,bool) {
	res, ok := currentLangPack.Phrases[name]
	return res, ok
}

func GetGlobalPermPhrase(name string) string {
	perm_phrase_mutex.RLock()
	defer perm_phrase_mutex.RUnlock()
	res, ok := currentLangPack.GlobalPermPhrases[name]
	if !ok {
		return "{name}"
	}
	return res
}

func GetLocalPermPhrase(name string) string {
	perm_phrase_mutex.RLock()
	defer perm_phrase_mutex.RUnlock()
	res, ok := currentLangPack.LocalPermPhrases[name]
	if !ok {
		return "{name}"
	}
	return res
}

func AddPhrase() {

}

func DeletePhrase() {

}

func ChangeLanguagePack(name string) (exists bool) {
	change_langpack_mutex.Lock()
	pack, ok := langpacks[name]
	if !ok {
		change_langpack_mutex.Unlock()
		return false
	}
	currentLangPack = pack
	change_langpack_mutex.Unlock()
	return true
}
