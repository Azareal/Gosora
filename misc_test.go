package main

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func miscinit(t *testing.T) {
	err := gloinit()
	if err != nil {
		t.Fatal(err)
	}
}

func recordMustExist(t *testing.T, err error, errmsg string, args ...interface{}) {
	if err == ErrNoRows {
		debug.PrintStack()
		t.Errorf(errmsg, args...)
	} else if err != nil {
		debug.PrintStack()
		t.Fatal(err)
	}
}

func recordMustNotExist(t *testing.T, err error, errmsg string, args ...interface{}) {
	if err == nil {
		debug.PrintStack()
		t.Errorf(errmsg, args...)
	} else if err != ErrNoRows {
		debug.PrintStack()
		t.Fatal(err)
	}
}

func TestUserStore(t *testing.T) {
	miscinit(t)
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
	userStoreTest(t, 5)
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

	var expectW = func(cond bool, expec bool, prefix string, suffix string) {
		midfix := "should not be"
		if expec {
			midfix = "should be"
		}
		expect(t, cond, prefix+" "+midfix+" "+suffix)
	}

	// TODO: Add email checks too? Do them seperately?
	var expectUser = func(user *common.User, uid int, name string, group int, super bool, admin bool, mod bool, banned bool) {
		expect(t, user.ID == uid, fmt.Sprintf("user.ID should be %d. Got '%d' instead.", uid, user.ID))
		expect(t, user.Name == name, fmt.Sprintf("user.Name should be '%s', not '%s'", name, user.Name))
		expectW(user.Group == group, true, user.Name, "in group"+strconv.Itoa(group))
		expectW(user.IsSuperAdmin == super, super, user.Name, "a super admin")
		expectW(user.IsAdmin == admin, admin, user.Name, "an admin")
		expectW(user.IsSuperMod == mod, mod, user.Name, "a super mod")
		expectW(user.IsMod == mod, mod, user.Name, "a mod")
		expectW(user.IsBanned == banned, banned, user.Name, "banned")
	}
	expectUser(user, 1, "Admin", 1, true, true, true, false)

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
	expect(t, len(userList) == 1, fmt.Sprintf("Returned map should have one result (UID #1), not %d", len(userList)))

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
	// TODO: Write tests for the registration validators
	uid, err := common.Users.Create("Sam", "ReallyBadPassword", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expect(t, uid == newUserID, fmt.Sprintf("The UID of the new user should be %d not %d", newUserID, uid))
	expect(t, common.Users.Exists(newUserID), fmt.Sprintf("UID #%d should exist", newUserID))

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", 5, false, false, false, false)

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
	var afterUserFlush = func(uid int) {
		if ucache != nil {
			expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
			_, err = ucache.Get(uid)
			recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", uid)
		}
	}
	afterUserFlush(newUserID)

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", common.Config.DefaultGroup, false, false, false, false)

	// Permanent ban
	duration, _ := time.ParseDuration("0")

	// TODO: Attempt a double ban, double activation, and double unban
	err = user.Ban(duration, 1)
	expectNilErr(t, err)
	expect(t, user.Group == common.Config.DefaultGroup, fmt.Sprintf("Sam should be in group %d, not %d", common.Config.DefaultGroup, user.Group))
	afterUserFlush(newUserID)

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", common.BanGroup, false, false, false, true)

	// TODO: Do tests against the scheduled updates table and the task system to make sure the ban exists there and gets revoked when it should

	err = user.Unban()
	expectNilErr(t, err)
	expectIntToBeX(t, user.Group, common.BanGroup, "Sam should still be in the ban group in this copy")
	afterUserFlush(newUserID)

	user, err = common.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", common.Config.DefaultGroup, false, false, false, false)

	var reportsForumID = 1 // TODO: Use the constant in common?
	var generalForumID = 2
	dummyResponseRecorder := httptest.NewRecorder()
	bytesBuffer := bytes.NewBuffer([]byte(""))
	dummyRequest1 := httptest.NewRequest("", "/forum/"+strconv.Itoa(reportsForumID), bytesBuffer)
	dummyRequest2 := httptest.NewRequest("", "/forum/"+strconv.Itoa(generalForumID), bytesBuffer)
	var user2 *common.User

	var changeGroupTest = func(oldGroup int, newGroup int) {
		err = user.ChangeGroup(newGroup)
		expectNilErr(t, err)
		// ! I don't think ChangeGroup should be changing the value of user... Investigate this.
		expect(t, oldGroup == user.Group, "Someone's mutated this pointer elsewhere")

		user, err = common.Users.Get(newUserID)
		recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
		user2 = common.BlankUser()
		*user2 = *user
	}

	var changeGroupTest2 = func(rank string, firstShouldBe bool, secondShouldBe bool) {
		head, err := common.UserCheck(dummyResponseRecorder, dummyRequest1, user)
		if err != nil {
			t.Fatal(err)
		}
		head2, err := common.UserCheck(dummyResponseRecorder, dummyRequest2, user2)
		if err != nil {
			t.Fatal(err)
		}
		ferr := common.ForumUserCheck(head, dummyResponseRecorder, dummyRequest1, user, reportsForumID)
		expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
		expect(t, user.Perms.ViewTopic == firstShouldBe, rank+" should be able to access the reports forum")
		ferr = common.ForumUserCheck(head2, dummyResponseRecorder, dummyRequest2, user2, generalForumID)
		expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
		expect(t, user2.Perms.ViewTopic == secondShouldBe, "Sam should be able to access the general forum")
	}

	changeGroupTest(common.Config.DefaultGroup, 1)
	expectUser(user, newUserID, "Sam", 1, false, true, true, false)
	changeGroupTest2("Admins", true, true)

	changeGroupTest(1, 2)
	expectUser(user, newUserID, "Sam", 2, false, false, true, false)
	changeGroupTest2("Mods", true, true)

	changeGroupTest(2, 3)
	expectUser(user, newUserID, "Sam", 3, false, false, false, false)
	changeGroupTest2("Members", false, true)
	expect(t, user.Perms.ViewTopic != user2.Perms.ViewTopic, "user.Perms.ViewTopic and user2.Perms.ViewTopic should never match")

	changeGroupTest(3, 4)
	expectUser(user, newUserID, "Sam", 4, false, false, false, true)
	changeGroupTest2("Members", false, true)

	changeGroupTest(4, 5)
	expectUser(user, newUserID, "Sam", 5, false, false, false, false)
	changeGroupTest2("Members", false, true)

	changeGroupTest(5, 6)
	expectUser(user, newUserID, "Sam", 6, false, false, false, false)
	changeGroupTest2("Members", false, true)

	err = user.ChangeGroup(common.Config.DefaultGroup)
	expectNilErr(t, err)
	expect(t, user.Group == 6, "Someone's mutated this pointer elsewhere")

	err = user.Delete()
	expectNilErr(t, err)
	expect(t, !common.Users.Exists(newUserID), fmt.Sprintf("UID #%d should no longer exist", newUserID))
	afterUserFlush(newUserID)

	_, err = common.Users.Get(newUserID)
	recordMustNotExist(t, err, "UID #%d shouldn't exist", newUserID)

	// And a unicode test, even though I doubt it'll fail
	uid, err = common.Users.Create("サム", "😀😀😀", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expect(t, uid == newUserID+1, fmt.Sprintf("The UID of the new user should be %d", newUserID+1))
	expect(t, common.Users.Exists(newUserID+1), fmt.Sprintf("UID #%d should exist", newUserID+1))

	user, err = common.Users.Get(newUserID + 1)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID+1)
	expectUser(user, newUserID+1, "サム", 5, false, false, false, false)

	err = user.Delete()
	expectNilErr(t, err)
	expect(t, !common.Users.Exists(newUserID+1), fmt.Sprintf("UID #%d should no longer exist", newUserID+1))

	// MySQL utf8mb4 username test
	uid, err = common.Users.Create("😀😀😀", "😀😀😀", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expect(t, uid == newUserID+2, fmt.Sprintf("The UID of the new user should be %d", newUserID+2))
	expect(t, common.Users.Exists(newUserID+2), fmt.Sprintf("UID #%d should exist", newUserID+2))

	user, err = common.Users.Get(newUserID + 2)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID+1)
	expectUser(user, newUserID+2, "😀😀😀", 5, false, false, false, false)

	err = user.Delete()
	expectNilErr(t, err)
	expect(t, !common.Users.Exists(newUserID+2), fmt.Sprintf("UID #%d should no longer exist", newUserID+2))

	// TODO: Add unicode login tests somewhere? Probably with the rest of the auth tests
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
	miscinit(t)
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
	miscinit(t)
	if !common.PluginsInited {
		common.InitPlugins()
	}

	var err error
	tcache := common.NewMemoryTopicCache(common.Config.TopicCacheCapacity)
	common.Topics, err = common.NewDefaultTopicStore(tcache)
	expectNilErr(t, err)
	topicStoreTest(t, 2)
	common.Topics, err = common.NewDefaultTopicStore(nil)
	expectNilErr(t, err)
	topicStoreTest(t, 3)
}
func topicStoreTest(t *testing.T, newID int) {
	var topic *common.Topic
	var err error

	_, err = common.Topics.Get(-1)
	recordMustNotExist(t, err, "TID #-1 shouldn't exist")
	_, err = common.Topics.Get(0)
	recordMustNotExist(t, err, "TID #0 shouldn't exist")

	topic, err = common.Topics.Get(1)
	recordMustExist(t, err, "Couldn't find TID #1")
	expect(t, topic.ID == 1, fmt.Sprintf("topic.ID does not match the requested TID. Got '%d' instead.", topic.ID))

	// TODO: Add BulkGetMap() to the TopicStore

	ok := common.Topics.Exists(-1)
	expect(t, !ok, "TID #-1 shouldn't exist")
	ok = common.Topics.Exists(0)
	expect(t, !ok, "TID #0 shouldn't exist")
	ok = common.Topics.Exists(1)
	expect(t, ok, "TID #1 should exist")

	count := common.Topics.GlobalCount()
	expect(t, count == 1, fmt.Sprintf("Global count for topics should be 1, not %d", count))

	//Create(fid int, topicName string, content string, uid int, ipaddress string) (tid int, err error)
	tid, err := common.Topics.Create(2, "Test Topic", "Topic Content", 1, "::1")
	expectNilErr(t, err)
	expect(t, tid == newID, fmt.Sprintf("TID for the new topic should be %d, not %d", newID, tid))
	expect(t, common.Topics.Exists(newID), fmt.Sprintf("TID #%d should exist", newID))

	count = common.Topics.GlobalCount()
	expect(t, count == 2, fmt.Sprintf("Global count for topics should be 2, not %d", count))

	var iFrag = func(cond bool) string {
		if !cond {
			return "n't"
		}
		return ""
	}

	var testTopic = func(tid int, title string, content string, createdBy int, ip string, parentID int, isClosed bool, sticky bool) {
		topic, err = common.Topics.Get(tid)
		recordMustExist(t, err, fmt.Sprintf("Couldn't find TID #%d", tid))
		expect(t, topic.ID == tid, fmt.Sprintf("topic.ID does not match the requested TID. Got '%d' instead.", topic.ID))
		expect(t, topic.GetID() == tid, fmt.Sprintf("topic.ID does not match the requested TID. Got '%d' instead.", topic.GetID()))
		expect(t, topic.Title == title, fmt.Sprintf("The topic's name should be '%s', not %s", title, topic.Title))
		expect(t, topic.Content == content, fmt.Sprintf("The topic's body should be '%s', not %s", content, topic.Content))
		expect(t, topic.CreatedBy == createdBy, fmt.Sprintf("The topic's creator should be %d, not %d", createdBy, topic.CreatedBy))
		expect(t, topic.IPAddress == ip, fmt.Sprintf("The topic's IP Address should be '%s', not %s", ip, topic.IPAddress))
		expect(t, topic.ParentID == parentID, fmt.Sprintf("The topic's parent forum should be %d, not %d", parentID, topic.ParentID))
		expect(t, topic.IsClosed == isClosed, fmt.Sprintf("This topic should%s be locked", iFrag(topic.IsClosed)))
		expect(t, topic.Sticky == sticky, fmt.Sprintf("This topic should%s be sticky", iFrag(topic.Sticky)))
		expect(t, topic.GetTable() == "topics", fmt.Sprintf("The topic's table should be 'topics', not %s", topic.GetTable()))
	}

	tcache := common.Topics.GetCache()
	var shouldNotBeIn = func(tid int) {
		if tcache != nil {
			_, err = tcache.Get(tid)
			recordMustNotExist(t, err, "Topic cache should be empty")
		}
	}
	if tcache != nil {
		_, err = tcache.Get(newID)
		expectNilErr(t, err)
	}

	testTopic(newID, "Test Topic", "Topic Content", 1, "::1", 2, false, false)

	expectNilErr(t, topic.Lock())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, "::1", 2, true, false)

	expectNilErr(t, topic.Unlock())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, "::1", 2, false, false)

	expectNilErr(t, topic.Stick())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, "::1", 2, false, true)

	expectNilErr(t, topic.Unstick())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, "::1", 2, false, false)

	expectNilErr(t, topic.MoveTo(1))
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, "::1", 1, false, false)
	// TODO: Add more tests for more *Topic methods

	expectNilErr(t, topic.Delete())
	shouldNotBeIn(newID)

	_, err = common.Topics.Get(newID)
	recordMustNotExist(t, err, fmt.Sprintf("TID #%d shouldn't exist", newID))
	expect(t, !common.Topics.Exists(newID), fmt.Sprintf("TID #%d shouldn't exist", newID))

	// TODO: Test topic creation and retrieving that created topic plus reload and inspecting the cache
}

