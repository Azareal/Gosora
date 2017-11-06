package main

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"runtime/debug"
	"strconv"
	"testing"
	"time"
)

func recordMustExist(t *testing.T, err error, errmsg string, args ...interface{}) {
	if err == ErrNoRows {
		t.Errorf(errmsg, args...)
	} else if err != nil {
		t.Fatal(err)
	}
}

func recordMustNotExist(t *testing.T, err error, errmsg string, args ...interface{}) {
	if err == nil {
		t.Errorf(errmsg, args...)
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

	var err error
	users, err = NewMemoryUserStore(config.UserCacheCapacity)
	expectNilErr(t, err)
	users.(UserCache).Flush()
	userStoreTest(t, 2)
	users, err = NewSQLUserStore()
	expectNilErr(t, err)
	userStoreTest(t, 3)
}
func userStoreTest(t *testing.T, newUserID int) {
	ucache, hasCache := users.(UserCache)
	// Go doesn't have short-circuiting, so this'll allow us to do one liner tests
	if !hasCache {
		ucache = &NullUserStore{}
	}
	expect(t, (!hasCache || ucache.Length() == 0), fmt.Sprintf("The initial ucache length should be zero, not %d", ucache.Length()))

	_, err := users.Get(-1)
	recordMustNotExist(t, err, "UID #-1 shouldn't exist")
	expect(t, !hasCache || ucache.Length() == 0, fmt.Sprintf("We found %d items in the user cache and it's supposed to be empty", ucache.Length()))

	_, err = users.Get(0)
	recordMustNotExist(t, err, "UID #0 shouldn't exist")
	expect(t, !hasCache || ucache.Length() == 0, fmt.Sprintf("We found %d items in the user cache and it's supposed to be empty", ucache.Length()))

	user, err := users.Get(1)
	recordMustExist(t, err, "Couldn't find UID #1")

	expect(t, user.ID == 1, fmt.Sprintf("user.ID should be 1. Got '%d' instead.", user.ID))
	expect(t, user.Name == "Admin", fmt.Sprintf("user.Name should be 'Admin', not '%s'", user.Name))
	expect(t, user.Group == 1, "Admin should be in group 1")
	expect(t, user.IsSuperAdmin, "Admin should be a super admin")
	expect(t, user.IsAdmin, "Admin should be an admin")
	expect(t, user.IsSuperMod, "Admin should be a super mod")
	expect(t, user.IsMod, "Admin should be a mod")
	expect(t, !user.IsBanned, "Admin should not be banned")

	_, err = users.Get(newUserID)
	recordMustNotExist(t, err, fmt.Sprintf("UID #%d shouldn't exist", newUserID))

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 1, "User cache length should be 1, not %d")

		_, err = ucache.CacheGet(-1)
		recordMustNotExist(t, err, "UID #-1 shouldn't exist, even in the cache")
		_, err = ucache.CacheGet(0)
		recordMustNotExist(t, err, "UID #0 shouldn't exist, even in the cache")
		user, err = ucache.CacheGet(1)
		recordMustExist(t, err, "Couldn't find UID #1 in the cache")

		if user.ID != 1 {
			t.Error("user.ID does not match the requested UID. Got '" + strconv.Itoa(user.ID) + "' instead.")
		}
		if user.Name != "Admin" {
			t.Error("user.Name should be 'Admin', not '" + user.Name + "'")
		}

		_, err = ucache.CacheGet(newUserID)
		recordMustNotExist(t, err, "UID #%d shouldn't exist, even in the cache", newUserID)

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
			t.Errorf("user.ID does not match the requested UID. Got '%d' instead.", user.ID)
		}
		ucache.Flush()
	}

	expect(t, !users.Exists(-1), "UID #-1 shouldn't exist")
	expect(t, !users.Exists(0), "UID #0 shouldn't exist")
	expect(t, users.Exists(1), "UID #1 should exist")
	expect(t, !users.Exists(newUserID), fmt.Sprintf("UID #%d shouldn't exist", newUserID))

	expect(t, !hasCache || ucache.Length() == 0, fmt.Sprintf("User cache length should be 0, not %d", ucache.Length()))
	expectIntToBeX(t, users.GlobalCount(), 1, "The number of users should be one, not %d")

	var awaitingActivation = 5
	uid, err := users.Create("Sam", "ReallyBadPassword", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expect(t, uid == newUserID, fmt.Sprintf("The UID of the new user should be %d", newUserID))
	expect(t, users.Exists(newUserID), fmt.Sprintf("UID #%d should exist", newUserID))

	user, err = users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	if user.ID != newUserID {
		t.Errorf("The UID of the user record should be %d", newUserID)
	}

	expect(t, user.Name == "Sam", "The user should be named Sam")
	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")
	expectIntToBeX(t, user.Group, 5, "Sam should be in group 5")

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 1, "User cache length should be 1, not %d")
		user, err = ucache.CacheGet(newUserID)
		recordMustExist(t, err, "Couldn't find UID #%d in the cache", newUserID)
		expect(t, user.ID == newUserID, fmt.Sprintf("user.ID does not match the requested UID. Got '%d' instead.", user.ID))
	}

	err = user.Activate()
	expectNilErr(t, err)
	expectIntToBeX(t, user.Group, 5, "Sam should still be in group 5 in this copy")

	// ? - What if we change the caching mechanism so it isn't hard purged and reloaded? We'll deal with that when we come to it, but for now, this is a sign of a cache bug
	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.CacheGet(newUserID)
		recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", newUserID)
	}

	user, err = users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)

	expect(t, user.ID == newUserID, fmt.Sprintf("The UID of the user record should be %d, not %d", newUserID, user.ID))
	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	expect(t, user.Group == config.DefaultGroup, fmt.Sprintf("Sam should be in group %d, not %d", config.DefaultGroup, user.Group))

	// Permanent ban
	duration, _ := time.ParseDuration("0")

	// TODO: Attempt a double ban, double activation, and double unban
	err = user.Ban(duration, 1)
	expectNilErr(t, err)
	expect(t, user.Group == config.DefaultGroup, fmt.Sprintf("Sam should be in group %d, not %d", config.DefaultGroup, user.Group))

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.CacheGet(2)
		recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", newUserID)
	}

	user, err = users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	if user.ID != newUserID {
		t.Errorf("The UID of the user record should be %d", newUserID)
	}

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, user.IsBanned, "Sam should be banned")

	expectIntToBeX(t, user.Group, banGroup, "Sam should be in group %d")

	// TODO: Do tests against the scheduled updates table and the task system to make sure the ban exists there and gets revoked when it should

	err = user.Unban()
	expectNilErr(t, err)
	expectIntToBeX(t, user.Group, banGroup, "Sam should still be in the ban group in this copy")

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.CacheGet(newUserID)
		recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", newUserID)
	}

	user, err = users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	expectIntToBeX(t, user.Group, config.DefaultGroup, "Sam should be back in group %d")

	var reportsForumID = 1
	var generalForumID = 2
	dummyResponseRecorder := httptest.NewRecorder()
	bytesBuffer := bytes.NewBuffer([]byte(""))
	dummyRequest1 := httptest.NewRequest("", "/forum/1", bytesBuffer)
	dummyRequest2 := httptest.NewRequest("", "/forum/2", bytesBuffer)

	err = user.ChangeGroup(1)
	expectNilErr(t, err)
	expect(t, user.Group == config.DefaultGroup, "Someone's mutated this pointer elsewhere")

	user, err = users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	var user2 *User = getDummyUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, user.IsAdmin, "Sam should be an admin")
	expect(t, user.IsSuperMod, "Sam should be a super mod")
	expect(t, user.IsMod, "Sam should be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	_, ferr := forumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user.Perms.ViewTopic, "Admins should be able to access the reports forum")
	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")

	err = user.ChangeGroup(2)
	expectNilErr(t, err)
	expect(t, user.Group == 1, "Someone's mutated this pointer elsewhere")

	user, err = users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	user2 = getDummyUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, user.IsSuperMod, "Sam should be a super mod")
	expect(t, user.IsMod, "Sam should be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user.Perms.ViewTopic, "Mods should be able to access the reports forum")
	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")

	err = user.ChangeGroup(3)
	expectNilErr(t, err)
	expect(t, user.Group == 2, "Someone's mutated this pointer elsewhere")

	user, err = users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	user2 = getDummyUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, !user.Perms.ViewTopic, "Members shouldn't be able to access the reports forum")
	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")
	expect(t, user.Perms.ViewTopic != user2.Perms.ViewTopic, "user.Perms.ViewTopic and user2.Perms.ViewTopic should never match")

	err = user.ChangeGroup(4)
	expectNilErr(t, err)
	expect(t, user.Group == 3, "Someone's mutated this pointer elsewhere")

	user, err = users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	user2 = getDummyUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, user.IsBanned, "Sam should be banned")

	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, !user.Perms.ViewTopic, "Members shouldn't be able to access the reports forum")
	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")

	err = user.ChangeGroup(5)
	expectNilErr(t, err)
	expect(t, user.Group == 4, "Someone's mutated this pointer elsewhere")

	user, err = users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	user2 = getDummyUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, !user.Perms.ViewTopic, "Members shouldn't be able to access the reports forum")
	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")

	err = user.ChangeGroup(6)
	expectNilErr(t, err)
	expect(t, user.Group == 5, "Someone's mutated this pointer elsewhere")

	user, err = users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	user2 = getDummyUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, !user.Perms.ViewTopic, "Members shouldn't be able to access the reports forum")
	_, ferr = forumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")

	err = user.ChangeGroup(config.DefaultGroup)
	expectNilErr(t, err)
	expect(t, user.Group == 6, "Someone's mutated this pointer elsewhere")

	err = user.Delete()
	expectNilErr(t, err)
	expect(t, !users.Exists(newUserID), fmt.Sprintf("UID #%d should no longer exist", newUserID))

	if hasCache {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.CacheGet(newUserID)
		recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", newUserID)
	}

	_, err = users.Get(newUserID)
	recordMustNotExist(t, err, "UID #%d shouldn't exist", newUserID)

	// TODO: Add tests for the Cache* methods
}

