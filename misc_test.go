package main

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"runtime/debug"
	"testing"
	"time"

	"./common"
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
	if !common.PluginsInited {
		common.InitPlugins()
	}

	var err error
	ucache := common.NewMemoryUserCache(common.Config.UserCacheCapacity)
	common.Users, err = common.NewDefaultUserStore(ucache)
	expectNilErr(t, err)
	ucache.Flush()
	userStoreTest(t, 2)
	common.Users, err = common.NewDefaultUserStore(nil)
	expectNilErr(t, err)
	userStoreTest(t, 3)
}
func userStoreTest(t *testing.T, newUserID int) {
	ucache := common.Users.GetCache()
	// Go doesn't have short-circuiting, so this'll allow us to do one liner tests
	isCacheLengthZero := func(ucache common.UserCache) bool {
		if ucache == nil {
			return true
		}
		return ucache.Length() == 0
	}
	cacheLength := func(ucache common.UserCache) int {
		if ucache == nil {
			return 0
		}
		return ucache.Length()
	}
	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("The initial ucache length should be zero, not %d", cacheLength(ucache)))

	_, err := common.Users.Get(-1)
	recordMustNotExist(t, err, "UID #-1 shouldn't exist")
	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("We found %d items in the user cache and it's supposed to be empty", cacheLength(ucache)))

	_, err = common.Users.Get(0)
	recordMustNotExist(t, err, "UID #0 shouldn't exist")
	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("We found %d items in the user cache and it's supposed to be empty", cacheLength(ucache)))

	user, err := common.Users.Get(1)
	recordMustExist(t, err, "Couldn't find UID #1")

	expect(t, user.ID == 1, fmt.Sprintf("user.ID should be 1. Got '%d' instead.", user.ID))
	expect(t, user.Name == "Admin", fmt.Sprintf("user.Name should be 'Admin', not '%s'", user.Name))
	expect(t, user.Group == 1, "Admin should be in group 1")
	expect(t, user.IsSuperAdmin, "Admin should be a super admin")
	expect(t, user.IsAdmin, "Admin should be an admin")
	expect(t, user.IsSuperMod, "Admin should be a super mod")
	expect(t, user.IsMod, "Admin should be a mod")
	expect(t, !user.IsBanned, "Admin should not be banned")

	_, err = common.Users.Get(newUserID)
	recordMustNotExist(t, err, fmt.Sprintf("UID #%d shouldn't exist", newUserID))

	if ucache != nil {
		expectIntToBeX(t, ucache.Length(), 1, "User cache length should be 1, not %d")

		_, err = ucache.Get(-1)
		recordMustNotExist(t, err, "UID #-1 shouldn't exist, even in the cache")
		_, err = ucache.Get(0)
		recordMustNotExist(t, err, "UID #0 shouldn't exist, even in the cache")
		user, err = ucache.Get(1)
		recordMustExist(t, err, "Couldn't find UID #1 in the cache")

		expect(t, user.ID == 1, fmt.Sprintf("user.ID does not match the requested UID. Got '%d' instead.", user.ID))
		expect(t, user.Name == "Admin", fmt.Sprintf("user.Name should be 'Admin', not '%s'", user.Name))

		_, err = ucache.Get(newUserID)
		recordMustNotExist(t, err, "UID #%d shouldn't exist, even in the cache", newUserID)

		ucache.Flush()
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
	}

	// TODO: Lock onto the specific error type. Is this even possible without sacrificing the detailed information in the error message?
	var userList map[int]*common.User
	userList, _ = common.Users.BulkGetMap([]int{-1})
	expect(t, len(userList) == 0, fmt.Sprintf("The userList length should be 0, not %d", len(userList)))
	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("User cache length should be 0, not %d", cacheLength(ucache)))

	userList, _ = common.Users.BulkGetMap([]int{0})
	expect(t, len(userList) == 0, fmt.Sprintf("The userList length should be 0, not %d", len(userList)))
	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("User cache length should be 0, not %d", cacheLength(ucache)))

	userList, _ = common.Users.BulkGetMap([]int{1})
	if len(userList) == 0 {
		t.Error("The returned map is empty for UID #1")
	} else if len(userList) > 1 {
		t.Error("Too many results were returned for UID #1")
	}

	user, ok := userList[1]
	if !ok {
		t.Error("We couldn't find UID #1 in the returned map")
		t.Error("userList", userList)
		return
	}
	expect(t, user.ID == 1, fmt.Sprintf("user.ID does not match the requested UID. Got '%d' instead.", user.ID))

	if ucache != nil {
		expectIntToBeX(t, ucache.Length(), 1, "User cache length should be 1, not %d")
		user, err = ucache.Get(1)
		recordMustExist(t, err, "Couldn't find UID #1 in the cache")

		expect(t, user.ID == 1, fmt.Sprintf("user.ID does not match the requested UID. Got '%d' instead.", user.ID))
		ucache.Flush()
	}

	expect(t, !common.Users.Exists(-1), "UID #-1 shouldn't exist")
	expect(t, !common.Users.Exists(0), "UID #0 shouldn't exist")
	expect(t, common.Users.Exists(1), "UID #1 should exist")
	expect(t, !common.Users.Exists(newUserID), fmt.Sprintf("UID #%d shouldn't exist", newUserID))

	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("User cache length should be 0, not %d", cacheLength(ucache)))
	expectIntToBeX(t, common.Users.GlobalCount(), 1, "The number of users should be one, not %d")

	var awaitingActivation = 5
	uid, err := common.Users.Create("Sam", "ReallyBadPassword", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expect(t, uid == newUserID, fmt.Sprintf("The UID of the new user should be %d", newUserID))
	expect(t, common.Users.Exists(newUserID), fmt.Sprintf("UID #%d should exist", newUserID))

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expect(t, user.ID == newUserID, fmt.Sprintf("The UID of the user record should be %d", newUserID))

	expect(t, user.Name == "Sam", "The user should be named Sam")
	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")
	expectIntToBeX(t, user.Group, 5, "Sam should be in group 5")

	if ucache != nil {
		expectIntToBeX(t, ucache.Length(), 1, "User cache length should be 1, not %d")
		user, err = ucache.Get(newUserID)
		recordMustExist(t, err, "Couldn't find UID #%d in the cache", newUserID)
		expect(t, user.ID == newUserID, fmt.Sprintf("user.ID does not match the requested UID. Got '%d' instead.", user.ID))
	}

	err = user.Activate()
	expectNilErr(t, err)
	expectIntToBeX(t, user.Group, 5, "Sam should still be in group 5 in this copy")

	// ? - What if we change the caching mechanism so it isn't hard purged and reloaded? We'll deal with that when we come to it, but for now, this is a sign of a cache bug
	if ucache != nil {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.Get(newUserID)
		recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", newUserID)
	}

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)

	expect(t, user.ID == newUserID, fmt.Sprintf("The UID of the user record should be %d, not %d", newUserID, user.ID))
	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	expect(t, user.Group == common.Config.DefaultGroup, fmt.Sprintf("Sam should be in group %d, not %d", common.Config.DefaultGroup, user.Group))

	// Permanent ban
	duration, _ := time.ParseDuration("0")

	// TODO: Attempt a double ban, double activation, and double unban
	err = user.Ban(duration, 1)
	expectNilErr(t, err)
	expect(t, user.Group == common.Config.DefaultGroup, fmt.Sprintf("Sam should be in group %d, not %d", common.Config.DefaultGroup, user.Group))

	if ucache != nil {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.Get(2)
		recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", newUserID)
	}

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)

	expect(t, user.ID == newUserID, fmt.Sprintf("The UID of the user record should be %d", newUserID))
	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, user.IsBanned, "Sam should be banned")

	expectIntToBeX(t, user.Group, common.BanGroup, "Sam should be in group %d")

	// TODO: Do tests against the scheduled updates table and the task system to make sure the ban exists there and gets revoked when it should

	err = user.Unban()
	expectNilErr(t, err)
	expectIntToBeX(t, user.Group, common.BanGroup, "Sam should still be in the ban group in this copy")

	if ucache != nil {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.Get(newUserID)
		recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", newUserID)
	}

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	expectIntToBeX(t, user.Group, common.Config.DefaultGroup, "Sam should be back in group %d")

	var reportsForumID = 1
	var generalForumID = 2
	dummyResponseRecorder := httptest.NewRecorder()
	bytesBuffer := bytes.NewBuffer([]byte(""))
	dummyRequest1 := httptest.NewRequest("", "/forum/1", bytesBuffer)
	dummyRequest2 := httptest.NewRequest("", "/forum/2", bytesBuffer)

	err = user.ChangeGroup(1)
	expectNilErr(t, err)
	expect(t, user.Group == common.Config.DefaultGroup, "Someone's mutated this pointer elsewhere")

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	var user2 *common.User = common.BlankUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, user.IsAdmin, "Sam should be an admin")
	expect(t, user.IsSuperMod, "Sam should be a super mod")
	expect(t, user.IsMod, "Sam should be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	_, ferr := common.ForumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user.Perms.ViewTopic, "Admins should be able to access the reports forum")
	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")

	err = user.ChangeGroup(2)
	expectNilErr(t, err)
	expect(t, user.Group == 1, "Someone's mutated this pointer elsewhere")

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	user2 = common.BlankUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, user.IsSuperMod, "Sam should be a super mod")
	expect(t, user.IsMod, "Sam should be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user.Perms.ViewTopic, "Mods should be able to access the reports forum")
	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")

	err = user.ChangeGroup(3)
	expectNilErr(t, err)
	expect(t, user.Group == 2, "Someone's mutated this pointer elsewhere")

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	user2 = common.BlankUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, !user.Perms.ViewTopic, "Members shouldn't be able to access the reports forum")
	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")
	expect(t, user.Perms.ViewTopic != user2.Perms.ViewTopic, "user.Perms.ViewTopic and user2.Perms.ViewTopic should never match")

	err = user.ChangeGroup(4)
	expectNilErr(t, err)
	expect(t, user.Group == 3, "Someone's mutated this pointer elsewhere")

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	user2 = common.BlankUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, user.IsBanned, "Sam should be banned")

	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, !user.Perms.ViewTopic, "Members shouldn't be able to access the reports forum")
	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")

	err = user.ChangeGroup(5)
	expectNilErr(t, err)
	expect(t, user.Group == 4, "Someone's mutated this pointer elsewhere")

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	user2 = common.BlankUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, !user.Perms.ViewTopic, "Members shouldn't be able to access the reports forum")
	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")

	err = user.ChangeGroup(6)
	expectNilErr(t, err)
	expect(t, user.Group == 5, "Someone's mutated this pointer elsewhere")

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectIntToBeX(t, user.ID, newUserID, "The UID of the user record should be %d")
	user2 = common.BlankUser()
	*user2 = *user

	expect(t, !user.IsSuperAdmin, "Sam should not be a super admin")
	expect(t, !user.IsAdmin, "Sam should not be an admin")
	expect(t, !user.IsSuperMod, "Sam should not be a super mod")
	expect(t, !user.IsMod, "Sam should not be a mod")
	expect(t, !user.IsBanned, "Sam should not be banned")

	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest1, user, reportsForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, !user.Perms.ViewTopic, "Members shouldn't be able to access the reports forum")
	_, ferr = common.ForumUserCheck(dummyResponseRecorder, dummyRequest2, user2, generalForumID)
	expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
	expect(t, user2.Perms.ViewTopic, "Sam should be able to access the general forum")

	err = user.ChangeGroup(common.Config.DefaultGroup)
	expectNilErr(t, err)
	expect(t, user.Group == 6, "Someone's mutated this pointer elsewhere")

	err = user.Delete()
	expectNilErr(t, err)
	expect(t, !common.Users.Exists(newUserID), fmt.Sprintf("UID #%d should no longer exist", newUserID))

	if ucache != nil {
		expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
		_, err = ucache.Get(newUserID)
		recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", newUserID)
	}

	_, err = common.Users.Get(newUserID)
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
	if !common.PluginsInited {
		common.InitPlugins()
	}

	dummyResponseRecorder := httptest.NewRecorder()
	bytesBuffer := bytes.NewBuffer([]byte(""))
	dummyRequest := httptest.NewRequest("", "/forum/1", bytesBuffer)
	user := common.BlankUser()

	ferr := common.SuperModOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Blank users shouldn't be supermods")

	user.IsSuperMod = false
	ferr = common.SuperModOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Non-supermods shouldn't be allowed through supermod gates")

	user.IsSuperMod = true
	ferr = common.SuperModOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr == nil, "Supermods should be allowed through supermod gates")

	// TODO: Loop over the Control Panel routes and make sure only supermods can get in

	user = common.BlankUser()

	ferr = common.MemberOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Blank users shouldn't be considered loggedin")

	user.Loggedin = false
	ferr = common.MemberOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Guests shouldn't be able to access member areas")

	user.Loggedin = true
	ferr = common.MemberOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr == nil, "Logged in users should be able to access member areas")

	// TODO: Loop over the /user/ routes and make sure only members can access the ones other than /user/username

	// TODO: Write tests for AdminOnly()

	user = common.BlankUser()

	ferr = common.SuperAdminOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Blank users shouldn't be considered super admins")

	user.IsSuperAdmin = false
	ferr = common.SuperAdminOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Non-super admins shouldn't be allowed through the super admin gate")

	user.IsSuperAdmin = true
	ferr = common.SuperAdminOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr == nil, "Super admins should be allowed through super admin gates")

	// TODO: Make sure only super admins can access the backups route

	//dummyResponseRecorder = httptest.NewRecorder()
	//bytesBuffer = bytes.NewBuffer([]byte(""))
	//dummyRequest = httptest.NewRequest("", "/panel/backups/", bytesBuffer)

}