func TestForumStore(t *testing.T) {
	miscinit(t)
	if !common.PluginsInited {
		common.InitPlugins()
	}

	_, err := common.Forums.Get(-1)
	recordMustNotExist(t, err, "FID #-1 shouldn't exist")
	_, err = common.Forums.Get(0)
	recordMustNotExist(t, err, "FID #0 shouldn't exist")

	forum, err := common.Forums.Get(1)
	recordMustExist(t, err, "Couldn't find FID #1")
	expect(t, forum.ID == 1, fmt.Sprintf("forum.ID doesn't not match the requested FID. Got '%d' instead.'", forum.ID))
	// TODO: Check the preset and forum permissions
	expect(t, forum.Name == "Reports", fmt.Sprintf("FID #0 is named '%s' and not 'Reports'", forum.Name))
	expect(t, !forum.Active, fmt.Sprintf("The reports forum shouldn't be active"))
	var expectDesc = "All the reports go here"
	expect(t, forum.Desc == expectDesc, fmt.Sprintf("The forum description should be '%s' not '%s'", expectDesc, forum.Desc))

	forum, err = common.Forums.Get(2)
	recordMustExist(t, err, "Couldn't find FID #2")

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

	expect(t, forum.ID == 3, fmt.Sprintf("The FID should be 3 not %d", forum.ID))
	expect(t, forum.Name == "Test Forum", fmt.Sprintf("The name of the forum should be 'Test Forum' not '%s'", forum.Name))
	expect(t, forum.Active, fmt.Sprintf("The test forum should be active"))
	expect(t, forum.Desc == "", fmt.Sprintf("The forum description should be blank not '%s'", forum.Desc))

	// TODO: More forum creation tests
	// TODO: Test forum deletion
	// TODO: Test forum update
}