// TODO: Add an error message to this?
func expectNilErr(t *testing.T, item error) {
	if item != nil {
		debug.PrintStack()
		t.Fatal(item)
	}
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
		t.Fatal(errmsg)
	}
}

func TestPermsMiddleware(t *testing.T) {
	if !gloinited {
		err := gloinit()
		if err != nil {
			t.Fatal(err)
		}
	}
	if !pluginsInited {
		initPlugins()
	}

	dummyResponseRecorder := httptest.NewRecorder()
	bytesBuffer := bytes.NewBuffer([]byte(""))
	dummyRequest := httptest.NewRequest("", "/forum/1", bytesBuffer)
	user := getDummyUser()

	ferr := SuperModOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Blank users shouldn't be supermods")

	user.IsSuperMod = false
	ferr = SuperModOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Non-supermods shouldn't be allowed through supermod gates")

	user.IsSuperMod = true
	ferr = SuperModOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr == nil, "Supermods should be allowed through supermod gates")

	// TODO: Loop over the Control Panel routes and make sure only supermods can get in

	user = getDummyUser()

	ferr = MemberOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Blank users shouldn't be considered loggedin")

	user.Loggedin = false
	ferr = MemberOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Guests shouldn't be able to access member areas")

	user.Loggedin = true
	ferr = MemberOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr == nil, "Logged in users should be able to access member areas")

	// TODO: Loop over the /user/ routes and make sure only members can access the ones other than /user/username
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

	var err error
	topics, err = NewMemoryTopicStore(config.TopicCacheCapacity)
	expectNilErr(t, err)
	topicStoreTest(t)
	topics, err = NewSQLTopicStore()
	expectNilErr(t, err)
	topicStoreTest(t)
}
func topicStoreTest(t *testing.T) {
	var topic *Topic
	var err error

	_, err = topics.Get(-1)
	recordMustNotExist(t, err, "TID #-1 shouldn't exist")

	_, err = topics.Get(0)
	recordMustNotExist(t, err, "TID #0 shouldn't exist")

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

	// TODO: Test topic creation and retrieving that created topic plus reload and inspecting the cache
}

func TestForumStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !pluginsInited {
		initPlugins()
	}

	_, err := fstore.Get(-1)
	recordMustNotExist(t, err, "FID #-1 shouldn't exist")

	_, err = fstore.Get(0)
	recordMustNotExist(t, err, "FID #0 shouldn't exist")

	forum, err := fstore.Get(1)
	recordMustExist(t, err, "Couldn't find FID #1")

	if forum.ID != 1 {
		t.Error("forum.ID doesn't not match the requested FID. Got '" + strconv.Itoa(forum.ID) + "' instead.'")
	}
	// TODO: Check the preset and forum permissions
	expect(t, forum.Name == "Reports", fmt.Sprintf("FID #0 is named '%s' and not 'Reports'", forum.Name))
	expect(t, !forum.Active, fmt.Sprintf("The reports forum shouldn't be active"))
	var expectDesc = "All the reports go here"
	expect(t, forum.Desc == expectDesc, fmt.Sprintf("The forum description should be '%s' not '%s'", expectDesc, forum.Desc))

	forum, err = fstore.Get(2)
	recordMustExist(t, err, "Couldn't find FID #1")

	expect(t, forum.ID == 2, fmt.Sprintf("The FID should be 2 not %d", forum.ID))
	expect(t, forum.Name == "General", fmt.Sprintf("The name of the forum should be 'General' not '%s'", forum.Name))
	expect(t, forum.Active, fmt.Sprintf("The general forum should be active"))
	expectDesc = "A place for general discussions which don't fit elsewhere"
	expect(t, forum.Desc == expectDesc, fmt.Sprintf("The forum description should be '%s' not '%s'", expectDesc, forum.Desc))

	ok := fstore.Exists(-1)
	expect(t, !ok, "FID #-1 shouldn't exist")
	ok = fstore.Exists(0)
	expect(t, !ok, "FID #0 shouldn't exist")
	ok = fstore.Exists(1)
	expect(t, ok, "FID #1 should exist")

	// TODO: Test forum creation
	// TODO: Test forum deletion
	// TODO: Test forum update
}

