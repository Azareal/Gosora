package main

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"testing"
	"time"
)

func recordMustExist(t *testing.T, err error, errmsg string) {
	if err == ErrNoRows {
		t.Error(errmsg)
	} else if err != nil {
		t.Fatal(err)
	}
}

func recordMustNotExist(t *testing.T, err error, errmsg string) {
	if err == nil {
		t.Error(errmsg)
	} else if err != ErrNoRows {
		t.Fatal(err)
	}
}

func TestUserStore(t *testing.T) {
	if !gloinited {
		err := gloinit()
		if err != nil {
			t.Fatal(err)
		}
	}
	if !pluginsInited {
		initPlugins()
	}

	users = NewMemoryUserStore(config.UserCacheCapacity)
	users.(UserCache).Flush()
	userStoreTest(t, 2)
	users = NewSQLUserStore()
	userStoreTest(t, 3)
}
func userStoreTest(t *testing.T, newUserID int) {
	var user *User
	var err error

	ucache, hasCache := users.(UserCache)
	if hasCache && ucache.Length() != 0 {
		t.Error("Initial ucache length isn't zero")
	}

	_, err = users.Get(-1)
	recordMustNotExist(t, err, "UID #-1 shouldn't exist")

	if hasCache && ucache.Length() != 0 {
		t.Error("There shouldn't be anything in the user cache")
	}

	_, err = users.Get(0)
	recordMustNotExist(t, err, "UID #0 shouldn't exist")

	if hasCache && ucache.Length() != 0 {
		t.Error("There shouldn't be anything in the user cache")
	}

	user, err = users.Get(1)
	recordMustExist(t, err, "Couldn't find UID #1")

	if user.ID != 1 {
		t.Error("user.ID does not match the requested UID. Got '" + strconv.Itoa(user.ID) + "' instead.")
	}
	if user.Name != "Admin" {
		t.Error("user.Name should be 'Admin', not '" + user.Name + "'")
	}
	if user.Group != 1 {
		t.Error("Admin should be in group 1")
	}

	user, err = users.Get(newUserID)
	recordMustNotExist(t, err, fmt.Sprintf("UID #%d shouldn't exist", newUserID))

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 1, "User cache length should be 1, not %d")

		user, err = ucache.CacheGet(-1)
		recordMustNotExist(t, err, "UID #-1 shouldn't exist, even in the cache")
		user, err = ucache.CacheGet(0)
		recordMustNotExist(t, err, "UID #0 shouldn't exist, even in the cache")
		user, err = ucache.CacheGet(1)
		recordMustExist(t, err, "Couldn't find UID #1 in the cache")

		if user.ID != 1 {
			t.Error("user.ID does not match the requested UID. Got '" + strconv.Itoa(user.ID) + "' instead.")
		}
		if user.Name != "Admin" {
			t.Error("user.Name should be 'Admin', not '" + user.Name + "'")
		}

		user, err = ucache.CacheGet(newUserID)
		recordMustNotExist(t, err, fmt.Sprintf("UID #%d shouldn't exist, even in the cache", newUserID))

		ucache.Flush()
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
	}

	// TODO: Lock onto the specific error type. Is this even possible without sacrificing the detailed information in the error message?
	var userList map[int]*User
	userList, _ = users.BulkGetMap([]int{-1})
	if len(userList) > 0 {
		t.Error("There shouldn't be any results for UID #-1")
	}

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
	}

	userList, _ = users.BulkGetMap([]int{0})
	if len(userList) > 0 {
		t.Error("There shouldn't be any results for UID #0")
	}

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
	}

	userList, _ = users.BulkGetMap([]int{1})
	if len(userList) == 0 {
		t.Error("The returned map is empty for UID #1")
	} else if len(userList) > 1 {
		t.Error("Too many results were returned for UID #1")
	}

	user, ok := userList[1]
	if !ok {
		t.Error("We couldn't find UID #1 in the returned map")
		t.Error("userList", userList)
	}

	if user.ID != 1 {
		t.Error("user.ID does not match the requested UID. Got '" + strconv.Itoa(user.ID) + "' instead.")
	}

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 1, "User cache length should be 1, not %d")
		user, err = ucache.CacheGet(1)
		recordMustExist(t, err, "Couldn't find UID #1 in the cache")

		if user.ID != 1 {
			t.Error("user.ID does not match the requested UID. Got '" + strconv.Itoa(user.ID) + "' instead.")
		}

		ucache.Flush()
	}

	expect(t, !users.Exists(-1), "UID #-1 shouldn't exist")
	expect(t, !users.Exists(0), "UID #0 shouldn't exist")
	expect(t, users.Exists(1), "UID #1 should exist")
	expect(t, !users.Exists(newUserID), fmt.Sprintf("UID #%d shouldn't exist", newUserID))

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
	}
	expectIntToBeX(t, users.GlobalCount(), 1, "The number of users should be one, not %d")

	var awaitingActivation = 5
	uid, err := users.Create("Sam", "ReallyBadPassword", "sam@localhost.loc", awaitingActivation, 0)
	if err != nil {
		t.Error(err)
	}
	if uid != newUserID {
		t.Errorf("The UID of the new user should be %d", newUserID)
	}
	if !users.Exists(newUserID) {
		t.Errorf("UID #%d should exist", newUserID)
	}

	user, err = users.Get(newUserID)
	recordMustExist(t, err, fmt.Sprintf("Couldn't find UID #%d", newUserID))
	if user.ID != newUserID {
		t.Errorf("The UID of the user record should be %d", newUserID)
	}
	if user.Name != "Sam" {
		t.Error("The user should be named Sam")
	}
	expectIntToBeX(t, user.Group, 5, "Sam should be in group 5")

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 1, "User cache length should be 1, not %d")
		user, err = ucache.CacheGet(newUserID)
		recordMustExist(t, err, fmt.Sprintf("Couldn't find UID #%d in the cache", newUserID))
		if user.ID != newUserID {
			t.Error("user.ID does not match the requested UID. Got '" + strconv.Itoa(user.ID) + "' instead.")
		}
	}

	err = user.Activate()
	if err != nil {
		t.Error(err)
	}
	expectIntToBeX(t, user.Group, 5, "Sam should still be in group 5 in this copy")

	// ? - What if we change the caching mechanism so it isn't hard purged and reloaded? We'll deal with that when we come to it, but for now, this is a sign of a cache bug
	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.CacheGet(newUserID)
		recordMustNotExist(t, err, fmt.Sprintf("UID #%d shouldn't be in the cache", newUserID))
	}

	user, err = users.Get(newUserID)
	recordMustExist(t, err, fmt.Sprintf("Couldn't find UID #%d", newUserID))
	if user.ID != newUserID {
		t.Errorf("The UID of the user record should be %d", newUserID)
	}
	expectIntToBeX(t, user.Group, config.DefaultGroup, "Sam should be in group "+strconv.Itoa(config.DefaultGroup))

	// Permanent ban
	duration, _ := time.ParseDuration("0")

	// TODO: Attempt a double ban, double activation, and double unban
	err = user.Ban(duration, 1)
	if err != nil {
		t.Error(err)
	}
	expectIntToBeX(t, user.Group, config.DefaultGroup, "Sam should still be in the default group in this copy")

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.CacheGet(2)
		recordMustNotExist(t, err, fmt.Sprintf("UID #%d shouldn't be in the cache", newUserID))
	}

	user, err = users.Get(newUserID)
	recordMustExist(t, err, fmt.Sprintf("Couldn't find UID #%d", newUserID))
	if user.ID != newUserID {
		t.Errorf("The UID of the user record should be %d", newUserID)
	}
	expectIntToBeX(t, user.Group, banGroup, "Sam should be in group "+strconv.Itoa(banGroup))

	// TODO: Do tests against the scheduled updates table and the task system to make sure the ban exists there and gets revoked when it should

	err = user.Unban()
	if err != nil {
		t.Error(err)
	}
	expectIntToBeX(t, user.Group, banGroup, "Sam should still be in the ban group in this copy")

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.CacheGet(newUserID)
		recordMustNotExist(t, err, fmt.Sprintf("UID #%d shouldn't be in the cache", newUserID))
	}

	user, err = users.Get(newUserID)
	recordMustExist(t, err, fmt.Sprintf("Couldn't find UID #%d", newUserID))
	if user.ID != newUserID {
		t.Errorf("The UID of the user record should be %d", newUserID)
	}
	expectIntToBeX(t, user.Group, config.DefaultGroup, "Sam should be back in group "+strconv.Itoa(config.DefaultGroup))

	err = user.Delete()
	if err != nil {
		t.Error(err)
	}
	expect(t, !users.Exists(newUserID), fmt.Sprintf("UID #%d should not longer exist", newUserID))

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.CacheGet(newUserID)
		recordMustNotExist(t, err, fmt.Sprintf("UID #%d shouldn't be in the cache", newUserID))
	}

	// TODO: Works for now but might cause a data race with the task system
	//ResetTables()
}