// TODO: Implement this
func TestForumPermsStore(t *testing.T) {
	miscinit(t)
	if !common.PluginsInited {
		common.InitPlugins()
	}
}

// TODO: Test the group permissions
func TestGroupStore(t *testing.T) {
	miscinit(t)
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

	expect(t, !common.Groups.Exists(-1), "GID #-1 shouldn't exist")
	// 0 aka Unknown, for system posts and other oddities
	expect(t, common.Groups.Exists(0), "GID #0 should exist")
	expect(t, common.Groups.Exists(1), "GID #1 should exist")

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
	miscinit(t)
	if !common.PluginsInited {
		common.InitPlugins()
	}

	_, err := common.Rstore.Get(-1)
	recordMustNotExist(t, err, "RID #-1 shouldn't exist")
	_, err = common.Rstore.Get(0)
	recordMustNotExist(t, err, "RID #0 shouldn't exist")

	var replyTest = func(rid int, parentID int, createdBy int, content string, ip string) {
		reply, err := common.Rstore.Get(rid)
		expectNilErr(t, err)
		expect(t, reply.ID == rid, fmt.Sprintf("RID #%d has the wrong ID. It should be %d not %d", rid, rid, reply.ID))
		expect(t, reply.ParentID == parentID, fmt.Sprintf("The parent topic of RID #%d should be %d not %d", rid, parentID, reply.ParentID))
		expect(t, reply.CreatedBy == createdBy, fmt.Sprintf("The creator of RID #%d should be %d not %d", rid, createdBy, reply.CreatedBy))
		expect(t, reply.Content == content, fmt.Sprintf("The contents of RID #%d should be '%s' not %s", rid, content, reply.Content))
		expect(t, reply.IPAddress == ip, fmt.Sprintf("The IPAddress of RID#%d should be '%s' not %s", rid, ip, reply.IPAddress))
	}
	replyTest(1, 1, 1, "A reply!", "::1")

	_, err = common.Rstore.Get(2)
	recordMustNotExist(t, err, "RID #2 shouldn't exist")

	// TODO: Test Create and Get
	//Create(tid int, content string, ipaddress string, fid int, uid int) (id int, err error)
	topic, err := common.Topics.Get(1)
	expectNilErr(t, err)
	expect(t, topic.PostCount == 1, fmt.Sprintf("TID #1's post count should be one, not %d", topic.PostCount))
	rid, err := common.Rstore.Create(topic, "Fofofo", "::1", 1)
	expectNilErr(t, err)
	expect(t, rid == 2, fmt.Sprintf("The next reply ID should be 2 not %d", rid))
	expect(t, topic.PostCount == 1, fmt.Sprintf("The old TID #1 in memory's post count should be one, not %d", topic.PostCount))
	// TODO: Test the reply count on the topic

	replyTest(2, 1, 1, "Fofofo", "::1")

	topic, err = common.Topics.Get(1)
	expectNilErr(t, err)
	expect(t, topic.PostCount == 2, fmt.Sprintf("TID #1's post count should be two, not %d", topic.PostCount))

	err = topic.CreateActionReply("destroy", "::1", 1)
	expectNilErr(t, err)
	expect(t, topic.PostCount == 2, fmt.Sprintf("The old TID #1 in memory's post count should be two, not %d", topic.PostCount))
	replyTest(3, 1, 1, "", "::1")
	// TODO: Check the actionType field of the reply, this might not be loaded by TopicStore, maybe we should add it there?

	topic, err = common.Topics.Get(1)
	expectNilErr(t, err)
	expect(t, topic.PostCount == 3, fmt.Sprintf("TID #1's post count should be three, not %d", topic.PostCount))
}

