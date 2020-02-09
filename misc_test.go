package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func miscinit(t *testing.T) {
	if err := gloinit(); err != nil {
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
	if !c.PluginsInited {
		c.InitPlugins()
	}

	var err error
	ucache := c.NewMemoryUserCache(c.Config.UserCacheCapacity)
	c.Users, err = c.NewDefaultUserStore(ucache)
	expectNilErr(t, err)
	ucache.Flush()
	userStoreTest(t, 2)
	c.Users, err = c.NewDefaultUserStore(nil)
	expectNilErr(t, err)
	userStoreTest(t, 5)
}
func userStoreTest(t *testing.T, newUserID int) {
	ucache := c.Users.GetCache()
	// Go doesn't have short-circuiting, so this'll allow us to do one liner tests
	isCacheLengthZero := func(uc c.UserCache) bool {
		if uc == nil {
			return true
		}
		return uc.Length() == 0
	}
	cacheLength := func(uc c.UserCache) int {
		if uc == nil {
			return 0
		}
		return uc.Length()
	}
	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("The initial ucache length should be zero, not %d", cacheLength(ucache)))

	_, err := c.Users.Get(-1)
	recordMustNotExist(t, err, "UID #-1 shouldn't exist")
	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("We found %d items in the user cache and it's supposed to be empty", cacheLength(ucache)))

	_, err = c.Users.Get(0)
	recordMustNotExist(t, err, "UID #0 shouldn't exist")
	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("We found %d items in the user cache and it's supposed to be empty", cacheLength(ucache)))

	user, err := c.Users.Get(1)
	recordMustExist(t, err, "Couldn't find UID #1")

	expectW := func(cond, expec bool, prefix, suffix string) {
		midfix := "should not be"
		if expec {
			midfix = "should be"
		}
		expect(t, cond, prefix+" "+midfix+" "+suffix)
	}

	// TODO: Add email checks too? Do them separately?
	expectUser := func(u *c.User, uid int, name string, group int, super, admin, mod, banned bool) {
		expect(t, u.ID == uid, fmt.Sprintf("u.ID should be %d. Got '%d' instead.", uid, u.ID))
		expect(t, u.Name == name, fmt.Sprintf("u.Name should be '%s', not '%s'", name, u.Name))
		expectW(u.Group == group, true, u.Name, "in group"+strconv.Itoa(group))
		expectW(u.IsSuperAdmin == super, super, u.Name, "a super admin")
		expectW(u.IsAdmin == admin, admin, u.Name, "an admin")
		expectW(u.IsSuperMod == mod, mod, u.Name, "a super mod")
		expectW(u.IsMod == mod, mod, u.Name, "a mod")
		expectW(u.IsBanned == banned, banned, u.Name, "banned")
	}
	expectUser(user, 1, "Admin", 1, true, true, true, false)

	_, err = c.Users.Get(newUserID)
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
	var userList map[int]*c.User
	userList, _ = c.Users.BulkGetMap([]int{-1})
	expect(t, len(userList) == 0, fmt.Sprintf("The userList length should be 0, not %d", len(userList)))
	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("User cache length should be 0, not %d", cacheLength(ucache)))

	userList, _ = c.Users.BulkGetMap([]int{0})
	expect(t, len(userList) == 0, fmt.Sprintf("The userList length should be 0, not %d", len(userList)))
	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("User cache length should be 0, not %d", cacheLength(ucache)))

	userList, _ = c.Users.BulkGetMap([]int{1})
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

	expect(t, !c.Users.Exists(-1), "UID #-1 shouldn't exist")
	expect(t, !c.Users.Exists(0), "UID #0 shouldn't exist")
	expect(t, c.Users.Exists(1), "UID #1 should exist")
	expect(t, !c.Users.Exists(newUserID), fmt.Sprintf("UID #%d shouldn't exist", newUserID))

	expect(t, isCacheLengthZero(ucache), fmt.Sprintf("User cache length should be 0, not %d", cacheLength(ucache)))
	expectIntToBeX(t, c.Users.Count(), 1, "The number of users should be one, not %d")

	awaitingActivation := 5
	// TODO: Write tests for the registration validators
	uid, err := c.Users.Create("Sam", "ReallyBadPassword", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expect(t, uid == newUserID, fmt.Sprintf("The UID of the new user should be %d not %d", newUserID, uid))
	expect(t, c.Users.Exists(newUserID), fmt.Sprintf("UID #%d should exist", newUserID))

	user, err = c.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", 5, false, false, false, false)

	if ucache != nil {
		expectIntToBeX(t, ucache.Length(), 1, "User cache length should be 1, not %d")
		user, err = ucache.Get(newUserID)
		recordMustExist(t, err, "Couldn't find UID #%d in the cache", newUserID)
		expect(t, user.ID == newUserID, fmt.Sprintf("user.ID does not match the requested UID. Got '%d' instead.", user.ID))
	}

	userList, _ = c.Users.BulkGetMap([]int{1, uid})
	expect(t, len(userList) == 2, fmt.Sprintf("Returned map should have two results, not %d", len(userList)))

	if ucache != nil {
		expectIntToBeX(t, ucache.Length(), 2, "User cache length should be 2, not %d")
		user, err = ucache.Get(1)
		recordMustExist(t, err, "Couldn't find UID #%d in the cache", 1)
		expect(t, user.ID == 1, fmt.Sprintf("user.ID does not match the requested UID. Got '%d' instead.", user.ID))
		user, err = ucache.Get(newUserID)
		recordMustExist(t, err, "Couldn't find UID #%d in the cache", newUserID)
		expect(t, user.ID == newUserID, fmt.Sprintf("user.ID does not match the requested UID. Got '%d' instead.", user.ID))
		ucache.Flush()
	}

	user, err = c.Users.Get(newUserID)
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
	afterUserFlush := func(uid int) {
		if ucache != nil {
			expectIntToBeX(t, ucache.Length(), 0, "User cache length should be 0, not %d")
			_, err = ucache.Get(uid)
			recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", uid)
		}
	}
	afterUserFlush(newUserID)

	user, err = c.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", c.Config.DefaultGroup, false, false, false, false)

	// Permanent ban
	duration, _ := time.ParseDuration("0")

	// TODO: Attempt a double ban, double activation, and double unban
	err = user.Ban(duration, 1)
	expectNilErr(t, err)
	expect(t, user.Group == c.Config.DefaultGroup, fmt.Sprintf("Sam should be in group %d, not %d", c.Config.DefaultGroup, user.Group))
	afterUserFlush(newUserID)

	user, err = c.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", c.BanGroup, false, false, false, true)

	// TODO: Do tests against the scheduled updates table and the task system to make sure the ban exists there and gets revoked when it should

	err = user.Unban()
	expectNilErr(t, err)
	expectIntToBeX(t, user.Group, c.BanGroup, "Sam should still be in the ban group in this copy")
	afterUserFlush(newUserID)

	user, err = c.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", c.Config.DefaultGroup, false, false, false, false)

	reportsForumID := 1 // TODO: Use the constant in common?
	generalForumID := 2
	dummyResponseRecorder := httptest.NewRecorder()
	bytesBuffer := bytes.NewBuffer([]byte(""))
	dummyRequest1 := httptest.NewRequest("", "/forum/"+strconv.Itoa(reportsForumID), bytesBuffer)
	dummyRequest2 := httptest.NewRequest("", "/forum/"+strconv.Itoa(generalForumID), bytesBuffer)
	var user2 *c.User

	changeGroupTest := func(oldGroup, newGroup int) {
		err = user.ChangeGroup(newGroup)
		expectNilErr(t, err)
		// ! I don't think ChangeGroup should be changing the value of user... Investigate this.
		expect(t, oldGroup == user.Group, "Someone's mutated this pointer elsewhere")

		user, err = c.Users.Get(newUserID)
		recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
		user2 = c.BlankUser()
		*user2 = *user
	}

	changeGroupTest2 := func(rank string, firstShouldBe, secondShouldBe bool) {
		head, err := c.UserCheck(dummyResponseRecorder, dummyRequest1, user)
		if err != nil {
			t.Fatal(err)
		}
		head2, err := c.UserCheck(dummyResponseRecorder, dummyRequest2, user2)
		if err != nil {
			t.Fatal(err)
		}
		ferr := c.ForumUserCheck(head, dummyResponseRecorder, dummyRequest1, user, reportsForumID)
		expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
		expect(t, user.Perms.ViewTopic == firstShouldBe, rank+" should be able to access the reports forum")
		ferr = c.ForumUserCheck(head2, dummyResponseRecorder, dummyRequest2, user2, generalForumID)
		expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
		expect(t, user2.Perms.ViewTopic == secondShouldBe, "Sam should be able to access the general forum")
	}

	changeGroupTest(c.Config.DefaultGroup, 1)
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

	err = user.ChangeGroup(c.Config.DefaultGroup)
	expectNilErr(t, err)
	expect(t, user.Group == 6, "Someone's mutated this pointer elsewhere")

	err = user.Delete()
	expectNilErr(t, err)
	expect(t, !c.Users.Exists(newUserID), fmt.Sprintf("UID #%d should no longer exist", newUserID))
	afterUserFlush(newUserID)

	_, err = c.Users.Get(newUserID)
	recordMustNotExist(t, err, "UID #%d shouldn't exist", newUserID)

	// And a unicode test, even though I doubt it'll fail
	uid, err = c.Users.Create("サム", "😀😀😀", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expect(t, uid == newUserID+1, fmt.Sprintf("The UID of the new user should be %d", newUserID+1))
	expect(t, c.Users.Exists(newUserID+1), fmt.Sprintf("UID #%d should exist", newUserID+1))

	user, err = c.Users.Get(newUserID + 1)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID+1)
	expectUser(user, newUserID+1, "サム", 5, false, false, false, false)

	err = user.Delete()
	expectNilErr(t, err)
	expect(t, !c.Users.Exists(newUserID+1), fmt.Sprintf("UID #%d should no longer exist", newUserID+1))

	// MySQL utf8mb4 username test
	uid, err = c.Users.Create("😀😀😀", "😀😀😀", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expect(t, uid == newUserID+2, fmt.Sprintf("The UID of the new user should be %d", newUserID+2))
	expect(t, c.Users.Exists(newUserID+2), fmt.Sprintf("UID #%d should exist", newUserID+2))

	user, err = c.Users.Get(newUserID + 2)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID+1)
	expectUser(user, newUserID+2, "😀😀😀", 5, false, false, false, false)

	err = user.Delete()
	expectNilErr(t, err)
	expect(t, !c.Users.Exists(newUserID+2), fmt.Sprintf("UID #%d should no longer exist", newUserID+2))

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

func expectIntToBeX(t *testing.T, item, expect int, errmsg string) {
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
	if !c.PluginsInited {
		c.InitPlugins()
	}

	dummyResponseRecorder := httptest.NewRecorder()
	bytesBuffer := bytes.NewBuffer([]byte(""))
	dummyRequest := httptest.NewRequest("", "/forum/1", bytesBuffer)
	user := c.BlankUser()

	ferr := c.SuperModOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Blank users shouldn't be supermods")

	user.IsSuperMod = false
	ferr = c.SuperModOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Non-supermods shouldn't be allowed through supermod gates")

	user.IsSuperMod = true
	ferr = c.SuperModOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr == nil, "Supermods should be allowed through supermod gates")

	// TODO: Loop over the Control Panel routes and make sure only supermods can get in

	user = c.BlankUser()

	ferr = c.MemberOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Blank users shouldn't be considered loggedin")

	user.Loggedin = false
	ferr = c.MemberOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Guests shouldn't be able to access member areas")

	user.Loggedin = true
	ferr = c.MemberOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr == nil, "Logged in users should be able to access member areas")

	// TODO: Loop over the /user/ routes and make sure only members can access the ones other than /user/username

	// TODO: Write tests for AdminOnly()

	user = c.BlankUser()

	ferr = c.SuperAdminOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Blank users shouldn't be considered super admins")

	user.IsSuperAdmin = false
	ferr = c.SuperAdminOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr != nil, "Non-super admins shouldn't be allowed through the super admin gate")

	user.IsSuperAdmin = true
	ferr = c.SuperAdminOnly(dummyResponseRecorder, dummyRequest, *user)
	expect(t, ferr == nil, "Super admins should be allowed through super admin gates")

	// TODO: Make sure only super admins can access the backups route

	//dummyResponseRecorder = httptest.NewRecorder()
	//bytesBuffer = bytes.NewBuffer([]byte(""))
	//dummyRequest = httptest.NewRequest("", "/panel/backups/", bytesBuffer)
}

func TestTopicStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	var err error
	tcache := c.NewMemoryTopicCache(c.Config.TopicCacheCapacity)
	c.Topics, err = c.NewDefaultTopicStore(tcache)
	expectNilErr(t, err)
	c.Config.DisablePostIP = false
	topicStoreTest(t, 2, "::1")
	c.Config.DisablePostIP = true
	topicStoreTest(t, 3, "")

	c.Topics, err = c.NewDefaultTopicStore(nil)
	expectNilErr(t, err)
	c.Config.DisablePostIP = false
	topicStoreTest(t, 4, "::1")
	c.Config.DisablePostIP = true
	topicStoreTest(t, 5, "")
}
func topicStoreTest(t *testing.T, newID int, ip string) {
	var topic *c.Topic
	var err error

	_, err = c.Topics.Get(-1)
	recordMustNotExist(t, err, "TID #-1 shouldn't exist")
	_, err = c.Topics.Get(0)
	recordMustNotExist(t, err, "TID #0 shouldn't exist")

	topic, err = c.Topics.Get(1)
	recordMustExist(t, err, "Couldn't find TID #1")
	expect(t, topic.ID == 1, fmt.Sprintf("topic.ID does not match the requested TID. Got '%d' instead.", topic.ID))

	// TODO: Add BulkGetMap() to the TopicStore

	expect(t, !c.Topics.Exists(-1), "TID #-1 shouldn't exist")
	expect(t, !c.Topics.Exists(0), "TID #0 shouldn't exist")
	expect(t, c.Topics.Exists(1), "TID #1 should exist")

	count := c.Topics.Count()
	expect(t, count == 1, fmt.Sprintf("Global count for topics should be 1, not %d", count))

	//Create(fid int, topicName string, content string, uid int, ip string) (tid int, err error)
	tid, err := c.Topics.Create(2, "Test Topic", "Topic Content", 1, ip)
	expectNilErr(t, err)
	expect(t, tid == newID, fmt.Sprintf("TID for the new topic should be %d, not %d", newID, tid))
	expect(t, c.Topics.Exists(newID), fmt.Sprintf("TID #%d should exist", newID))

	count = c.Topics.Count()
	expect(t, count == 2, fmt.Sprintf("Global count for topics should be 2, not %d", count))

	iFrag := func(cond bool) string {
		if !cond {
			return "n't"
		}
		return ""
	}

	testTopic := func(tid int, title, content string, createdBy int, ip string, parentID int, isClosed, sticky bool) {
		topic, err = c.Topics.Get(tid)
		recordMustExist(t, err, fmt.Sprintf("Couldn't find TID #%d", tid))
		expect(t, topic.ID == tid, fmt.Sprintf("topic.ID does not match the requested TID. Got '%d' instead.", topic.ID))
		expect(t, topic.GetID() == tid, fmt.Sprintf("topic.ID does not match the requested TID. Got '%d' instead.", topic.GetID()))
		expect(t, topic.Title == title, fmt.Sprintf("The topic's name should be '%s', not %s", title, topic.Title))
		expect(t, topic.Content == content, fmt.Sprintf("The topic's body should be '%s', not %s", content, topic.Content))
		expect(t, topic.CreatedBy == createdBy, fmt.Sprintf("The topic's creator should be %d, not %d", createdBy, topic.CreatedBy))
		expect(t, topic.IP == ip, fmt.Sprintf("The topic's IP should be '%s', not %s", ip, topic.IP))
		expect(t, topic.ParentID == parentID, fmt.Sprintf("The topic's parent forum should be %d, not %d", parentID, topic.ParentID))
		expect(t, topic.IsClosed == isClosed, fmt.Sprintf("This topic should%s be locked", iFrag(topic.IsClosed)))
		expect(t, topic.Sticky == sticky, fmt.Sprintf("This topic should%s be sticky", iFrag(topic.Sticky)))
		expect(t, topic.GetTable() == "topics", fmt.Sprintf("The topic's table should be 'topics', not %s", topic.GetTable()))
	}

	tcache := c.Topics.GetCache()
	shouldNotBeIn := func(tid int) {
		if tcache != nil {
			_, err = tcache.Get(tid)
			recordMustNotExist(t, err, "Topic cache should be empty")
		}
	}
	if tcache != nil {
		_, err = tcache.Get(newID)
		expectNilErr(t, err)
	}

	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 2, false, false)

	expectNilErr(t, topic.Lock())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 2, true, false)

	expectNilErr(t, topic.Unlock())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 2, false, false)

	expectNilErr(t, topic.Stick())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 2, false, true)

	expectNilErr(t, topic.Unstick())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 2, false, false)

	expectNilErr(t, topic.MoveTo(1))
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 1, false, false)
	// TODO: Add more tests for more *Topic methods

	expectNilErr(t, topic.Delete())
	shouldNotBeIn(newID)

	_, err = c.Topics.Get(newID)
	recordMustNotExist(t, err, fmt.Sprintf("TID #%d shouldn't exist", newID))
	expect(t, !c.Topics.Exists(newID), fmt.Sprintf("TID #%d shouldn't exist", newID))

	// TODO: Test topic creation and retrieving that created topic plus reload and inspecting the cache
}

func TestForumStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}
	// TODO: Test ForumStore.Reload

	fcache, ok := c.Forums.(c.ForumCache)
	expect(t, ok, "Unable to cast ForumStore to ForumCache")

	expect(t, c.Forums.Count() == 2, "The forumstore global count should be 2")
	expect(t, fcache.Length() == 2, "The forum cache length should be 2")

	_, err := c.Forums.Get(-1)
	recordMustNotExist(t, err, "FID #-1 shouldn't exist")
	_, err = c.Forums.Get(0)
	recordMustNotExist(t, err, "FID #0 shouldn't exist")

	forum, err := c.Forums.Get(1)
	recordMustExist(t, err, "Couldn't find FID #1")
	expect(t, forum.ID == 1, fmt.Sprintf("forum.ID doesn't not match the requested FID. Got '%d' instead.'", forum.ID))
	// TODO: Check the preset and forum permissions
	expect(t, forum.Name == "Reports", fmt.Sprintf("FID #0 is named '%s' and not 'Reports'", forum.Name))
	expect(t, !forum.Active, fmt.Sprintf("The reports forum shouldn't be active"))
	expectDesc := "All the reports go here"
	expect(t, forum.Desc == expectDesc, fmt.Sprintf("The forum description should be '%s' not '%s'", expectDesc, forum.Desc))
	forum, err = c.Forums.BypassGet(1)
	recordMustExist(t, err, "Couldn't find FID #1")

	forum, err = c.Forums.Get(2)
	recordMustExist(t, err, "Couldn't find FID #2")
	forum, err = c.Forums.BypassGet(2)
	recordMustExist(t, err, "Couldn't find FID #2")

	expect(t, forum.ID == 2, fmt.Sprintf("The FID should be 2 not %d", forum.ID))
	expect(t, forum.Name == "General", fmt.Sprintf("The name of the forum should be 'General' not '%s'", forum.Name))
	expect(t, forum.Active, fmt.Sprintf("The general forum should be active"))
	expectDesc = "A place for general discussions which don't fit elsewhere"
	expect(t, forum.Desc == expectDesc, fmt.Sprintf("The forum description should be '%s' not '%s'", expectDesc, forum.Desc))

	// Forum reload test, kind of hacky but gets the job done
	/*
		CacheGet(id int) (*Forum, error)
		CacheSet(forum *Forum) error
	*/
	expect(t, ok, "ForumCache should be available")
	forum.Name = "nanana"
	fcache.CacheSet(forum)
	forum, err = c.Forums.Get(2)
	recordMustExist(t, err, "Couldn't find FID #2")
	expect(t, forum.Name == "nanana", fmt.Sprintf("The faux name should be nanana not %s", forum.Name))
	expectNilErr(t, c.Forums.Reload(2))
	forum, err = c.Forums.Get(2)
	recordMustExist(t, err, "Couldn't find FID #2")
	expect(t, forum.Name == "General", fmt.Sprintf("The proper name should be 2 not %s", forum.Name))

	expect(t, !c.Forums.Exists(-1), "FID #-1 shouldn't exist")
	expect(t, !c.Forums.Exists(0), "FID #0 shouldn't exist")
	expect(t, c.Forums.Exists(1), "FID #1 should exist")
	expect(t, c.Forums.Exists(2), "FID #2 should exist")
	expect(t, !c.Forums.Exists(3), "FID #3 shouldn't exist")

	_, err = c.Forums.Create("", "", true, "all")
	expect(t, err != nil, "A forum shouldn't be successfully created, if it has a blank name")

	fid, err := c.Forums.Create("Test Forum", "", true, "all")
	expectNilErr(t, err)
	expect(t, fid == 3, "The first forum we create should have an ID of 3")
	expect(t, c.Forums.Exists(3), "FID #2 should exist")

	expect(t, c.Forums.Count() == 3, "The forumstore global count should be 3")
	expect(t, fcache.Length() == 3, "The forum cache length should be 3")

	forum, err = c.Forums.Get(3)
	recordMustExist(t, err, "Couldn't find FID #3")
	forum, err = c.Forums.BypassGet(3)
	recordMustExist(t, err, "Couldn't find FID #3")

	expect(t, forum.ID == 3, fmt.Sprintf("The FID should be 3 not %d", forum.ID))
	expect(t, forum.Name == "Test Forum", fmt.Sprintf("The name of the forum should be 'Test Forum' not '%s'", forum.Name))
	expect(t, forum.Active, fmt.Sprintf("The test forum should be active"))
	expect(t, forum.Desc == "", fmt.Sprintf("The forum description should be blank not '%s'", forum.Desc))

	// TODO: More forum creation tests

	expectNilErr(t, c.Forums.Delete(3))
	expect(t, forum.ID == 3, fmt.Sprintf("forum pointer shenanigans"))
	expect(t, c.Forums.Count() == 2, "The forumstore global count should be 2")
	expect(t, fcache.Length() == 2, "The forum cache length should be 2")
	expect(t, !c.Forums.Exists(3), "FID #3 shouldn't exist after being deleted")
	_, err = c.Forums.Get(3)
	recordMustNotExist(t, err, "FID #3 shouldn't exist after being deleted")
	_, err = c.Forums.BypassGet(3)
	recordMustNotExist(t, err, "FID #3 shouldn't exist after being deleted")

	expect(t, c.Forums.Delete(c.ReportForumID) != nil, "The reports forum shouldn't be deletable")
	expect(t, c.Forums.Exists(c.ReportForumID), fmt.Sprintf("FID #%d should still exist", c.ReportForumID))
	_, err = c.Forums.Get(c.ReportForumID)
	expect(t, err == nil, fmt.Sprintf("FID #%d should still exist", c.ReportForumID))
	_, err = c.Forums.BypassGet(c.ReportForumID)
	expect(t, err == nil, fmt.Sprintf("FID #%d should still exist", c.ReportForumID))

	eforums := map[int]bool{1: true, 2: true}
	{
		forums, err := c.Forums.GetAll()
		expectNilErr(t, err)
		found := make(map[int]*c.Forum)
		for _, forum := range forums {
			_, ok := eforums[forum.ID]
			expect(t, ok, fmt.Sprintf("unknown forum #%d in forums", forum.ID))
			found[forum.ID] = forum
		}
		for fid, _ := range eforums {
			_, ok := found[fid]
			expect(t, ok, fmt.Sprintf("unable to find expected forum #%d in forums", fid))
		}
	}

	{
		fids, err := c.Forums.GetAllIDs()
		expectNilErr(t, err)
		found := make(map[int]bool)
		for _, fid := range fids {
			_, ok := eforums[fid]
			expect(t, ok, fmt.Sprintf("unknown fid #%d in fids", fid))
			found[fid] = true
		}
		for fid, _ := range eforums {
			_, ok := found[fid]
			expect(t, ok, fmt.Sprintf("unable to find expected fid #%d in fids", fid))
		}
	}

	vforums := map[int]bool{2: true}
	{
		forums, err := c.Forums.GetAllVisible()
		expectNilErr(t, err)
		found := make(map[int]*c.Forum)
		for _, forum := range forums {
			_, ok := vforums[forum.ID]
			expect(t, ok, fmt.Sprintf("unknown forum #%d in forums", forum.ID))
			found[forum.ID] = forum
		}
		for fid, _ := range vforums {
			_, ok := found[fid]
			expect(t, ok, fmt.Sprintf("unable to find expected forum #%d in forums", fid))
		}
	}

	{
		fids, err := c.Forums.GetAllVisibleIDs()
		expectNilErr(t, err)
		found := make(map[int]bool)
		for _, fid := range fids {
			_, ok := vforums[fid]
			expect(t, ok, fmt.Sprintf("unknown fid #%d in fids", fid))
			found[fid] = true
		}
		for fid, _ := range vforums {
			_, ok := found[fid]
			expect(t, ok, fmt.Sprintf("unable to find expected fid #%d in fids", fid))
		}
	}

	// TODO: Test forum update
	// TODO: Other forumstore stuff and forumcache?
}