func TestTopicStore(t *testing.T) {
	if !gloinited {
		err := gloinit()
		if err != nil {
			t.Fatal(err)
		}
	}
	if !common.PluginsInited {
		common.InitPlugins()
	}

	var err error
	tcache := common.NewMemoryTopicCache(common.Config.TopicCacheCapacity)
	common.Topics, err = common.NewDefaultTopicStore(tcache)
	expectNilErr(t, err)
	topicStoreTest(t)
	common.Topics, err = common.NewDefaultTopicStore(nil)
	expectNilErr(t, err)
	topicStoreTest(t)
}
func topicStoreTest(t *testing.T) {
	var topic *common.Topic
	var err error

	_, err = common.Topics.Get(-1)
	recordMustNotExist(t, err, "TID #-1 shouldn't exist")

	_, err = common.Topics.Get(0)
	recordMustNotExist(t, err, "TID #0 shouldn't exist")

	topic, err = common.Topics.Get(1)
	recordMustExist(t, err, "Couldn't find TID #1")

	if topic.ID != 1 {
		t.Error("topic.ID does not match the requested TID. Got '%d' instead.", topic.ID)
	}

	// TODO: Add BulkGetMap() to the TopicStore

	ok := common.Topics.Exists(-1)
	expect(t, !ok, "TID #-1 shouldn't exist")

	ok = common.Topics.Exists(0)
	expect(t, !ok, "TID #0 shouldn't exist")

	ok = common.Topics.Exists(1)
	expect(t, ok, "TID #1 should exist")

	count := common.Topics.GlobalCount()
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
	if !common.PluginsInited {
		common.InitPlugins()
	}

	_, err := common.Forums.Get(-1)
	recordMustNotExist(t, err, "FID #-1 shouldn't exist")

	_, err = common.Forums.Get(0)
	recordMustNotExist(t, err, "FID #0 shouldn't exist")

	forum, err := common.Forums.Get(1)
	recordMustExist(t, err, "Couldn't find FID #1")

	if forum.ID != 1 {
		t.Error("forum.ID doesn't not match the requested FID. Got '%d' instead.'", forum.ID)
	}
	// TODO: Check the preset and forum permissions
	expect(t, forum.Name == "Reports", fmt.Sprintf("FID #0 is named '%s' and not 'Reports'", forum.Name))
	expect(t, !forum.Active, fmt.Sprintf("The reports forum shouldn't be active"))
	var expectDesc = "All the reports go here"
	expect(t, forum.Desc == expectDesc, fmt.Sprintf("The forum description should be '%s' not '%s'", expectDesc, forum.Desc))

	forum, err = common.Forums.Get(2)
	recordMustExist(t, err, "Couldn't find FID #1")

	expect(t, forum.ID == 2, fmt.Sprintf("The FID should be 2 not %d", forum.ID))
	expect(t, forum.Name == "General", fmt.Sprintf("The name of the forum should be 'General' not '%s'", forum.Name))
	expect(t, forum.Active, fmt.Sprintf("The general forum should be active"))
	expectDesc = "A place for general discussions which don't fit elsewhere"
	expect(t, forum.Desc == expectDesc, fmt.Sprintf("The forum description should be '%s' not '%s'", expectDesc, forum.Desc))

	ok := common.Forums.Exists(-1)
	expect(t, !ok, "FID #-1 shouldn't exist")
	ok = common.Forums.Exists(0)
	expect(t, !ok, "FID #0 shouldn't exist")
	ok = common.Forums.Exists(1)
	expect(t, ok, "FID #1 should exist")
	ok = common.Forums.Exists(2)
	expect(t, ok, "FID #2 should exist")
	ok = common.Forums.Exists(3)
	expect(t, !ok, "FID #3 shouldn't exist")

	fid, err := common.Forums.Create("Test Forum", "", true, "all")
	expectNilErr(t, err)
	expect(t, fid == 3, "The first forum we create should have an ID of 3")
	ok = common.Forums.Exists(3)
	expect(t, ok, "FID #2 should exist")

	forum, err = common.Forums.Get(3)
	recordMustExist(t, err, "Couldn't find FID #3")

	expect(t, forum.ID == 2, fmt.Sprintf("The FID should be 3 not %d", forum.ID))
	expect(t, forum.Name == "Test Forum", fmt.Sprintf("The name of the forum should be 'Test Forum' not '%s'", forum.Name))
	expect(t, forum.Active, fmt.Sprintf("The test forum should be active"))
	expect(t, forum.Desc == "", fmt.Sprintf("The forum description should be blank not '%s'", expectDesc, forum.Desc))

	// TODO: More forum creation tests
	// TODO: Test forum deletion
	// TODO: Test forum update
}