func TestProfileReplyStore(t *testing.T) {
	miscinit(t)
	if !common.PluginsInited {
		common.InitPlugins()
	}

	_, err := common.Prstore.Get(-1)
	recordMustNotExist(t, err, "PRID #-1 shouldn't exist")
	_, err = common.Prstore.Get(0)
	recordMustNotExist(t, err, "PRID #0 shouldn't exist")
	_, err = common.Prstore.Get(1)
	recordMustNotExist(t, err, "PRID #1 shouldn't exist")

	// ? - Commented this one out as strong constraints like this put an unreasonable load on the database, we only want errors if a delete which should succeed fails
	//profileReply := common.BlankProfileReply(1)
	//err = profileReply.Delete()
	//expect(t,err != nil,"You shouldn't be able to delete profile replies which don't exist")

	var profileID = 1
	prid, err := common.Prstore.Create(profileID, "Haha", 1, "::1")
	expectNilErr(t, err)
	expect(t, prid == 1, "The first profile reply should have an ID of 1")

	profileReply, err := common.Prstore.Get(1)
	expectNilErr(t, err)
	expect(t, profileReply.ID == 1, fmt.Sprintf("The profile reply should have an ID of 1 not %d", profileReply.ID))
	expect(t, profileReply.ParentID == 1, fmt.Sprintf("The parent ID of the profile reply should be 1 not %d", profileReply.ParentID))
	expect(t, profileReply.Content == "Haha", fmt.Sprintf("The profile reply's contents should be 'Haha' not '%s'", profileReply.Content))
	expect(t, profileReply.CreatedBy == 1, fmt.Sprintf("The profile reply's creator should be 1 not %d", profileReply.CreatedBy))
	expect(t, profileReply.IPAddress == "::1", fmt.Sprintf("The profile reply's IP Address should be '::1' not '%s'", profileReply.IPAddress))

	err = profileReply.Delete()
	expectNilErr(t, err)
	_, err = common.Prstore.Get(1)
	expect(t, err != nil, "PRID #1 shouldn't exist after being deleted")

	// TODO: Test profileReply.SetBody() and profileReply.Creator()
}

func TestLogs(t *testing.T) {
	miscinit(t)
	gTests := func(store common.LogStore, phrase string) {
		expect(t, store.GlobalCount() == 0, "There shouldn't be any "+phrase)
		logs, err := store.GetOffset(0, 25)
		expectNilErr(t, err)
		expect(t, len(logs) == 0, "The log slice should be empty")
	}
	gTests(common.ModLogs, "modlogs")
	gTests(common.AdminLogs, "adminlogs")

	gTests2 := func(store common.LogStore, phrase string) {
		err := store.Create("something", 0, "bumblefly", "::1", 1)
		expectNilErr(t, err)
		count := store.GlobalCount()
		expect(t, count == 1, fmt.Sprintf("store.GlobalCount should return one, not %d", count))
		logs, err := store.GetOffset(0, 25)
		recordMustExist(t, err, "We should have at-least one "+phrase)
		expect(t, len(logs) == 1, "The length of the log slice should be one")

		log := logs[0]
		expect(t, log.Action == "something", "log.Action is not something")
		expect(t, log.ElementID == 0, "log.ElementID is not 0")
		expect(t, log.ElementType == "bumblefly", "log.ElementType is not bumblefly")
		expect(t, log.IPAddress == "::1", "log.IPAddress is not ::1")
		expect(t, log.ActorID == 1, "log.ActorID is not 1")
		// TODO: Add a test for log.DoneAt? Maybe throw in some dates and times which are clearly impossible but which may occur due to timezone bugs?
	}
	gTests2(common.ModLogs, "modlog")
	gTests2(common.AdminLogs, "adminlog")
}