// TODO: Implement this
func TestForumPermsStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	f := func(fid, gid int, msg string, inv ...bool) {
		fp, err := c.FPStore.Get(fid, gid)
		expectNilErr(t, err)
		vt := fp.ViewTopic
		if len(inv) > 0 && inv[0] == true {
			vt = !vt
		}
		expect(t, vt, msg)
	}

	initialState := func() {
		f(1, 1, "admins should be able to see reports")
		f(1, 2, "mods should be able to see reports")
		f(1, 3, "members should not be able to see reports", true)
		f(1, 4, "banned users should not be able to see reports", true)
		f(2, 1, "admins should be able to see general")
		f(2, 3, "members should be able to see general")
		f(2, 6, "guests should be able to see general")
	}
	initialState()

	expectNilErr(t, c.FPStore.Reload(1))
	initialState()
	expectNilErr(t, c.FPStore.Reload(2))
	initialState()
}

// TODO: Test the group permissions
// TODO: Test group.CanSee for forum presets + group perms
func TestGroupStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	_, err := c.Groups.Get(-1)
	recordMustNotExist(t, err, "GID #-1 shouldn't exist")

	// TODO: Refactor the group store to remove GID #0
	group, err := c.Groups.Get(0)
	recordMustExist(t, err, "Couldn't find GID #0")

	expect(t, group.ID == 0, fmt.Sprintf("group.ID doesn't not match the requested GID. Got '%d' instead.", group.ID))
	expect(t, group.Name == "Unknown", fmt.Sprintf("GID #0 is named '%s' and not 'Unknown'", group.Name))

	group, err = c.Groups.Get(1)
	recordMustExist(t, err, "Couldn't find GID #1")
	expect(t, group.ID == 1, fmt.Sprintf("group.ID doesn't not match the requested GID. Got '%d' instead.'", group.ID))
	expect(t, len(group.CanSee) > 0, "group.CanSee should not be zero")

	expect(t, !c.Groups.Exists(-1), "GID #-1 shouldn't exist")
	// 0 aka Unknown, for system posts and other oddities
	expect(t, c.Groups.Exists(0), "GID #0 should exist")
	expect(t, c.Groups.Exists(1), "GID #1 should exist")

	isAdmin := true
	isMod := true
	isBanned := false
	gid, err := c.Groups.Create("Testing", "Test", isAdmin, isMod, isBanned)
	expectNilErr(t, err)
	expect(t, c.Groups.Exists(gid), "The group we just made doesn't exist")

	group, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, group.IsAdmin, "This should be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")
	expect(t, len(group.CanSee) == 0, "group.CanSee should be empty")

	isAdmin = false
	isMod = true
	isBanned = true
	gid, err = c.Groups.Create("Testing 2", "Test", isAdmin, isMod, isBanned)
	expectNilErr(t, err)
	expect(t, c.Groups.Exists(gid), "The group we just made doesn't exist")

	group, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This should not be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	// TODO: Make sure this pointer doesn't change once we refactor the group store to stop updating the pointer
	err = group.ChangeRank(false, false, true)
	expectNilErr(t, err)

	group, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This shouldn't be an admin group")
	expect(t, !group.IsMod, "This shouldn't be a mod group")
	expect(t, group.IsBanned, "This should be a ban group")

	err = group.ChangeRank(true, true, true)
	expectNilErr(t, err)

	group, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, group.IsAdmin, "This should be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")
	expect(t, len(group.CanSee) == 0, "len(group.CanSee) should be 0")

	err = group.ChangeRank(false, true, true)
	expectNilErr(t, err)

	forum, err := c.Forums.Get(2)
	expectNilErr(t, err)
	forumPerms, err := c.FPStore.GetCopy(2, gid)
	if err == sql.ErrNoRows {
		forumPerms = *c.BlankForumPerms()
	} else if err != nil {
		expectNilErr(t, err)
	}
	forumPerms.ViewTopic = true

	err = forum.SetPerms(&forumPerms, "custom", gid)
	expectNilErr(t, err)

	group, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This shouldn't be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")
	expect(t, group.CanSee != nil, "group.CanSee must not be nil")
	expect(t, len(group.CanSee) == 1, "len(group.CanSee) should not be one")
	expect(t, group.CanSee[0] == 2, "group.CanSee[0] should be 2")
	canSee := group.CanSee

	// Make sure the data is static
	c.Groups.Reload(gid)

	group, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, group.ID == gid, "The group ID should match the requested ID")
	expect(t, !group.IsAdmin, "This shouldn't be an admin group")
	expect(t, group.IsMod, "This should be a mod group")
	expect(t, !group.IsBanned, "This shouldn't be a ban group")

	// TODO: Don't enforce a specific order here
	canSeeTest := func(a, b []int) bool {
		if (a == nil) != (b == nil) {
			return false
		}
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}

	expect(t, canSeeTest(group.CanSee, canSee), "group.CanSee is not being reused")

	// TODO: Test group deletion
	// TODO: Test group reload
	// TODO: Test group cache set
}

func TestGroupPromotions(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	_, err := c.GroupPromotions.Get(-1)
	recordMustNotExist(t, err, "GP #-1 shouldn't exist")
	_, err = c.GroupPromotions.Get(0)
	recordMustNotExist(t, err, "GP #0 shouldn't exist")
	_, err = c.GroupPromotions.Get(1)
	recordMustNotExist(t, err, "GP #1 shouldn't exist")
	expectNilErr(t, c.GroupPromotions.Delete(1))

	//GetByGroup(gid int) (gps []*GroupPromotion, err error)

	testPromo := func(exid, from, to, level, posts, registeredFor int, shouldFail bool) {
		gpid, err := c.GroupPromotions.Create(from, to, false, level, posts, registeredFor)
		expect(t, gpid == exid, fmt.Sprintf("gpid should be %d not %d", exid, gpid))
		//fmt.Println("gpid:", gpid)
		gp, err := c.GroupPromotions.Get(gpid)
		expectNilErr(t, err)
		expect(t, gp.ID == gpid, fmt.Sprintf("gp.ID should be %d not %d", gpid, gp.ID))
		expect(t, gp.From == from, fmt.Sprintf("gp.From should be %d not %d", from, gp.From))
		expect(t, gp.To == to, fmt.Sprintf("gp.To should be %d not %d", to, gp.To))
		expect(t, !gp.TwoWay, "gp.TwoWay should be false not true")
		expect(t, gp.Level == level, fmt.Sprintf("gp.Level should be %d not %d", level, gp.Level))
		expect(t, gp.Posts == posts, fmt.Sprintf("gp.Posts should be %d not %d", posts, gp.Posts))
		expect(t, gp.MinTime == 0, fmt.Sprintf("gp.MinTime should be %d not %d", 0, gp.MinTime))
		expect(t, gp.RegisteredFor == registeredFor, fmt.Sprintf("gp.RegisteredFor should be %d not %d", registeredFor, gp.RegisteredFor))

		uid, err := c.Users.Create("Lord_"+strconv.Itoa(gpid), "I_Rule", "", from, false)
		expectNilErr(t, err)
		u, err := c.Users.Get(uid)
		expectNilErr(t, err)
		expect(t, u.ID == uid, fmt.Sprintf("u.ID should be %d not %d", uid, u.ID))
		expect(t, u.Group == from, fmt.Sprintf("u.Group should be %d not %d", from, u.Group))
		err = c.GroupPromotions.PromoteIfEligible(u, u.Level, u.Posts, u.CreatedAt)
		expectNilErr(t, err)
		u.CacheRemove()
		u, err = c.Users.Get(uid)
		expectNilErr(t, err)
		expect(t, u.ID == uid, fmt.Sprintf("u.ID should be %d not %d", uid, u.ID))
		if shouldFail {
			expect(t, u.Group == from, fmt.Sprintf("u.Group should be (from-group) %d not %d", from, u.Group))
		} else {
			expect(t, u.Group == to, fmt.Sprintf("u.Group should be (to-group)%d not %d", to, u.Group))
		}

		expectNilErr(t, c.GroupPromotions.Delete(gpid))
		_, err = c.GroupPromotions.Get(gpid)
		recordMustNotExist(t, err, fmt.Sprintf("GP #%d should no longer exist", gpid))
	}
	testPromo(1, 1, 2, 0, 0, 0, false)
	testPromo(2, 1, 2, 5, 5, 0, true)
	testPromo(3, 1, 2, 0, 0, 1, true)
}

func TestReplyStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	_, err := c.Rstore.Get(-1)
	recordMustNotExist(t, err, "RID #-1 shouldn't exist")
	_, err = c.Rstore.Get(0)
	recordMustNotExist(t, err, "RID #0 shouldn't exist")

	c.Config.DisablePostIP = false
	testReplyStore(t, 2, 1, "::1")
	c.Config.DisablePostIP = true
	testReplyStore(t, 5, 3, "")
}

func testReplyStore(t *testing.T, newID, newPostCount int, ip string) {
	replyTest2 := func(r *c.Reply, err error, rid, parentID, createdBy int, content, ip string) {
		expectNilErr(t, err)
		expect(t, r.ID == rid, fmt.Sprintf("RID #%d has the wrong ID. It should be %d not %d", rid, rid, r.ID))
		expect(t, r.ParentID == parentID, fmt.Sprintf("The parent topic of RID #%d should be %d not %d", rid, parentID, r.ParentID))
		expect(t, r.CreatedBy == createdBy, fmt.Sprintf("The creator of RID #%d should be %d not %d", rid, createdBy, r.CreatedBy))
		expect(t, r.Content == content, fmt.Sprintf("The contents of RID #%d should be '%s' not %s", rid, content, r.Content))
		expect(t, r.IP == ip, fmt.Sprintf("The IP of RID#%d should be '%s' not %s", rid, ip, r.IP))
	}

	replyTest := func(rid, parentID, createdBy int, content, ip string) {
		reply, err := c.Rstore.Get(rid)
		replyTest2(reply, err, rid, parentID, createdBy, content, ip)
		reply, err = c.Rstore.GetCache().Get(rid)
		replyTest2(reply, err, rid, parentID, createdBy, content, ip)
	}
	replyTest(1, 1, 1, "A reply!", "::1")

	// ! This is hard to do deterministically as the system may pre-load certain items but let's give it a try:
	//_, err = c.Rstore.GetCache().Get(1)
	//recordMustNotExist(t, err, "RID #1 shouldn't be in the cache")

	_, err := c.Rstore.Get(newID)
	recordMustNotExist(t, err, "RID #2 shouldn't exist")

	topic, err := c.Topics.Get(1)
	expectNilErr(t, err)
	expect(t, topic.PostCount == newPostCount, fmt.Sprintf("TID #%d's post count should be %d, not %d", topic.ID, newPostCount, topic.PostCount))

	_, err = c.Rstore.GetCache().Get(newID)
	recordMustNotExist(t, err, "RID #%d shouldn't be in the cache", newID)

	rid, err := c.Rstore.Create(topic, "Fofofo", ip, 1)
	expectNilErr(t, err)
	expect(t, rid == newID, fmt.Sprintf("The next reply ID should be %d not %d", newID, rid))
	expect(t, topic.PostCount == newPostCount, fmt.Sprintf("The old TID #1 in memory's post count should be %d, not %d", newPostCount+1, topic.PostCount))
	// TODO: Test the reply count on the topic

	replyTest(newID, 1, 1, "Fofofo", ip)

	topic, err = c.Topics.Get(1)
	expectNilErr(t, err)
	expect(t, topic.PostCount == newPostCount+1, fmt.Sprintf("TID #1's post count should be %d, not %d", newPostCount+1, topic.PostCount))

	err = topic.CreateActionReply("destroy", ip, 1)
	expectNilErr(t, err)
	expect(t, topic.PostCount == newPostCount+1, fmt.Sprintf("The old TID #1 in memory's post count should be %d, not %d", newPostCount+1, topic.PostCount))
	replyTest(newID+1, 1, 1, "", ip)
	// TODO: Check the actionType field of the reply, this might not be loaded by TopicStore, maybe we should add it there?

	topic, err = c.Topics.Get(1)
	expectNilErr(t, err)
	expect(t, topic.PostCount == newPostCount+2, fmt.Sprintf("TID #1's post count should be %d, not %d", newPostCount+2, topic.PostCount))

	// TODO: Expand upon this
	rid, err = c.Rstore.Create(topic, "hiii", ip, 1)
	expectNilErr(t, err)
	replyTest(rid, topic.ID, 1, "hiii", ip)

	reply, err := c.Rstore.Get(rid)
	expectNilErr(t, err)
	expectNilErr(t, reply.SetPost("huuu"))
	expect(t, reply.Content == "hiii", fmt.Sprintf("topic.Content should be hiii, not %s", reply.Content))
	reply, err = c.Rstore.Get(rid)
	expectNilErr(t, err)
	expect(t, reply.Content == "huuu", fmt.Sprintf("topic.Content should be huuu, not %s", reply.Content))
	expectNilErr(t, reply.Delete())
	// No pointer shenanigans x.x
	expect(t, reply.ID == rid, fmt.Sprintf("pointer shenanigans"))

	_, err = c.Rstore.GetCache().Get(rid)
	recordMustNotExist(t, err, fmt.Sprintf("RID #%d shouldn't be in the cache", rid))
	_, err = c.Rstore.Get(rid)
	recordMustNotExist(t, err, fmt.Sprintf("RID #%d shouldn't exist", rid))

	// TODO: Write a test for this
	//(topic *TopicUser) Replies(offset int, pFrag int, user *User) (rlist []*ReplyUser, ogdesc string, err error)

	// TODO: Add tests for *Reply
	// TODO: Add tests for ReplyCache
}

func TestProfileReplyStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	_, err := c.Prstore.Get(-1)
	recordMustNotExist(t, err, "PRID #-1 shouldn't exist")
	_, err = c.Prstore.Get(0)
	recordMustNotExist(t, err, "PRID #0 shouldn't exist")
	_, err = c.Prstore.Get(1)
	recordMustNotExist(t, err, "PRID #1 shouldn't exist")

	c.Config.DisablePostIP = false
	testProfileReplyStore(t, 1, "::1")
	c.Config.DisablePostIP = true
	testProfileReplyStore(t, 2, "")
}
func testProfileReplyStore(t *testing.T, newID int, ip string) {
	// ? - Commented this one out as strong constraints like this put an unreasonable load on the database, we only want errors if a delete which should succeed fails
	//profileReply := c.BlankProfileReply(1)
	//err = profileReply.Delete()
	//expect(t,err != nil,"You shouldn't be able to delete profile replies which don't exist")

	profileID := 1
	prid, err := c.Prstore.Create(profileID, "Haha", 1, ip)
	expectNilErr(t, err)
	expect(t, prid == newID, fmt.Sprintf("The first profile reply should have an ID of %d", newID))

	pr, err := c.Prstore.Get(newID)
	expectNilErr(t, err)
	expect(t, pr.ID == newID, fmt.Sprintf("The profile reply should have an ID of %d not %d", newID, pr.ID))
	expect(t, pr.ParentID == 1, fmt.Sprintf("The parent ID of the profile reply should be 1 not %d", pr.ParentID))
	expect(t, pr.Content == "Haha", fmt.Sprintf("The profile reply's contents should be 'Haha' not '%s'", pr.Content))
	expect(t, pr.CreatedBy == 1, fmt.Sprintf("The profile reply's creator should be 1 not %d", pr.CreatedBy))
	expect(t, pr.IP == ip, fmt.Sprintf("The profile reply's IP should be '%s' not '%s'", ip, pr.IP))

	err = pr.Delete()
	expectNilErr(t, err)
	_, err = c.Prstore.Get(newID)
	expect(t, err != nil, fmt.Sprintf("PRID #%d shouldn't exist after being deleted", newID))

	// TODO: Test pr.SetBody() and pr.Creator()
}

