package main

import "strconv"
import "testing"

// TO-DO: Generate a test database to work with rather than a live one
func TestUserStore(t *testing.T) {
  if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

  var user *User
  var err error

  user, err = users.CascadeGet(-1)
  if err == nil {
    t.Error("UID #-1 shouldn't exist")
  } else if err != ErrNoRows {
    t.Fatal(err)
  }

  user, err = users.CascadeGet(0)
  if err == nil {
    t.Error("UID #0 shouldn't exist")
  } else if err != ErrNoRows {
    t.Fatal(err)
  }

  user, err = users.CascadeGet(1)
  if err == ErrNoRows {
    t.Error("Couldn't find UID #1")
  } else if err != nil {
    t.Fatal(err)
  }

  if user.ID != 1 {
    t.Error("user.ID doesn't not match the requested UID. Got '" + strconv.Itoa(user.ID) + "' instead.")
  }
}

func TestForumStore(t *testing.T) {
  if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

  var forum *Forum
  var err error

  forum, err = fstore.CascadeGet(-1)
  if err == nil {
    t.Error("FID #-1 shouldn't exist")
  } else if err != ErrNoRows {
    t.Fatal(err)
  }

  forum, err = fstore.CascadeGet(0)
  if err == ErrNoRows {
    t.Error("Couldn't find FID #0")
  } else if err != nil {
    t.Fatal(err)
  }

  if forum.ID != 0 {
    t.Error("forum.ID doesn't not match the requested UID. Got '" + strconv.Itoa(forum.ID) + "' instead.")
  }
  if forum.Name != "Uncategorised" {
    t.Error("FID #0 is named '" + forum.Name + "' and not 'Uncategorised'")
  }

  forum, err = fstore.CascadeGet(1)
  if err == ErrNoRows {
    t.Error("Couldn't find FID #1")
  } else if err != nil {
    t.Fatal(err)
  }

  if forum.ID != 1 {
    t.Error("forum.ID doesn't not match the requested UID. Got '" + strconv.Itoa(forum.ID) + "' instead.'")
  }
  if forum.Name != "Reports" {
    t.Error("FID #0 is named '" + forum.Name + "' and not 'Reports'")
  }

  forum, err = fstore.CascadeGet(2)
  if err == ErrNoRows {
    t.Error("Couldn't find FID #2")
  } else if err != nil {
    t.Fatal(err)
  }
}

func TestSlugs(t *testing.T) {
  var res string
  var msgList []ME_Pair

  msgList = addMEPair(msgList,"Unknown","unknown")
  msgList = addMEPair(msgList,"Unknown2","unknown2")
  msgList = addMEPair(msgList,"Unknown ","unknown")
  msgList = addMEPair(msgList,"Unknown 2","unknown-2")
  msgList = addMEPair(msgList,"Unknown  2","unknown-2")
  msgList = addMEPair(msgList,"Admin Alice","admin-alice")
  msgList = addMEPair(msgList,"Admin_Alice","adminalice")
  msgList = addMEPair(msgList,"Admin_Alice-","adminalice")
  msgList = addMEPair(msgList,"-Admin_Alice-","adminalice")
  msgList = addMEPair(msgList,"-Admin@Alice-","adminalice")
  msgList = addMEPair(msgList,"-Admin😀Alice-","adminalice")
  msgList = addMEPair(msgList,"u","u")
  msgList = addMEPair(msgList,"","untitled")
  msgList = addMEPair(msgList," ","untitled")
  msgList = addMEPair(msgList,"-","untitled")
  msgList = addMEPair(msgList,"--","untitled")
  msgList = addMEPair(msgList,"é","é")
  msgList = addMEPair(msgList,"-é-","é")

  for _, item := range msgList {
    t.Log("Testing string '"+item.Msg+"'")
    res = name_to_slug(item.Msg)
    if res != item.Expects {
      t.Error("Bad output:","'"+res+"'")
      t.Error("Expected:",item.Expects)
    }
  }
}

func TestAuth(t *testing.T) {
  // bcrypt likes doing stupid things, so this test will probably fail
  var real_password string
  var hashed_password string
  var password string
  var salt string
  var err error

  /* No extra salt tests, we might not need this extra salt, as bcrypt has it's own? */
  real_password = "Madame Cassandra's Mystic Orb"
  t.Log("Set real_password to '" + real_password + "'")
  t.Log("Hashing the real password")
  hashed_password, err = BcryptGeneratePasswordNoSalt(real_password)
  if err != nil {
    t.Error(err)
  }

  password = real_password
  t.Log("Testing password '" + password + "'")
  t.Log("Testing salt '" + salt + "'")
  err = CheckPassword(hashed_password,password,salt)
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
  err = CheckPassword(hashed_password,password,salt)
  if err == ErrPasswordTooLong {
    t.Error("CheckPassword thinks the password is too long")
  } else if err == nil {
    t.Error("The two shouldn't match!")
  }

  password = "Madame Cassandra's Mystic"
  t.Log("Testing password '" + password + "'")
  t.Log("Testing salt '" + salt + "'")
  err = CheckPassword(hashed_password,password,salt)
  if err == ErrPasswordTooLong {
    t.Error("CheckPassword thinks the password is too long")
  } else if err == nil {
    t.Error("The two shouldn't match!")
  }
}