// TODO: Add tests for registration logs

func TestPluginManager(t *testing.T) {
	miscinit(t)
	if !common.PluginsInited {
		common.InitPlugins()
	}

	_, ok := common.Plugins["fairy-dust"]
	expect(t, !ok, "Plugin fairy-dust shouldn't exist")
	plugin, ok := common.Plugins["bbcode"]
	expect(t, ok, "Plugin bbcode should exist")
	expect(t, !plugin.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !plugin.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !plugin.Active, "Plugin bbcode shouldn't be active")
	active, err := plugin.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err := plugin.InDatabase()
	expectNilErr(t, err)
	expect(t, !hasPlugin, "Plugin bbcode shouldn't exist in the database")
	// TODO: Add some test cases for SetActive and SetInstalled before calling AddToDatabase

	expectNilErr(t, plugin.AddToDatabase(true, false))
	expect(t, !plugin.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !plugin.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, plugin.Active, "Plugin bbcode should be active")
	active, err = plugin.BypassActive()
	expectNilErr(t, err)
	expect(t, active, "Plugin bbcode should be active in the database too")
	hasPlugin, err = plugin.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should exist in the database")
	expect(t, plugin.Init != nil, "Plugin bbcode should have an init function")
	expectNilErr(t, plugin.Init())

	expectNilErr(t, plugin.SetActive(true))
	expect(t, !plugin.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !plugin.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, plugin.Active, "Plugin bbcode should still be active")
	active, err = plugin.BypassActive()
	expectNilErr(t, err)
	expect(t, active, "Plugin bbcode should still be active in the database too")
	hasPlugin, err = plugin.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	expectNilErr(t, plugin.SetActive(false))
	expect(t, !plugin.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !plugin.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !plugin.Active, "Plugin bbcode shouldn't be active")
	active, err = plugin.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database")
	hasPlugin, err = plugin.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")
	expect(t, plugin.Deactivate != nil, "Plugin bbcode should have an init function")
	plugin.Deactivate() // Returns nothing

	// Not installable, should not be mutated
	expect(t, plugin.SetInstalled(true) == common.ErrPluginNotInstallable, "Plugin was set as installed despite not being installable")
	expect(t, !plugin.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !plugin.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !plugin.Active, "Plugin bbcode shouldn't be active")
	active, err = plugin.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = plugin.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	expect(t, plugin.SetInstalled(false) == common.ErrPluginNotInstallable, "Plugin was set as not installed despite not being installable")
	expect(t, !plugin.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !plugin.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !plugin.Active, "Plugin bbcode shouldn't be active")
	active, err = plugin.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = plugin.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	// This isn't really installable, but we want to get a few tests done before getting plugins which are stateful
	plugin.Installable = true
	expectNilErr(t, plugin.SetInstalled(true))
	expect(t, plugin.Installable, "Plugin bbcode should be installable")
	expect(t, plugin.Installed, "Plugin bbcode should be 'installed'")
	expect(t, !plugin.Active, "Plugin bbcode shouldn't be active")
	active, err = plugin.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = plugin.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	expectNilErr(t, plugin.SetInstalled(false))
	expect(t, plugin.Installable, "Plugin bbcode should be installable")
	expect(t, !plugin.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !plugin.Active, "Plugin bbcode shouldn't be active")
	active, err = plugin.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = plugin.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	// Bugs sometimes arise when we try to delete a hook when there are multiple, so test for that
	// TODO: Do a finer grained test for that case...? A bigger test might catch more odd cases with multiple plugins
	plugin2, ok := common.Plugins["markdown"]
	expect(t, ok, "Plugin markdown should exist")
	expect(t, !plugin2.Installable, "Plugin markdown shouldn't be installable")
	expect(t, !plugin2.Installed, "Plugin markdown shouldn't be 'installed'")
	expect(t, !plugin2.Active, "Plugin markdown shouldn't be active")
	active, err = plugin2.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin markdown shouldn't be active in the database either")
	hasPlugin, err = plugin2.InDatabase()
	expectNilErr(t, err)
	expect(t, !hasPlugin, "Plugin markdown shouldn't exist in the database")

	expectNilErr(t, plugin2.AddToDatabase(true, false))
	expectNilErr(t, plugin2.Init())
	expectNilErr(t, plugin.SetActive(true))
	expectNilErr(t, plugin.Init())
	plugin2.Deactivate()
	expectNilErr(t, plugin2.SetActive(false))
	plugin.Deactivate()
	expectNilErr(t, plugin.SetActive(false))

	// Hook tests
	expect(t, common.GetHookTable().Sshook("haha", "ho") == "ho", "Sshook shouldn't have anything bound to it yet")
	handle := func(in string) (out string) {
		return in + "hi"
	}
	plugin.AddHook("haha", handle)
	expect(t, common.GetHookTable().Sshook("haha", "ho") == "hohi", "Sshook didn't give hohi")
	plugin.RemoveHook("haha", handle)
	expect(t, common.GetHookTable().Sshook("haha", "ho") == "ho", "Sshook shouldn't have anything bound to it anymore")

	expect(t, common.GetHookTable().Hook("haha", "ho") == "ho", "Hook shouldn't have anything bound to it yet")
	handle2 := func(inI interface{}) (out interface{}) {
		return inI.(string) + "hi"
	}
	plugin.AddHook("hehe", handle2)
	expect(t, common.GetHookTable().Hook("hehe", "ho").(string) == "hohi", "Hook didn't give hohi")
	plugin.RemoveHook("hehe", handle2)
	expect(t, common.GetHookTable().Hook("hehe", "ho").(string) == "ho", "Hook shouldn't have anything bound to it anymore")

	// TODO: Add tests for more hook types
}

