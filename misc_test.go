package main

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
    t.Error("user.ID doesn't not match the requested UID")
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