// TODO: Implement this
func TestForumPermsStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !pluginsInited {
		initPlugins()
	}
}

// TODO: Test the group permissions
func TestGroupStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !pluginsInited {
		initPlugins()
	}

	_, err := gstore.Get(-1)
	recordMustNotExist(t, err, "GID #-1 shouldn't exist")

	// TODO: Refactor the group store to remove GID #0
	group, err := gstore.Get(0)
	recordMustExist(t, err, "Couldn't find GID #0")

	if group.ID != 0 {
		t.Errorf("group.ID doesn't not match the requested GID. Got '%d' instead.", group.ID)
	}
	expect(t, group.Name == "Unknown", fmt.Sprintf("GID #0 is named '%s' and not 'Unknown'", group.Name))

	group, err = gstore.Get(1)
	recordMustExist(t, err, "Couldn't find GID #1")

	if group.ID != 1 {
		t.Errorf("group.ID doesn't not match the requested GID. Got '%d' instead.'", group.ID)
	}

	ok := gstore.Exists(-1)
	expect(t, !ok, "GID #-1 shouldn't exist")

	// 0 aka Unknown, for system posts and other oddities
	ok = gstore.Exists(0)
	expect(t, ok, "GID #0 should exist")

	ok = gstore.Exists(1)
	expect(t, ok, "GID #1 should exist")

	var isAdmin = true
	var isMod = true
	var isBanned = false
	gid, err := gstore.Create("Testing", "Test", isAdmin, isMod, isBanned)
	expectNilErr(t, err)
	expect(t, gstore.Exists(gid), "The group we just made doesn't exist")

	group, err = gstore.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, group.IsAdmin, "This should be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	isAdmin = false
	isMod = true
	isBanned = true
	gid, err = gstore.Create("Testing 2", "Test", isAdmin, isMod, isBanned)
	expectNilErr(t, err)
	expect(t, gstore.Exists(gid), "The group we just made doesn't exist")

	group, err = gstore.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This should not be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	// TODO: Make sure this pointer doesn't change once we refactor the group store to stop updating the pointer
	err = group.ChangeRank(false, false, true)
	expectNilErr(t, err)

	group, err = gstore.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This shouldn't be an admin group")
	expect(t, !group.IsMod, "This shouldn't be a mod group")
	expect(t, group.IsBanned, "This should be a ban group")

	err = group.ChangeRank(true, true, true)
	expectNilErr(t, err)

	group, err = gstore.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, group.IsAdmin, "This should be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	err = group.ChangeRank(false, true, true)
	expectNilErr(t, err)

	group, err = gstore.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This shouldn't be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	// Make sure the data is static
	gstore.Reload(gid)

	group, err = gstore.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This shouldn't be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	// TODO: Test group deletion
	// TODO: Test group reload
	// TODO: Test group cache set
}

func TestReplyStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !pluginsInited {
		initPlugins()
	}

	_, err := rstore.Get(-1)
	recordMustNotExist(t, err, "RID #-1 shouldn't exist")

	_, err = rstore.Get(0)
	recordMustNotExist(t, err, "RID #0 shouldn't exist")

	reply, err := rstore.Get(1)
	expectNilErr(t, err)
	expect(t, reply.ID == 1, fmt.Sprintf("RID #1 has the wrong ID. It should be 1 not %d", reply.ID))
	expect(t, reply.ParentID == 1, fmt.Sprintf("The parent topic of RID #1 should be 1 not %d", reply.ParentID))
	expect(t, reply.CreatedBy == 1, fmt.Sprintf("The creator of RID #1 should be 1 not %d", reply.CreatedBy))
	expect(t, reply.Content == "A reply!", fmt.Sprintf("The contents of RID #1 should be 'A reply!' not %s", reply.Content))
	expect(t, reply.IPAddress == "::1", fmt.Sprintf("The IPAddress of RID#1 should be '::1' not %s", reply.IPAddress))

	_, err = rstore.Get(2)
	recordMustNotExist(t, err, "RID #2 shouldn't exist")

	// TODO: Test Create and Get
	//Create(tid int, content string, ipaddress string, fid int, uid int) (id int, err error)
	rid, err := rstore.Create(1, "Fofofo", "::1", 2, 1)
	expectNilErr(t, err)
	expect(t, rid == 2, fmt.Sprintf("The next reply ID should be 2 not %d", rid))

	reply, err = rstore.Get(2)
	expectNilErr(t, err)
	expect(t, reply.ID == 2, fmt.Sprintf("RID #2 has the wrong ID. It should be 2 not %d", reply.ID))
	expect(t, reply.ParentID == 1, fmt.Sprintf("The parent topic of RID #2 should be 1 not %d", reply.ParentID))
	expect(t, reply.CreatedBy == 1, fmt.Sprintf("The creator of RID #2 should be 1 not %d", reply.CreatedBy))
	expect(t, reply.Content == "Fofofo", fmt.Sprintf("The contents of RID #2 should be 'Fofofo' not %s", reply.Content))
	expect(t, reply.IPAddress == "::1", fmt.Sprintf("The IPAddress of RID #2 should be '::1' not %s", reply.IPAddress))
}

func TestProfileReplyStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !pluginsInited {
		initPlugins()
	}

	_, err := prstore.Get(-1)
	recordMustNotExist(t, err, "PRID #-1 shouldn't exist")

	_, err = prstore.Get(0)
	recordMustNotExist(t, err, "PRID #0 shouldn't exist")

	_, err = prstore.Get(1)
	recordMustNotExist(t, err, "PRID #0 shouldn't exist")

	var profileID = 1
	prid, err := prstore.Create(profileID, "Haha", 1, "::1")
	expect(t, err == nil, "Unable to create a profile reply")
	expect(t, prid == 1, "The first profile reply should have an ID of 1")

	profileReply, err := prstore.Get(1)
	expect(t, err == nil, "PRID #1 should exist")
	expect(t, profileReply.ID == 1, fmt.Sprintf("The profile reply should have an ID of 1 not %d", profileReply.ID))
	expect(t, profileReply.ParentID == 1, fmt.Sprintf("The parent ID of the profile reply should be 1 not %d", profileReply.ParentID))
	expect(t, profileReply.Content == "Haha", fmt.Sprintf("The profile reply's contents should be 'Haha' not '%s'", profileReply.Content))
	expect(t, profileReply.CreatedBy == 1, fmt.Sprintf("The profile reply's creator should be 1 not %d", profileReply.CreatedBy))
	expect(t, profileReply.IPAddress == "::1", fmt.Sprintf("The profile reply's IP Address should be '::1' not '%s'", profileReply.IPAddress))

	//Get(id int) (*Reply, error)
	//Create(profileID int, content string, createdBy int, ipaddress string) (id int, err error)

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
	msgList = addMEPair(msgList, "-你好-", "untitled")

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
	t.Log("Set realPassword to '" + realPassword + "'")
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