func TestPhrases(t *testing.T) {
	expect(t, phrases.GetGlobalPermPhrase("BanUsers") == "Can ban users", "Not the expected phrase")
	expect(t, phrases.GetGlobalPermPhrase("NoSuchPerm") == "{lang.perms[NoSuchPerm]}", "Not the expected phrase")
	expect(t, phrases.GetLocalPermPhrase("ViewTopic") == "Can view topics", "Not the expected phrase")
	expect(t, phrases.GetLocalPermPhrase("NoSuchPerm") == "{lang.perms[NoSuchPerm]}", "Not the expected phrase")

	// TODO: Cover the other phrase types, also try switching between languages to see if anything strange happens
}

func TestWordFilters(t *testing.T) {
	// TODO: Test the word filters and their store
	expect(t, common.WordFilters.Length() == 0, "Word filter list should be empty")
	expect(t, common.WordFilters.EstCount() == 0, "Word filter list should be empty")
	expect(t, common.WordFilters.GlobalCount() == 0, "Word filter list should be empty")
	filters, err := common.WordFilters.GetAll()
	expectNilErr(t, err) // TODO: Slightly confusing that we don't get ErrNoRow here
	expect(t, len(filters) == 0, "Word filter map should be empty")
	// TODO: Add a test for ParseMessage relating to word filters

	err = common.WordFilters.Create("imbecile", "lovely")
	expectNilErr(t, err)
	expect(t, common.WordFilters.Length() == 1, "Word filter list should not be empty")
	expect(t, common.WordFilters.EstCount() == 1, "Word filter list should not be empty")
	expect(t, common.WordFilters.GlobalCount() == 1, "Word filter list should not be empty")
	filters, err = common.WordFilters.GetAll()
	expectNilErr(t, err)
	expect(t, len(filters) == 1, "Word filter map should not be empty")
	filter := filters[1]
	expect(t, filter.ID == 1, "Word filter ID should be 1")
	expect(t, filter.Find == "imbecile", "Word filter needle should be imbecile")
	expect(t, filter.Replacement == "lovely", "Word filter replacement should be lovely")
	// TODO: Add a test for ParseMessage relating to word filters

	// TODO: Add deletion tests
}

