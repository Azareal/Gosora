package main

import "log"
import "fmt"
import "strconv"
import "encoding/json"

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
	log.Print("Adding the uncategorised forum")
	forums = append(forums, Forum{0,"Uncategorised","",uncategorised_forum_visible,"all",0,"",0,"",0,""})

	//rows, err = db.Query("SELECT fid, name, active, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime FROM forums")
	//rows, err = db.Query("select `fid`, `name`, `desc`, `active`, `preset`, `topicCount`, `lastTopic`, `lastTopicID`, `lastReplyer`, `lastReplyerID`, `lastTopicTime` from forums order by fid asc")
  rows, err = get_forums_stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	i = 1
	for ;rows.Next();i++ {
		forum := Forum{ID:0,Name:"",Active:true,Preset:"all"}
		err := rows.Scan(&forum.ID, &forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
		if err != nil {
			return err
		}

		// Ugh, you really shouldn't physically delete these items, it makes a big mess of things
		if forum.ID != i {
			log.Print("Stop physically deleting forums. You are messing up the IDs. Use the Forum Manager or delete_forum() instead x.x")
			fill_forum_id_gap(i, forum.ID)
		}

		if forum.Name == "" {
			if debug {
				log.Print("Adding a placeholder forum")
			}
		} else {
			log.Print("Adding the " + forum.Name + " forum")
		}
		forums = append(forums,forum)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	forumCapCount = i

	//log.Print("Adding the reports forum")
	//forums[-1] = Forum{-1,"Reports",false,0,"",0,"",0,""}

	log.Print("Loading the forum permissions")
	rows, err = get_forums_permissions_stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	if debug {
		log.Print("Adding the forum permissions")
	}
	// Temporarily store the forum perms in a map before transferring it to a much faster and thread-safe slice
	forum_perms = make(map[int]map[int]ForumPerms)
	for rows.Next() {
		var gid, fid int
		var perms []byte
		var pperms ForumPerms
		err := rows.Scan(&gid, &fid, &perms)
		if err != nil {
			return err
		}
		err = json.Unmarshal(perms, &pperms)
		if err != nil {
			return err
		}
		pperms.ExtData = make(map[string]bool)
		pperms.Overrides = true
		_, ok := forum_perms[gid]
		if !ok {
			forum_perms[gid] = make(map[int]ForumPerms)
		}
		forum_perms[gid][fid] = pperms
	}
	for gid, _ := range groups {
		if debug {
			log.Print("Adding the forum permissions for Group #" + strconv.Itoa(gid) + " - " + groups[gid].Name)
		}
		//groups[gid].Forums = append(groups[gid].Forums,BlankForumPerms) // GID 0. I sometimes wish MySQL's AUTO_INCREMENT would start at zero
		for fid, _ := range forums {
			forum_perm, ok := forum_perms[gid][fid]
			if ok {
				// Override group perms
				//log.Print("Overriding permissions for forum #" + strconv.Itoa(fid))
				groups[gid].Forums = append(groups[gid].Forums,forum_perm)
			} else {
				// Inherit from Group
				//log.Print("Inheriting from default for forum #" + strconv.Itoa(fid))
				forum_perm = BlankForumPerms
				groups[gid].Forums = append(groups[gid].Forums,forum_perm)
			}

			if forum_perm.Overrides {
				if forum_perm.ViewTopic {
					groups[gid].CanSee = append(groups[gid].CanSee, fid)
				}
			} else if groups[gid].Perms.ViewTopic {
				groups[gid].CanSee = append(groups[gid].CanSee, fid)
			}
		}
		//fmt.Printf("%+v\n", groups[gid].CanSee)
		//fmt.Printf("%+v\n", groups[gid].Forums)
    //fmt.Println(len(groups[gid].CanSee))
		//fmt.Println(len(groups[gid].Forums))
	}

	log.Print("Loading the settings.")
	rows, err = get_full_settings_stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var sname, scontent, stype, sconstraints string
	for rows.Next() {
		err := rows.Scan(&sname, &scontent, &stype, &sconstraints)
		if err != nil {
			return err
		}
		errmsg := parseSetting(sname, scontent, stype, sconstraints)
		if errmsg != "" {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	log.Print("Loading the plugins.")
	rows, err = get_plugins_stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var uname string
	var active bool
	for rows.Next() {
		err := rows.Scan(&uname, &active)
		if err != nil {
			return err
		}

		// Was the plugin deleted at some point?
		plugin, ok := plugins[uname]
		if !ok {
			continue
		}
		plugin.Active = active
		plugins[uname] = plugin
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	log.Print("Loading the themes.")
	rows, err = get_themes_stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var defaultThemeSwitch bool
	for rows.Next() {
		err := rows.Scan(&uname, &defaultThemeSwitch)
		if err != nil {
			return err
		}

		// Was the theme deleted at some point?
		theme, ok := themes[uname]
		if !ok {
			continue
		}

		if defaultThemeSwitch {
			log.Print("Loading the theme '" + theme.Name + "'")
			theme.Active = true
			defaultTheme = uname
			add_theme_static_files(uname)
			map_theme_templates(theme)
		} else {
			theme.Active = false
		}
		themes[uname] = theme
	}
	err = rows.Err()
	if err != nil {
		return err
	}

  return nil
}