func TestActivityStream(t *testing.T) {
	miscinit(t)

	expect(t, c.Activity.Count() == 0, "activity stream count should be 0")

	_, err := c.Activity.Get(-1)
	recordMustNotExist(t, err, "activity item -1 shouldn't exist")
	_, err = c.Activity.Get(0)
	recordMustNotExist(t, err, "activity item 0 shouldn't exist")
	_, err = c.Activity.Get(1)
	recordMustNotExist(t, err, "activity item 1 shouldn't exist")

	a := c.Alert{ActorID: 1, TargetUserID: 1, Event: "like", ElementType: "topic", ElementID: 1}
	id, err := c.Activity.Add(a)
	expectNilErr(t, err)
	expect(t, id == 1, "new activity item id should be 1")

	expect(t, c.Activity.Count() == 1, "activity stream count should be 1")
	alert, err := c.Activity.Get(1)
	expectNilErr(t, err)
	expect(t, alert.ActorID == 1, "alert actorid should be 1")
	expect(t, alert.TargetUserID == 1, "alert targetuserid should be 1")
	expect(t, alert.Event == "like", "alert event type should be like")
	expect(t, alert.ElementType == "topic", "alert element type should be topic")
	expect(t, alert.ElementID == 1, "alert element id should be 1")
}

func TestLogs(t *testing.T) {
	miscinit(t)
	gTests := func(s c.LogStore, phrase string) {
		expect(t, s.Count() == 0, "There shouldn't be any "+phrase)
		logs, err := s.GetOffset(0, 25)
		expectNilErr(t, err)
		expect(t, len(logs) == 0, "The log slice should be empty")
	}
	gTests(c.ModLogs, "modlogs")
	gTests(c.AdminLogs, "adminlogs")

	gTests2 := func(s c.LogStore, phrase string) {
		err := s.Create("something", 0, "bumblefly", "::1", 1)
		expectNilErr(t, err)
		count := s.Count()
		expect(t, count == 1, fmt.Sprintf("store.Count should return one, not %d", count))
		logs, err := s.GetOffset(0, 25)
		recordMustExist(t, err, "We should have at-least one "+phrase)
		expect(t, len(logs) == 1, "The length of the log slice should be one")

		log := logs[0]
		expect(t, log.Action == "something", "log.Action is not something")
		expect(t, log.ElementID == 0, "log.ElementID is not 0")
		expect(t, log.ElementType == "bumblefly", "log.ElementType is not bumblefly")
		expect(t, log.IP == "::1", "log.IP is not ::1")
		expect(t, log.ActorID == 1, "log.ActorID is not 1")
		// TODO: Add a test for log.DoneAt? Maybe throw in some dates and times which are clearly impossible but which may occur due to timezone bugs?
	}
	gTests2(c.ModLogs, "modlog")
	gTests2(c.AdminLogs, "adminlog")
}

// TODO: Add tests for registration logs