func expectIntToBeX(t *testing.T, item int, expect int, errmsg string) {
	if item != expect {
		debug.PrintStack()
		t.Fatalf(errmsg, item)
	}
}

func expect(t *testing.T, item bool, errmsg string) {
	if !item {
		debug.PrintStack()
		t.Fatalf(errmsg)
	}
}

func TestTopicStore(t *testing.T) {
	if !gloinited {
		err := gloinit()
		if err != nil {
			t.Fatal(err)
		}
	}
	if !pluginsInited {
		initPlugins()
	}

	topics = NewMemoryTopicStore(config.TopicCacheCapacity)
	topicStoreTest(t)
	topics = NewSQLTopicStore()
	topicStoreTest(t)
}
func topicStoreTest(t *testing.T) {
	var topic *Topic
	var err error

	_, err = topics.Get(-1)
	if err == nil {
		t.Error("TID #-1 shouldn't exist")
	} else if err != ErrNoRows {
		t.Fatal(err)
	}

	_, err = topics.Get(0)
	if err == nil {
		t.Error("TID #0 shouldn't exist")
	} else if err != ErrNoRows {
		t.Fatal(err)
	}

	topic, err = topics.Get(1)
	recordMustExist(t, err, "Couldn't find TID #1")

	if topic.ID != 1 {
		t.Error("topic.ID does not match the requested TID. Got '" + strconv.Itoa(topic.ID) + "' instead.")
	}

	// TODO: Add BulkGetMap() to the TopicStore

	ok := topics.Exists(-1)
	if ok {
		t.Error("TID #-1 shouldn't exist")
	}

	ok = topics.Exists(0)
	if ok {
		t.Error("TID #0 shouldn't exist")
	}

	ok = topics.Exists(1)
	if !ok {
		t.Error("TID #1 should exist")
	}

	count := topics.GlobalCount()
	if count <= 0 {
		t.Error("The number of topics should be bigger than zero")
		t.Error("count", count)
	}
}

func TestForumStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !pluginsInited {
		initPlugins()
	}

	var forum *Forum
	var err error

	_, err = fstore.Get(-1)
	if err == nil {
		t.Error("FID #-1 shouldn't exist")
	} else if err != ErrNoRows {
		t.Fatal(err)
	}

	forum, err = fstore.Get(0)
	if err == nil {
		t.Error("FID #0 shouldn't exist")
	} else if err != ErrNoRows {
		t.Fatal(err)
	}

	forum, err = fstore.Get(1)
	recordMustExist(t, err, "Couldn't find FID #1")

	if forum.ID != 1 {
		t.Error("forum.ID doesn't not match the requested FID. Got '" + strconv.Itoa(forum.ID) + "' instead.'")
	}
	if forum.Name != "Reports" {
		t.Error("FID #0 is named '" + forum.Name + "' and not 'Reports'")
	}

	forum, err = fstore.Get(2)
	recordMustExist(t, err, "Couldn't find FID #1")

	_ = forum

	ok := fstore.Exists(-1)
	if ok {
		t.Error("FID #-1 shouldn't exist")
	}

	ok = fstore.Exists(0)
	if ok {
		t.Error("FID #0 shouldn't exist")
	}

	ok = fstore.Exists(1)
	if !ok {
		t.Error("FID #1 should exist")
	}
}

func TestGroupStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !pluginsInited {
		initPlugins()
	}

	var group *Group
	var err error

	_, err = gstore.Get(-1)
	if err == nil {
		t.Error("GID #-1 shouldn't exist")
	} else if err != ErrNoRows {
		t.Fatal(err)
	}

	// TODO: Refactor the group store to remove GID #0
	group, err = gstore.Get(0)
	recordMustExist(t, err, "Couldn't find GID #0")

	if group.ID != 0 {
		t.Error("group.ID doesn't not match the requested GID. Got '" + strconv.Itoa(group.ID) + "' instead.")
	}
	if group.Name != "Unknown" {
		t.Error("GID #0 is named '" + group.Name + "' and not 'Unknown'")
	}

	group, err = gstore.Get(1)
	recordMustExist(t, err, "Couldn't find GID #1")

	if group.ID != 1 {
		t.Error("group.ID doesn't not match the requested GID. Got '" + strconv.Itoa(group.ID) + "' instead.'")
	}

	_ = group

	ok := gstore.Exists(-1)
	if ok {
		t.Error("GID #-1 shouldn't exist")
	}

	ok = gstore.Exists(0)
	if !ok {
		t.Error("GID #0 should exist")
	}

	ok = gstore.Exists(1)
	if !ok {
		t.Error("GID #1 should exist")
	}
}

func TestReplyStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !pluginsInited {
		initPlugins()
	}

	reply, err := rstore.Get(-1)
	if err == nil {
		t.Error("RID #-1 shouldn't exist")
	}

	reply, err = rstore.Get(0)
	if err == nil {
		t.Error("RID #0 shouldn't exist")
	}

	reply, err = rstore.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if reply.ID != 1 {
		t.Error("RID #1 has the wrong ID. It should be 1 not " + strconv.Itoa(reply.ID))
	}
	if reply.ParentID != 1 {
		t.Error("The parent topic of RID #1 should be 1 not " + strconv.Itoa(reply.ParentID))
	}
	if reply.CreatedBy != 1 {
		t.Error("The creator of RID #1 should be 1 not " + strconv.Itoa(reply.CreatedBy))
	}
}

func TestProfileReplyStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !pluginsInited {
		initPlugins()
	}

	_, err := prstore.Get(-1)
	if err == nil {
		t.Error("RID #-1 shouldn't exist")
	}

	_, err = prstore.Get(0)
	if err == nil {
		t.Error("RID #0 shouldn't exist")
	}
}

func TestSlugs(t *testing.T) {
	var res string
	var msgList []MEPair

	msgList = addMEPair(msgList, "Unknown", "unknown")
	msgList = addMEPair(msgList, "Unknown2", "unknown2")
	msgList = addMEPair(msgList, "Unknown ", "unknown")
	msgList = addMEPair(msgList, "Unknown 2", "unknown-2")
	msgList = addMEPair(msgList, "Unknown  2", "unknown-2")
	msgList = addMEPair(msgList, "Admin Alice", "admin-alice")
	msgList = addMEPair(msgList, "Admin_Alice", "adminalice")
	msgList = addMEPair(msgList, "Admin_Alice-", "adminalice")
	msgList = addMEPair(msgList, "-Admin_Alice-", "adminalice")
	msgList = addMEPair(msgList, "-Admin@Alice-", "adminalice")
	msgList = addMEPair(msgList, "-Admin😀Alice-", "adminalice")
	msgList = addMEPair(msgList, "u", "u")
	msgList = addMEPair(msgList, "", "untitled")
	msgList = addMEPair(msgList, " ", "untitled")
	msgList = addMEPair(msgList, "-", "untitled")
	msgList = addMEPair(msgList, "--", "untitled")
	msgList = addMEPair(msgList, "é", "é")
	msgList = addMEPair(msgList, "-é-", "é")

	for _, item := range msgList {
		t.Log("Testing string '" + item.Msg + "'")
		res = nameToSlug(item.Msg)
		if res != item.Expects {
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", item.Expects)
		}
	}
}

func TestAuth(t *testing.T) {
	// bcrypt likes doing stupid things, so this test will probably fail
	var realPassword string
	var hashedPassword string
	var password string
	var salt string
	var err error

	/* No extra salt tests, we might not need this extra salt, as bcrypt has it's own? */
	realPassword = "Madame Cassandra's Mystic Orb"
	t.Log("Set real_password to '" + realPassword + "'")
	t.Log("Hashing the real password")
	hashedPassword, err = BcryptGeneratePasswordNoSalt(realPassword)
	if err != nil {
		t.Error(err)
	}

	password = realPassword
	t.Log("Testing password '" + password + "'")
	t.Log("Testing salt '" + salt + "'")
	err = CheckPassword(hashedPassword, password, salt)
	if err == ErrMismatchedHashAndPassword {
		t.Error("The two don't match")
	} else if err == ErrPasswordTooLong {
		t.Error("CheckPassword thinks the password is too long")
	} else if err != nil {
		t.Error(err)
	}

	password = "hahaha"
	t.Log("Testing password '" + password + "'")
	t.Log("Testing salt '" + salt + "'")
	err = CheckPassword(hashedPassword, password, salt)
	if err == ErrPasswordTooLong {
		t.Error("CheckPassword thinks the password is too long")
	} else if err == nil {
		t.Error("The two shouldn't match!")
	}

	password = "Madame Cassandra's Mystic"
	t.Log("Testing password '" + password + "'")
	t.Log("Testing salt '" + salt + "'")
	err = CheckPassword(hashedPassword, password, salt)
	if err == ErrPasswordTooLong {
		t.Error("CheckPassword thinks the password is too long")
	} else if err == nil {
		t.Error("The two shouldn't match!")
	}
}
