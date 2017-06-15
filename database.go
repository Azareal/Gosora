package main

import "log"
import "fmt"
import "encoding/json"
import "database/sql"

var db *sql.DB
var db_version string
var db_collation string = "utf8mb4_general_ci"

func init_database() (err error) {
  // Engine specific code
  err = _init_database()
  if err != nil {
    return err
  }

  log.Print("Loading the usergroups.")
	groups = append(groups, Group{ID:0,Name:"System"})

	rows, err := get_groups_stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 1
	for ;rows.Next();i++ {
		group := Group{ID: 0,}
		err := rows.Scan(&group.ID, &group.Name, &group.PermissionsText, &group.Is_Mod, &group.Is_Admin, &group.Is_Banned, &group.Tag)
		if err != nil {
			return err
		}

		// Ugh, you really shouldn't physically delete these items, it makes a big mess of things
		if group.ID != i {
			log.Print("Stop physically deleting groups. You are messing up the IDs. Use the Group Manager or delete_group() instead x.x")
			fill_group_id_gap(i, group.ID)
		}

		err = json.Unmarshal(group.PermissionsText, &group.Perms)
		if err != nil {
			return err
		}
		if debug {
			log.Print(group.Name + ": ")
			fmt.Printf("%+v\n", group.Perms)
		}

		group.Perms.ExtData = make(map[string]bool)
		groups = append(groups, group)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	groupCapCount = i

	log.Print("Binding the Not Loggedin Group")
	GuestPerms = groups[6].Perms

	log.Print("Loading the forums.")
	err = LoadForums()
	if err != nil {
		return err
	}

	log.Print("Loading the forum permissions.")
	err = build_forum_permissions()
	if err != nil {
		return err
	}
  fstore = NewStaticForumStore()

	log.Print("Loading the settings.")
	err = LoadSettings()
	if err != nil {
		return err
	}

	log.Print("Loading the plugins.")
	err = LoadPlugins()
	if err != nil {
		return err
	}

	log.Print("Loading the themes.")
	err = LoadThemes()
	if err != nil {
		return err
	}

  return nil
}