func TestPluginManager(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	_, ok := c.Plugins["fairy-dust"]
	expect(t, !ok, "Plugin fairy-dust shouldn't exist")
	pl, ok := c.Plugins["bbcode"]
	expect(t, ok, "Plugin bbcode should exist")
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err := pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err := pl.InDatabase()
	expectNilErr(t, err)
	expect(t, !hasPlugin, "Plugin bbcode shouldn't exist in the database")
	// TODO: Add some test cases for SetActive and SetInstalled before calling AddToDatabase

	expectNilErr(t, pl.AddToDatabase(true, false))
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, pl.Active, "Plugin bbcode should be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, active, "Plugin bbcode should be active in the database too")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should exist in the database")
	expect(t, pl.Init != nil, "Plugin bbcode should have an init function")
	expectNilErr(t, pl.Init(pl))

	expectNilErr(t, pl.SetActive(true))
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, pl.Active, "Plugin bbcode should still be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, active, "Plugin bbcode should still be active in the database too")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	expectNilErr(t, pl.SetActive(false))
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")
	expect(t, pl.Deactivate != nil, "Plugin bbcode should have an init function")
	pl.Deactivate(pl) // Returns nothing

	// Not installable, should not be mutated
	expect(t, pl.SetInstalled(true) == c.ErrPluginNotInstallable, "Plugin was set as installed despite not being installable")
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	expect(t, pl.SetInstalled(false) == c.ErrPluginNotInstallable, "Plugin was set as not installed despite not being installable")
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	// This isn't really installable, but we want to get a few tests done before getting plugins which are stateful
	pl.Installable = true
	expectNilErr(t, pl.SetInstalled(true))
	expect(t, pl.Installable, "Plugin bbcode should be installable")
	expect(t, pl.Installed, "Plugin bbcode should be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	expectNilErr(t, pl.SetInstalled(false))
	expect(t, pl.Installable, "Plugin bbcode should be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	// Bugs sometimes arise when we try to delete a hook when there are multiple, so test for that
	// TODO: Do a finer grained test for that case...? A bigger test might catch more odd cases with multiple plugins
	pl2, ok := c.Plugins["markdown"]
	expect(t, ok, "Plugin markdown should exist")
	expect(t, !pl2.Installable, "Plugin markdown shouldn't be installable")
	expect(t, !pl2.Installed, "Plugin markdown shouldn't be 'installed'")
	expect(t, !pl2.Active, "Plugin markdown shouldn't be active")
	active, err = pl2.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin markdown shouldn't be active in the database either")
	hasPlugin, err = pl2.InDatabase()
	expectNilErr(t, err)
	expect(t, !hasPlugin, "Plugin markdown shouldn't exist in the database")

	expectNilErr(t, pl2.AddToDatabase(true, false))
	expectNilErr(t, pl2.Init(pl2))
	expectNilErr(t, pl.SetActive(true))
	expectNilErr(t, pl.Init(pl))
	pl2.Deactivate(pl2)
	expectNilErr(t, pl2.SetActive(false))
	pl.Deactivate(pl)
	expectNilErr(t, pl.SetActive(false))

	// Hook tests
	ht := func() *c.HookTable {
		return c.GetHookTable()
	}
	expect(t, ht().Sshook("haha", "ho") == "ho", "Sshook shouldn't have anything bound to it yet")
	handle := func(in string) (out string) {
		return in + "hi"
	}
	pl.AddHook("haha", handle)
	expect(t, ht().Sshook("haha", "ho") == "hohi", "Sshook didn't give hohi")
	pl.RemoveHook("haha", handle)
	expect(t, ht().Sshook("haha", "ho") == "ho", "Sshook shouldn't have anything bound to it anymore")

	expect(t, ht().Hook("haha", "ho") == "ho", "Hook shouldn't have anything bound to it yet")
	handle2 := func(inI interface{}) (out interface{}) {
		return inI.(string) + "hi"
	}
	pl.AddHook("hehe", handle2)
	expect(t, ht().Hook("hehe", "ho").(string) == "hohi", "Hook didn't give hohi")
	pl.RemoveHook("hehe", handle2)
	expect(t, ht().Hook("hehe", "ho").(string) == "ho", "Hook shouldn't have anything bound to it anymore")

	// TODO: Add tests for more hook types
}

func TestPhrases(t *testing.T) {
	getPhrase := phrases.GetPermPhrase
	tp := func(name, expects string) {
		res := getPhrase(name)
		expect(t, res == expects, "Not the expected phrase, got '"+res+"' instead")
	}
	tp("BanUsers", "Can ban users")
	tp("NoSuchPerm", "{lang.perms[NoSuchPerm]}")
	tp("ViewTopic", "Can view topics")
	tp("NoSuchPerm", "{lang.perms[NoSuchPerm]}")

	// TODO: Cover the other phrase types, also try switching between languages to see if anything strange happens
}

func TestMetaStore(t *testing.T) {
	m, err := c.Meta.Get("magic")
	expect(t, m == "", "meta var magic should be empty")
	recordMustNotExist(t, err, "meta var magic should not exist")

	err = c.Meta.Set("magic", "lol")
	expectNilErr(t, err)

	m, err = c.Meta.Get("magic")
	expectNilErr(t, err)
	expect(t, m == "lol", "meta var magic should be lol")

	err = c.Meta.Set("magic", "wha")
	expectNilErr(t, err)

	m, err = c.Meta.Get("magic")
	expectNilErr(t, err)
	expect(t, m == "wha", "meta var magic should be wha")

	m, err = c.Meta.Get("giggle")
	expect(t, m == "", "meta var giggle should be empty")
	recordMustNotExist(t, err, "meta var giggle should not exist")
}

func TestPages(t *testing.T) {
	expect(t, c.Pages.Count() == 0, "Page count should be 0")
	_, err := c.Pages.Get(1)
	recordMustNotExist(t, err, "Page 1 should not exist yet")
	expectNilErr(t, c.Pages.Delete(-1))
	expectNilErr(t, c.Pages.Delete(0))
	expectNilErr(t, c.Pages.Delete(1))
	_, err = c.Pages.Get(1)
	recordMustNotExist(t, err, "Page 1 should not exist yet")
	//err = c.Pages.Reload(1)
	//recordMustNotExist(t,err,"Page 1 should not exist yet")

	ipage := c.BlankCustomPage()
	ipage.Name = "test"
	ipage.Title = "Test"
	ipage.Body = "A test page"
	pid, err := ipage.Create()
	expectNilErr(t, err)
	expect(t, pid == 1, "The first page should have an ID of 1")
	expect(t, c.Pages.Count() == 1, "Page count should be 1")

	page, err := c.Pages.Get(1)
	expectNilErr(t, err)
	expect(t, page.Name == ipage.Name, "The page name should be "+ipage.Name)
	expect(t, page.Title == ipage.Title, "The page title should be "+ipage.Title)
	expect(t, page.Body == ipage.Body, "The page body should be "+ipage.Body)

	opage, err := c.Pages.Get(1)
	expectNilErr(t, err)
	opage.Name = "t"
	opage.Title = "T"
	opage.Body = "testing"
	expectNilErr(t, opage.Commit())

	page, err = c.Pages.Get(1)
	expectNilErr(t, err)
	expect(t, page.Name == opage.Name, "The page name should be "+opage.Name)
	expect(t, page.Title == opage.Title, "The page title should be "+opage.Title)
	expect(t, page.Body == opage.Body, "The page body should be "+opage.Body)

	err = c.Pages.Delete(1)
	expectNilErr(t, err)
	expect(t, c.Pages.Count() == 0, "Page count should be 0")
	_, err = c.Pages.Get(1)
	recordMustNotExist(t, err, "Page 1 should not exist")
	//err = c.Pages.Reload(1)
	//recordMustNotExist(t,err,"Page 1 should not exist")

	// TODO: More tests
}

func TestWordFilters(t *testing.T) {
	// TODO: Test the word filters and their store
	expect(t, c.WordFilters.Length() == 0, "Word filter list should be empty")
	expect(t, c.WordFilters.EstCount() == 0, "Word filter list should be empty")
	expect(t, c.WordFilters.Count() == 0, "Word filter list should be empty")
	filters, err := c.WordFilters.GetAll()
	expectNilErr(t, err) // TODO: Slightly confusing that we don't get ErrNoRow here
	expect(t, len(filters) == 0, "Word filter map should be empty")
	// TODO: Add a test for ParseMessage relating to word filters
	_, err = c.WordFilters.Get(1)
	recordMustNotExist(t, err, "filter 1 should not exist")

	wfid, err := c.WordFilters.Create("imbecile", "lovely")
	expectNilErr(t, err)
	expect(t, wfid == 1, "The first word filter should have an ID of 1")
	expect(t, c.WordFilters.Length() == 1, "Word filter list should not be empty")
	expect(t, c.WordFilters.EstCount() == 1, "Word filter list should not be empty")
	expect(t, c.WordFilters.Count() == 1, "Word filter list should not be empty")

	filters, err = c.WordFilters.GetAll()
	expectNilErr(t, err)
	expect(t, len(filters) == 1, "Word filter map should not be empty")
	filter := filters[1]
	expect(t, filter.ID == 1, "Word filter ID should be 1")
	expect(t, filter.Find == "imbecile", "Word filter needle should be imbecile")
	expect(t, filter.Replace == "lovely", "Word filter replacement should be lovely")

	filter, err = c.WordFilters.Get(1)
	expectNilErr(t, err)
	expect(t, filter.ID == 1, "Word filter ID should be 1")
	expect(t, filter.Find == "imbecile", "Word filter needle should be imbecile")
	expect(t, filter.Replace == "lovely", "Word filter replacement should be lovely")

	// Update
	expectNilErr(t, c.WordFilters.Update(1, "b", "a"))

	expect(t, c.WordFilters.Length() == 1, "Word filter list should not be empty")
	expect(t, c.WordFilters.EstCount() == 1, "Word filter list should not be empty")
	expect(t, c.WordFilters.Count() == 1, "Word filter list should not be empty")

	filters, err = c.WordFilters.GetAll()
	expectNilErr(t, err)
	expect(t, len(filters) == 1, "Word filter map should not be empty")
	filter = filters[1]
	expect(t, filter.ID == 1, "Word filter ID should be 1")
	expect(t, filter.Find == "b", "Word filter needle should be b")
	expect(t, filter.Replace == "a", "Word filter replacement should be a")

	filter, err = c.WordFilters.Get(1)
	expectNilErr(t, err)
	expect(t, filter.ID == 1, "Word filter ID should be 1")
	expect(t, filter.Find == "b", "Word filter needle should be imbecile")
	expect(t, filter.Replace == "a", "Word filter replacement should be a")

	// TODO: Add a test for ParseMessage relating to word filters

	err = c.WordFilters.Delete(1)
	expectNilErr(t, err)

	expect(t, c.WordFilters.Length() == 0, "Word filter list should be empty")
	expect(t, c.WordFilters.EstCount() == 0, "Word filter list should be empty")
	expect(t, c.WordFilters.Count() == 0, "Word filter list should be empty")
	filters, err = c.WordFilters.GetAll()
	expectNilErr(t, err) // TODO: Slightly confusing that we don't get ErrNoRow here
	expect(t, len(filters) == 0, "Word filter map should be empty")
	_, err = c.WordFilters.Get(1)
	recordMustNotExist(t, err, "filter 1 should not exist")

	// TODO: Any more tests we could do?
}

// TODO: Expand upon the valid characters which can go in URLs?
func TestSlugs(t *testing.T) {
	l := &MEPairList{nil}
	c.Config.BuildSlugs = true // Flip this switch, otherwise all the tests will fail

	l.Add("Unknown", "unknown")
	l.Add("Unknown2", "unknown2")
	l.Add("Unknown ", "unknown")
	l.Add("Unknown 2", "unknown-2")
	l.Add("Unknown  2", "unknown-2")
	l.Add("Admin Alice", "admin-alice")
	l.Add("Admin_Alice", "adminalice")
	l.Add("Admin_Alice-", "adminalice")
	l.Add("-Admin_Alice-", "adminalice")
	l.Add("-Admin@Alice-", "adminalice")
	l.Add("-Admin😀Alice-", "adminalice")
	l.Add("u", "u")
	l.Add("", "untitled")
	l.Add(" ", "untitled")
	l.Add("-", "untitled")
	l.Add("--", "untitled")
	l.Add("é", "é")
	l.Add("-é-", "é")
	l.Add("-你好-", "untitled")
	l.Add("-こにちは-", "untitled")

	for _, item := range l.Items {
		t.Log("Testing string '" + item.Msg + "'")
		res := c.NameToSlug(item.Msg)
		if res != item.Expects {
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", item.Expects)
		}
	}
}

func TestWidgets(t *testing.T) {
	_, err := c.Widgets.Get(1)
	recordMustNotExist(t, err, "There shouldn't be any widgets by default")
	widgets := c.Docks.RightSidebar.Items
	expect(t, len(widgets) == 0, fmt.Sprintf("RightSidebar should have 0 items, not %d", len(widgets)))

	widget := &c.Widget{Position: 0, Side: "rightSidebar", Type: "simple", Enabled: true, Location: "global"}
	ewidget := &c.WidgetEdit{widget, map[string]string{"Name": "Test", "Text": "Testing"}}
	wid, err := ewidget.Create()
	expectNilErr(t, err)
	expect(t, wid == 1, "wid should be 1")

	wtest := func(w, w2 *c.Widget) {
		expect(t, w.Position == w2.Position, "wrong position")
		expect(t, w.Side == w2.Side, "wrong side")
		expect(t, w.Type == w2.Type, "wrong type")
		expect(t, w2.Enabled, "not enabled")
		expect(t, w.Location == w2.Location, "wrong location")
	}

	// TODO: Do a test for the widget body
	widget2, err := c.Widgets.Get(1)
	expectNilErr(t, err)
	wtest(widget, widget2)

	widgets = c.Docks.RightSidebar.Items
	expect(t, len(widgets) == 1, fmt.Sprintf("RightSidebar should have 1 item, not %d", len(widgets)))
	wtest(widget, widgets[0])

	widget2.Enabled = false
	ewidget = &c.WidgetEdit{widget2, map[string]string{"Name": "Test", "Text": "Testing"}}
	err = ewidget.Commit()
	expectNilErr(t, err)

	widget2, err = c.Widgets.Get(1)
	expectNilErr(t, err)
	expect(t, widget2.Position == widget.Position, "wrong position")
	expect(t, widget2.Side == widget.Side, "wrong side")
	expect(t, widget2.Type == widget.Type, "wrong type")
	expect(t, !widget2.Enabled, "should not be enabled")
	expect(t, widget2.Location == widget.Location, "wrong location")

	widgets = c.Docks.RightSidebar.Items
	expect(t, len(widgets) == 1, fmt.Sprintf("RightSidebar should have 1 item, not %d", len(widgets)))
	expect(t, widgets[0].Position == widget.Position, "wrong position")
	expect(t, widgets[0].Side == widget.Side, "wrong side")
	expect(t, widgets[0].Type == widget.Type, "wrong type")
	expect(t, !widgets[0].Enabled, "should not be enabled")
	expect(t, widgets[0].Location == widget.Location, "wrong location")

	err = widget2.Delete()
	expectNilErr(t, err)

	_, err = c.Widgets.Get(1)
	recordMustNotExist(t, err, "There shouldn't be any widgets anymore")
	widgets = c.Docks.RightSidebar.Items
	expect(t, len(widgets) == 0, fmt.Sprintf("RightSidebar should have 0 items, not %d", len(widgets)))
}

func TestAuth(t *testing.T) {
	// bcrypt likes doing stupid things, so this test will probably fail
	realPassword := "Madame Cassandra's Mystic Orb"
	t.Logf("Set realPassword to '%s'", realPassword)
	t.Log("Hashing the real password with bcrypt")
	hashedPassword, _, err := c.BcryptGeneratePassword(realPassword)
	if err != nil {
		t.Error(err)
	}
	passwordTest(t, realPassword, hashedPassword)
	// TODO: Peek at the prefix to verify this is a bcrypt hash

	t.Log("Hashing the real password")
	hashedPassword2, _, err := c.GeneratePassword(realPassword)
	if err != nil {
		t.Error(err)
	}
	passwordTest(t, realPassword, hashedPassword2)
	// TODO: Peek at the prefix to verify this is a bcrypt hash

	_, err, _ = c.Auth.Authenticate("None", "password")
	errmsg := "Username None shouldn't exist"
	if err != nil {
		errmsg += "\n" + err.Error()
	}
	expect(t, err == c.ErrNoUserByName, errmsg)

	uid, err, _ := c.Auth.Authenticate("Admin", "password")
	expectNilErr(t, err)
	expect(t, uid == 1, fmt.Sprintf("Default admin uid should be 1 not %d", uid))

	_, err, _ = c.Auth.Authenticate("Sam", "ReallyBadPassword")
	errmsg = "Username Sam shouldn't exist"
	if err != nil {
		errmsg += "\n" + err.Error()
	}
	expect(t, err == c.ErrNoUserByName, errmsg)

	admin, err := c.Users.Get(1)
	expectNilErr(t, err)
	// TODO: Move this into the user store tests to provide better coverage? E.g. To see if the installer and the user creator initialise the field differently
	expect(t, admin.Session == "", "Admin session should be blank")

	session, err := c.Auth.CreateSession(1)
	expectNilErr(t, err)
	expect(t, session != "", "Admin session shouldn't be blank")
	// TODO: Test the actual length set in the setting in addition to this "too short" test
	// TODO: We might be able to push up this minimum requirement
	expect(t, len(session) > 10, "Admin session shouldn't be too short")
	expect(t, admin.Session != session, "Old session should not match new one")
	admin, err = c.Users.Get(1)
	expectNilErr(t, err)
	expect(t, admin.Session == session, "Sessions should match")

	// TODO: Create a user with a unicode password and see if we can login as them
	// TODO: Tests for SessionCheck, GetCookies, and ForceLogout
}

// TODO: Vary the salts? Keep in mind that some algorithms store the salt in the hash therefore the salt string may be blank
func passwordTest(t *testing.T, realPassword, hashedPassword string) {
	if len(hashedPassword) < 10 {
		t.Error("Hash too short")
	}
	salt := ""
	password := realPassword
	t.Logf("Testing password '%s'", password)
	t.Logf("Testing salt '%s'", salt)
	err := c.CheckPassword(hashedPassword, password, salt)
	if err == c.ErrMismatchedHashAndPassword {
		t.Error("The two don't match")
	} else if err == c.ErrPasswordTooLong {
		t.Error("CheckPassword thinks the password is too long")
	} else if err != nil {
		t.Error(err)
	}

	password = "hahaha"
	t.Logf("Testing password '%s'", password)
	t.Logf("Testing salt '%s'", salt)
	err = c.CheckPassword(hashedPassword, password, salt)
	if err == c.ErrPasswordTooLong {
		t.Error("CheckPassword thinks the password is too long")
	} else if err == nil {
		t.Error("The two shouldn't match!")
	}

	password = "Madame Cassandra's Mystic"
	t.Logf("Testing password '%s'", password)
	t.Logf("Testing salt '%s'", salt)
	err = c.CheckPassword(hashedPassword, password, salt)
	expect(t, err != c.ErrPasswordTooLong, "CheckPassword thinks the password is too long")
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

func (l *METriList) Add(args ...string) {
	if len(args) < 2 {
		panic("need 2 or more args")
	}
	if len(args) > 2 {
		l.Items = append(l.Items, METri{args[0], args[1], args[2]})
	} else {
		l.Items = append(l.Items, METri{"", args[0], args[1]})
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

func (l *CountTestList) Add(name, msg string, expects int) {
	l.Items = append(l.Items, CountTest{name, msg, expects})
}

func TestWordCount(t *testing.T) {
	l := &CountTestList{nil}
	l.Add("blank", "", 0)
	l.Add("single-letter", "h", 1)
	l.Add("single-kana", "お", 1)
	l.Add("single-letter-words", "h h", 2)
	l.Add("two-letter", "h", 1)
	l.Add("two-kana", "おは", 1)
	l.Add("two-letter-words", "hh hh", 2)
	l.Add("", "h,h", 2)
	l.Add("", "h,,h", 2)
	l.Add("", "h, h", 2)
	l.Add("", "  h, h", 2)
	l.Add("", "h, h  ", 2)
	l.Add("", "  h, h  ", 2)
	l.Add("", "h,  h", 2)
	l.Add("", "h\nh", 2)
	l.Add("", "h\"h", 2)
	l.Add("", "h[r]h", 3)
	l.Add("", "お,お", 2)
	l.Add("", "お、お", 2)
	l.Add("", "お\nお", 2)
	l.Add("", "お”お", 2)
	l.Add("", "お「あ」お", 3)

	for _, item := range l.Items {
		res := c.WordCount(item.Msg)
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
