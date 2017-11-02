package main

var fpstore *ForumPermsStore

type ForumPermsStore struct {
}

func NewForumPermsStore() *ForumPermsStore {
	return &ForumPermsStore{}
}

func (fps *ForumPermsStore) Get(fid int, gid int) (fperms ForumPerms, err error) {
	// TODO: Add a hook here and have plugin_guilds use it
	group, err := gstore.Get(gid)
	if err != nil {
		return fperms, ErrNoRows
	}
	return group.Forums[fid], nil
}
