package main

import "sync"

// I wish we had constant maps x.x
var phrase_mutex sync.RWMutex
var perm_phrase_mutex sync.RWMutex
var phrases map[string]string
var global_perm_phrases map[string]string = map[string]string{
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
	"ViewIPs": "Can view IP addresses",
}

var local_perm_phrases map[string]string = map[string]string{
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
}

// We might not need to use a mutex for this, we shouldn't need to change the phrases after start-up, and when we do we could overwrite the entire map
func GetPhrase(name string) (string,bool) {
	phrase_mutex.RLock()
	defer perm_phrase_mutex.RUnlock()
	res, ok := phrases[name]
	return res, ok
}

func GetPhraseUnsafe(name string) (string,bool) {
	res, ok := phrases[name]
	return res, ok
}

func GetGlobalPermPhrase(name string) string {
	perm_phrase_mutex.RLock()
	defer perm_phrase_mutex.RUnlock()
	res, ok := global_perm_phrases[name]
	if !ok {
		return "{name}"
	}
	return res
}

func GetLocalPermPhrase(name string) string {
	perm_phrase_mutex.RLock()
	defer perm_phrase_mutex.RUnlock()
	res, ok := local_perm_phrases[name]
	if !ok {
		return "{name}"
	}
	return res
}

func AddPhrase() {
	
}

func DeletePhrase() {
	
}