func TestSlugs(t *testing.T) {
	var res string
	var msgList = &MEPairList{nil}
	common.Config.BuildSlugs = true // Flip this switch, otherwise all the tests will fail

	msgList.Add("Unknown", "unknown")
	msgList.Add("Unknown2", "unknown2")
	msgList.Add("Unknown ", "unknown")
	msgList.Add("Unknown 2", "unknown-2")
	msgList.Add("Unknown  2", "unknown-2")
	msgList.Add("Admin Alice", "admin-alice")
	msgList.Add("Admin_Alice", "adminalice")
	msgList.Add("Admin_Alice-", "adminalice")
	msgList.Add("-Admin_Alice-", "adminalice")
	msgList.Add("-Admin@Alice-", "adminalice")
	msgList.Add("-Admin😀Alice-", "adminalice")
	msgList.Add("u", "u")
	msgList.Add("", "untitled")
	msgList.Add(" ", "untitled")
	msgList.Add("-", "untitled")
	msgList.Add("--", "untitled")
	msgList.Add("é", "é")
	msgList.Add("-é-", "é")
	msgList.Add("-你好-", "untitled")
	msgList.Add("-こにちは-", "untitled")

	for _, item := range msgList.Items {
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
	realPassword := "Madame Cassandra's Mystic Orb"
	t.Logf("Set realPassword to '%s'", realPassword)
	t.Log("Hashing the real password with bcrypt")
	hashedPassword, _, err := common.BcryptGeneratePassword(realPassword)
	if err != nil {
		t.Error(err)
	}
	passwordTest(t, realPassword, hashedPassword)
	// TODO: Peek at the prefix to verify this is a bcrypt hash

	t.Log("Hashing the real password")
	hashedPassword2, _, err := common.GeneratePassword(realPassword)
	if err != nil {
		t.Error(err)
	}
	passwordTest(t, realPassword, hashedPassword2)
	// TODO: Peek at the prefix to verify this is a bcrypt hash

	_, err, _ = common.Auth.Authenticate("None", "password")
	errmsg := "Username None shouldn't exist"
	if err != nil {
		errmsg += "\n" + err.Error()
	}
	expect(t, err == common.ErrNoUserByName, errmsg)

	uid, err, _ := common.Auth.Authenticate("Admin", "password")
	expectNilErr(t, err)
	expect(t, uid == 1, fmt.Sprintf("Default admin uid should be 1 not %d", uid))

	_, err, _ = common.Auth.Authenticate("Sam", "ReallyBadPassword")
	errmsg = "Username Sam shouldn't exist"
	if err != nil {
		errmsg += "\n" + err.Error()
	}
	expect(t, err == common.ErrNoUserByName, errmsg)

	admin, err := common.Users.Get(1)
	expectNilErr(t, err)
	// TODO: Move this into the user store tests to provide better coverage? E.g. To see if the installer and the user creator initialise the field differently
	expect(t, admin.Session == "", "Admin session should be blank")

	session, err := common.Auth.CreateSession(1)
	expectNilErr(t, err)
	expect(t, session != "", "Admin session shouldn't be blank")
	// TODO: Test the actual length set in the setting in addition to this "too short" test
	// TODO: We might be able to push up this minimum requirement
	expect(t, len(session) > 10, "Admin session shouldn't be too short")
	expect(t, admin.Session != session, "Old session should not match new one")
	admin, err = common.Users.Get(1)
	expectNilErr(t, err)
	expect(t, admin.Session == session, "Sessions should match")

	// TODO: Create a user with a unicode password and see if we can login as them
	// TODO: Tests for SessionCheck, GetCookies, and ForceLogout
}

// TODO: Vary the salts? Keep in mind that some algorithms store the salt in the hash therefore the salt string may be blank
func passwordTest(t *testing.T, realPassword string, hashedPassword string) {
	if len(hashedPassword) < 10 {
		t.Error("Hash too short")
	}
	salt := ""
	password := realPassword
	t.Logf("Testing password '%s'", password)
	t.Logf("Testing salt '%s'", salt)
	err := common.CheckPassword(hashedPassword, password, salt)
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

type METri struct {
	Name    string // Optional, this is here for tests involving invisible characters so we know what's going in
	Msg     string
	Expects string
}

type METriList struct {
	Items []METri
}

func (tlist *METriList) Add(args ...string) {
	if len(args) < 2 {
		panic("need 2 or more args")
	}
	if len(args) > 2 {
		tlist.Items = append(tlist.Items, METri{args[0], args[1], args[2]})
	} else {
		tlist.Items = append(tlist.Items, METri{"", args[0], args[1]})
	}
}

type CountTest struct {
	Name    string
	Msg     string
	Expects int
}

type CountTestList struct {
	Items []CountTest
}

func (tlist *CountTestList) Add(name string, msg string, expects int) {
	tlist.Items = append(tlist.Items, CountTest{name, msg, expects})
}

func TestWordCount(t *testing.T) {
	var msgList = &CountTestList{nil}

	msgList.Add("blank", "", 0)
	msgList.Add("single-letter", "h", 1)
	msgList.Add("single-kana", "お", 1)
	msgList.Add("single-letter-words", "h h", 2)
	msgList.Add("two-letter", "h", 1)
	msgList.Add("two-kana", "おは", 1)
	msgList.Add("two-letter-words", "hh hh", 2)
	msgList.Add("", "h,h", 2)
	msgList.Add("", "h,,h", 2)
	msgList.Add("", "h, h", 2)
	msgList.Add("", "  h, h", 2)
	msgList.Add("", "h, h  ", 2)
	msgList.Add("", "  h, h  ", 2)
	msgList.Add("", "h,  h", 2)
	msgList.Add("", "h\nh", 2)
	msgList.Add("", "h\"h", 2)
	msgList.Add("", "h[r]h", 3)
	msgList.Add("", "お,お", 2)
	msgList.Add("", "お、お", 2)
	msgList.Add("", "お\nお", 2)
	msgList.Add("", "お”お", 2)
	msgList.Add("", "お「あ」お", 3)

	for _, item := range msgList.Items {
		res := common.WordCount(item.Msg)
		if res != item.Expects {
			if item.Name != "" {
				t.Error("Name: ", item.Name)
			}
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", res)
			t.Error("Expected:", item.Expects)
		}
	}
}

func TestPreparser(t *testing.T) {
	var msgList = &METriList{nil}

	// Note: The open tag is evaluated without knowledge of the close tag for efficiency and simplicity, so the parser autofills the associated close tag when it finds an open tag without a partner
	msgList.Add("", "")
	msgList.Add(" ", "")
	msgList.Add(" hi", "hi")
	msgList.Add("hi ", "hi")
	msgList.Add("hi", "hi")
	msgList.Add(":grinning:", "😀")
	msgList.Add("😀", "😀")
	msgList.Add("&nbsp;", "")
	msgList.Add("<p>", "")
	msgList.Add("</p>", "")
	msgList.Add("<p></p>", "")

	msgList.Add("<", "&lt;")
	msgList.Add(">", "&gt;")
	msgList.Add("<meow>", "&lt;meow&gt;")
	msgList.Add("&lt;", "&amp;lt;")
	msgList.Add("&", "&amp;")

	// Note: strings.TrimSpace strips newlines, if there's nothing before or after them
	msgList.Add("<br>", "")
	msgList.Add("<br />", "")
	msgList.Add("\\n", "\n", "")
	msgList.Add("\\n\\n", "\n\n", "")
	msgList.Add("\\n\\n\\n", "\n\n\n", "")
	msgList.Add("\\r\\n", "\r\n", "") // Windows style line ending
	msgList.Add("\\n\\r", "\n\r", "")

	msgList.Add("ho<br>ho", "ho\n\nho")
	msgList.Add("ho<br />ho", "ho\n\nho")
	msgList.Add("ho\\nho", "ho\nho", "ho\nho")
	msgList.Add("ho\\n\\nho", "ho\n\nho", "ho\n\nho")
	//msgList.Add("ho\\n\\n\\n\\nho", "ho\n\n\n\nho", "ho\n\n\nho")
	msgList.Add("ho\\r\\nho", "ho\r\nho", "ho\nho") // Windows style line ending
	msgList.Add("ho\\n\\rho", "ho\n\rho", "ho\nho")

	msgList.Add("<b></b>", "<strong></strong>")
	msgList.Add("<b>hi</b>", "<strong>hi</strong>")
	msgList.Add("<s>hi</s>", "<del>hi</del>")
	msgList.Add("<del>hi</del>", "<del>hi</del>")
	msgList.Add("<u>hi</u>", "<u>hi</u>")
	msgList.Add("<em>hi</em>", "<em>hi</em>")
	msgList.Add("<i>hi</i>", "<em>hi</em>")
	msgList.Add("<strong>hi</strong>", "<strong>hi</strong>")
	msgList.Add("<b><i>hi</i></b>", "<strong><em>hi</em></strong>")
	msgList.Add("<strong><em>hi</em></strong>", "<strong><em>hi</em></strong>")
	msgList.Add("<b><i><b>hi</b></i></b>", "<strong><em><strong>hi</strong></em></strong>")
	msgList.Add("<strong><em><strong>hi</strong></em></strong>", "<strong><em><strong>hi</strong></em></strong>")
	msgList.Add("<div>hi</div>", "&lt;div&gt;hi&lt;/div&gt;")
	msgList.Add("<span>hi</span>", "hi") // This is stripped since the editor (Trumbowyg) likes blasting useless spans
	msgList.Add("<span   >hi</span>", "hi")
	msgList.Add("<span style='background-color: yellow;'>hi</span>", "hi")
	msgList.Add("<span style='background-color: yellow;'>>hi</span>", "&gt;hi")
	msgList.Add("<b>hi", "<strong>hi</strong>")
	msgList.Add("hi</b>", "hi&lt;/b&gt;")
	msgList.Add("</b>", "&lt;/b&gt;")
	msgList.Add("</del>", "&lt;/del&gt;")
	msgList.Add("</strong>", "&lt;/strong&gt;")
	msgList.Add("<b>", "<strong></strong>")
	msgList.Add("<span style='background-color: yellow;'>hi", "hi")
	msgList.Add("hi</span>", "hi")
	msgList.Add("</span>", "")
	msgList.Add("<span></span>", "")
	msgList.Add("<span   ></span>", "")
	msgList.Add("<></>", "&lt;&gt;&lt;/&gt;")
	msgList.Add("</><>", "&lt;/&gt;&lt;&gt;")
	msgList.Add("<>", "&lt;&gt;")
	msgList.Add("</>", "&lt;/&gt;")
	msgList.Add("@", "@")
	msgList.Add("@Admin", "@1")
	msgList.Add("@Bah", "@Bah")
	msgList.Add(" @Admin", "@1")
	msgList.Add("\n@Admin", "@1")
	msgList.Add("@Admin\n", "@1")
	msgList.Add("@Admin\ndd", "@1\ndd")
	msgList.Add("d@Admin", "d@Admin")
	//msgList.Add("byte 0", string([]byte{0}), "")
	msgList.Add("byte 'a'", string([]byte{'a'}), "a")
	//msgList.Add("byte 255", string([]byte{255}), "")
	//msgList.Add("rune 0", string([]rune{0}), "")
	// TODO: Do a test with invalid UTF-8 input

	for _, item := range msgList.Items {
		res := common.PreparseMessage(item.Msg)
		if res != item.Expects {
			if item.Name != "" {
				t.Error("Name: ", item.Name)
			}
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			//t.Error("Ouput in bytes:", []byte(res))
			t.Error("Expected:", "'"+item.Expects+"'")
		}
	}
}

func TestParser(t *testing.T) {
	var msgList = &METriList{nil}

	msgList.Add("//github.com/Azareal/Gosora", "<a href='//github.com/Azareal/Gosora'>//github.com/Azareal/Gosora</a>")
	msgList.Add("https://github.com/Azareal/Gosora", "<a href='https://github.com/Azareal/Gosora'>https://github.com/Azareal/Gosora</a>")
	msgList.Add("http://github.com/Azareal/Gosora", "<a href='http://github.com/Azareal/Gosora'>http://github.com/Azareal/Gosora</a>")
	msgList.Add("//github.com/Azareal/Gosora\n", "<a href='//github.com/Azareal/Gosora'>//github.com/Azareal/Gosora</a><br>")
	msgList.Add("\n//github.com/Azareal/Gosora", "<br><a href='//github.com/Azareal/Gosora'>//github.com/Azareal/Gosora</a>")
	msgList.Add("\n//github.com/Azareal/Gosora\n", "<br><a href='//github.com/Azareal/Gosora'>//github.com/Azareal/Gosora</a><br>")
	msgList.Add("//github.com/Azareal/Gosora\n//github.com/Azareal/Gosora", "<a href='//github.com/Azareal/Gosora'>//github.com/Azareal/Gosora</a><br><a href='//github.com/Azareal/Gosora'>//github.com/Azareal/Gosora</a>")
	msgList.Add("//github.com/Azareal/Gosora\n\n//github.com/Azareal/Gosora", "<a href='//github.com/Azareal/Gosora'>//github.com/Azareal/Gosora</a><br><br><a href='//github.com/Azareal/Gosora'>//github.com/Azareal/Gosora</a>")
	msgList.Add("//"+common.Site.URL, "<a href='//"+common.Site.URL+"'>//"+common.Site.URL+"</a>")
	msgList.Add("//"+common.Site.URL+"\n", "<a href='//"+common.Site.URL+"'>//"+common.Site.URL+"</a><br>")
	msgList.Add("//"+common.Site.URL+"\n//"+common.Site.URL, "<a href='//"+common.Site.URL+"'>//"+common.Site.URL+"</a><br><a href='//"+common.Site.URL+"'>//"+common.Site.URL+"</a>")

	msgList.Add("#tid-1", "<a href='/topic/1'>#tid-1</a>")
	msgList.Add("https://github.com/Azareal/Gosora/#tid-1", "<a href='https://github.com/Azareal/Gosora/#tid-1'>https://github.com/Azareal/Gosora/#tid-1</a>")
	msgList.Add("#fid-1", "<a href='/forum/1'>#fid-1</a>")
	msgList.Add("@1", "<a href='/user/admin.1' class='mention'>@Admin</a>")
	msgList.Add("@0", "<span style='color: red;'>[Invalid Profile]</span>")
	msgList.Add("@-1", "<span style='color: red;'>[Invalid Profile]</span>1")

	for _, item := range msgList.Items {
		res := common.ParseMessage(item.Msg, 1, "forums")
		if res != item.Expects {
			if item.Name != "" {
				t.Error("Name: ", item.Name)
			}
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", "'"+item.Expects+"'")
		}
	}
}