// TODO: Implement this
func TestForumPermsStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !common.PluginsInited {
		common.InitPlugins()
	}
}

// TODO: Test the group permissions
func TestGroupStore(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !common.PluginsInited {
		common.InitPlugins()
	}

	_, err := common.Groups.Get(-1)
	recordMustNotExist(t, err, "GID #-1 shouldn't exist")

	// TODO: Refactor the group store to remove GID #0
	group, err := common.Groups.Get(0)
	recordMustExist(t, err, "Couldn't find GID #0")

	expect(t, group.ID == 0, fmt.Sprintf("group.ID doesn't not match the requested GID. Got '%d' instead.", group.ID))
	expect(t, group.Name == "Unknown", fmt.Sprintf("GID #0 is named '%s' and not 'Unknown'", group.Name))

	group, err = common.Groups.Get(1)
	recordMustExist(t, err, "Couldn't find GID #1")
	expect(t, group.ID == 1, fmt.Sprintf("group.ID doesn't not match the requested GID. Got '%d' instead.'", group.ID))

	ok := common.Groups.Exists(-1)
	expect(t, !ok, "GID #-1 shouldn't exist")

	// 0 aka Unknown, for system posts and other oddities
	ok = common.Groups.Exists(0)
	expect(t, ok, "GID #0 should exist")

	ok = common.Groups.Exists(1)
	expect(t, ok, "GID #1 should exist")

	var isAdmin = true
	var isMod = true
	var isBanned = false
	gid, err := common.Groups.Create("Testing", "Test", isAdmin, isMod, isBanned)
	expectNilErr(t, err)
	expect(t, common.Groups.Exists(gid), "The group we just made doesn't exist")

	group, err = common.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, group.IsAdmin, "This should be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	isAdmin = false
	isMod = true
	isBanned = true
	gid, err = common.Groups.Create("Testing 2", "Test", isAdmin, isMod, isBanned)
	expectNilErr(t, err)
	expect(t, common.Groups.Exists(gid), "The group we just made doesn't exist")

	group, err = common.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This should not be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	// TODO: Make sure this pointer doesn't change once we refactor the group store to stop updating the pointer
	err = group.ChangeRank(false, false, true)
	expectNilErr(t, err)

	group, err = common.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This shouldn't be an admin group")
	expect(t, !group.IsMod, "This shouldn't be a mod group")
	expect(t, group.IsBanned, "This should be a ban group")

	err = group.ChangeRank(true, true, true)
	expectNilErr(t, err)

	group, err = common.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, group.IsAdmin, "This should be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	err = group.ChangeRank(false, true, true)
	expectNilErr(t, err)

	group, err = common.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This shouldn't be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	// Make sure the data is static
	common.Groups.Reload(gid)

	group, err = common.Groups.Get(gid)
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
	if !common.PluginsInited {
		common.InitPlugins()
	}

	_, err := common.Rstore.Get(-1)
	recordMustNotExist(t, err, "RID #-1 shouldn't exist")

	_, err = common.Rstore.Get(0)
	recordMustNotExist(t, err, "RID #0 shouldn't exist")

	reply, err := common.Rstore.Get(1)
	expectNilErr(t, err)
	expect(t, reply.ID == 1, fmt.Sprintf("RID #1 has the wrong ID. It should be 1 not %d", reply.ID))
	expect(t, reply.ParentID == 1, fmt.Sprintf("The parent topic of RID #1 should be 1 not %d", reply.ParentID))
	expect(t, reply.CreatedBy == 1, fmt.Sprintf("The creator of RID #1 should be 1 not %d", reply.CreatedBy))
	expect(t, reply.Content == "A reply!", fmt.Sprintf("The contents of RID #1 should be 'A reply!' not %s", reply.Content))
	expect(t, reply.IPAddress == "::1", fmt.Sprintf("The IPAddress of RID#1 should be '::1' not %s", reply.IPAddress))

	_, err = common.Rstore.Get(2)
	recordMustNotExist(t, err, "RID #2 shouldn't exist")

	// TODO: Test Create and Get
	//Create(tid int, content string, ipaddress string, fid int, uid int) (id int, err error)
	topic, err := common.Topics.Get(1)
	expectNilErr(t, err)
	rid, err := common.Rstore.Create(topic, "Fofofo", "::1", 1)
	expectNilErr(t, err)
	expect(t, rid == 2, fmt.Sprintf("The next reply ID should be 2 not %d", rid))

	reply, err = common.Rstore.Get(2)
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
	if !common.PluginsInited {
		common.InitPlugins()
	}

	_, err := common.Prstore.Get(-1)
	recordMustNotExist(t, err, "PRID #-1 shouldn't exist")

	_, err = common.Prstore.Get(0)
	recordMustNotExist(t, err, "PRID #0 shouldn't exist")

	_, err = common.Prstore.Get(1)
	recordMustNotExist(t, err, "PRID #1 shouldn't exist")

	var profileID = 1
	prid, err := common.Prstore.Create(profileID, "Haha", 1, "::1")
	expect(t, err == nil, "Unable to create a profile reply")
	expect(t, prid == 1, "The first profile reply should have an ID of 1")

	profileReply, err := common.Prstore.Get(1)
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
	common.Config.BuildSlugs = true // Flip this switch, otherwise all the tests will fail

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
		res = common.NameToSlug(item.Msg)
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
	t.Logf("Set realPassword to '%s'", realPassword)
	t.Log("Hashing the real password")
	hashedPassword, err = common.BcryptGeneratePasswordNoSalt(realPassword)
	if err != nil {
		t.Error(err)
	}

	password = realPassword
	t.Logf("Testing password '%s'", password)
	t.Logf("Testing salt '%s'", salt)
	err = common.CheckPassword(hashedPassword, password, salt)
	if err == common.ErrMismatchedHashAndPassword {
		t.Error("The two don't match")
	} else if err == common.ErrPasswordTooLong {
		t.Error("CheckPassword thinks the password is too long")
	} else if err != nil {
		t.Error(err)
	}

	password = "hahaha"
	t.Logf("Testing password '%s'", password)
	t.Logf("Testing salt '%s'", salt)
	err = common.CheckPassword(hashedPassword, password, salt)
	if err == common.ErrPasswordTooLong {
		t.Error("CheckPassword thinks the password is too long")
	} else if err == nil {
		t.Error("The two shouldn't match!")
	}

	password = "Madame Cassandra's Mystic"
	t.Logf("Testing password '%s'", password)
	t.Logf("Testing salt '%s'", salt)
	err = common.CheckPassword(hashedPassword, password, salt)
	expect(t, err != common.ErrPasswordTooLong, "CheckPassword thinks the password is too long")
	expect(t, err != nil, "The two shouldn't match!")
}
