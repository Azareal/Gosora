package main

import "strconv"
import "testing"

// TODO: Generate a test database to work with rather than a live one
// TODO: We might need to refactor TestUserStore soon, as it's getting fairly complex
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
	userStoreTest(t)
	users = NewSQLUserStore()
	userStoreTest(t)
}
func userStoreTest(t *testing.T) {
	var user *User
	var err error
	var length int

	ucache, hasCache := users.(UserCache)
	if hasCache && ucache.GetLength() != 0 {
		t.Error("Initial ucache length isn't zero")
	}

	_, err = users.Get(-1)
	if err == nil {
		t.Error("UID #-1 shouldn't exist")
	} else if err != ErrNoRows {
		t.Fatal(err)
	}

	if hasCache && ucache.GetLength() != 0 {
		t.Error("There shouldn't be anything in the user cache")
	}

	_, err = users.Get(0)
	if err == nil {
		t.Error("UID #0 shouldn't exist")
	} else if err != ErrNoRows {
		t.Fatal(err)
	}

	if hasCache && ucache.GetLength() != 0 {
		t.Error("There shouldn't be anything in the user cache")
	}

	user, err = users.Get(1)
	if err == ErrNoRows {
		t.Error("Couldn't find UID #1")
	} else if err != nil {
		t.Fatal(err)
	}

	if user.ID != 1 {
		t.Error("user.ID does not match the requested UID. Got '" + strconv.Itoa(user.ID) + "' instead.")
	}

	if hasCache {
		length = ucache.GetLength()
		if length != 1 {
			t.Error("User cache length should be 1, not " + strconv.Itoa(length))
		}

		user, err = ucache.CacheGet(1)
		if err == ErrNoRows {
			t.Error("Couldn't find UID #1 in the cache")
		} else if err != nil {
			t.Fatal(err)
		}

		if user.ID != 1 {
			t.Error("user.ID does not match the requested UID. Got '" + strconv.Itoa(user.ID) + "' instead.")
		}

		ucache.Flush()
		length = ucache.GetLength()
		if length != 0 {
			t.Error("User cache length should be 0, not " + strconv.Itoa(length))
		}
	}

	// TODO: Lock onto the specific error type. Is this even possible without sacrificing the detailed information in the error message?
	var userList map[int]*User
	userList, _ = users.BulkGetMap([]int{-1})
	if len(userList) > 0 {
		t.Error("There shouldn't be any results for UID #-1")
	}

	if hasCache {
		length = ucache.GetLength()
		if length != 0 {
			t.Error("User cache length should be 0, not " + strconv.Itoa(length))
		}
	}

	userList, _ = users.BulkGetMap([]int{0})
	if len(userList) > 0 {
		t.Error("There shouldn't be any results for UID #0")
	}

	if hasCache {
		length = ucache.GetLength()
		if length != 0 {
			t.Error("User cache length should be 0, not " + strconv.Itoa(length))
		}
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
		length = ucache.GetLength()
		if length != 1 {
			t.Error("User cache length should be 1, not " + strconv.Itoa(length))
		}

		user, err = ucache.CacheGet(1)
		if err == ErrNoRows {
			t.Error("Couldn't find UID #1 in the cache")
		} else if err != nil {
			t.Fatal(err)
		}

		if user.ID != 1 {
			t.Error("user.ID does not match the requested UID. Got '" + strconv.Itoa(user.ID) + "' instead.")
		}

		ucache.Flush()
	}

	ok = users.Exists(-1)
	if ok {
		t.Error("UID #-1 shouldn't exist")
	}

	ok = users.Exists(0)
	if ok {
		t.Error("UID #0 shouldn't exist")
	}

	ok = users.Exists(1)
	if !ok {
		t.Error("UID #1 should exist")
	}

	if hasCache {
		length = ucache.GetLength()
		if length != 0 {
			t.Error("User cache length should be 0, not " + strconv.Itoa(length))
		}
	}

	count := users.GetGlobalCount()
	if count <= 0 {
		t.Error("The number of users should be bigger than zero")
		t.Error("count", count)
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
	if err == ErrNoRows {
		t.Error("Couldn't find TID #1")
	} else if err != nil {
		t.Fatal(err)
	}

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

	count := topics.GetGlobalCount()
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
	if err == ErrNoRows {
		t.Error("Couldn't find FID #1")
	} else if err != nil {
		t.Fatal(err)
	}

	if forum.ID != 1 {
		t.Error("forum.ID doesn't not match the requested FID. Got '" + strconv.Itoa(forum.ID) + "' instead.'")
	}
	if forum.Name != "Reports" {
		t.Error("FID #0 is named '" + forum.Name + "' and not 'Reports'")
	}

	forum, err = fstore.Get(2)
	if err == ErrNoRows {
		t.Error("Couldn't find FID #2")
	} else if err != nil {
		t.Fatal(err)
	}

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

	group, err = gstore.Get(0)
	if err == ErrNoRows {
		t.Error("Couldn't find GID #0")
	} else if err != nil {
		t.Fatal(err)
	}

	if group.ID != 0 {
		t.Error("group.ID doesn't not match the requested GID. Got '" + strconv.Itoa(group.ID) + "' instead.")
	}
	if group.Name != "Unknown" {
		t.Error("GID #0 is named '" + group.Name + "' and not 'Unknown'")
	}

	// ? - What if they delete this group? x.x
	// ? - Maybe, pick a random group ID? That would take an extra query, and I'm not sure if I want to be rewriting custom test queries. Possibly, a Random() method on the GroupStore? Seems useless for normal use, it might have some merit for the TopicStore though
	group, err = gstore.Get(1)
	if err == ErrNoRows {
		t.Error("Couldn't find GID #1")
	} else if err != nil {
		t.Fatal(err)
	}

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
