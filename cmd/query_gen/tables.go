package main

import qgen "github.com/Azareal/Gosora/query_gen"

var mysqlPre = "utf8mb4"
var mysqlCol = "utf8mb4_general_ci"

type tblColumn = qgen.DBTableColumn
type tC = tblColumn
type tblKey = qgen.DBTableKey

func createTables(adapter qgen.Adapter) (err error) {
	createTable := func(table string, charset string, collation string, columns []tC, keys []tblKey) {
		if err != nil {
			return
		}
		err = qgen.Install.CreateTable(table, charset, collation, columns, keys)
	}
	createTable("users", mysqlPre, mysqlCol,
		[]tC{
			tC{"uid", "int", 0, false, true, ""},
			tC{"name", "varchar", 100, false, false, ""},
			tC{"password", "varchar", 100, false, false, ""},

			tC{"salt", "varchar", 80, false, false, "''"},
			tC{"group", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"active", "boolean", 0, false, false, "0"},
			tC{"is_super_admin", "boolean", 0, false, false, "0"},
			tC{"createdAt", "createdAt", 0, false, false, ""},
			tC{"lastActiveAt", "datetime", 0, false, false, ""},
			tC{"session", "varchar", 200, false, false, "''"},
			//tC{"authToken", "varchar", 200, false, false, "''"},
			tC{"last_ip", "varchar", 200, false, false, "''"},
			tC{"enable_embeds", "int", 0, false, false, "-1"},
			tC{"email", "varchar", 200, false, false, "''"},
			tC{"avatar", "varchar", 100, false, false, "''"},
			tC{"message", "text", 0, false, false, "''"},

			// TODO: Drop these columns?
			tC{"url_prefix", "varchar", 20, false, false, "''"},
			tC{"url_name", "varchar", 100, false, false, "''"},
			//tC{"pub_key", "text", 0, false, false, "''"},

			tC{"level", "smallint", 0, false, false, "0"},
			tC{"score", "int", 0, false, false, "0"},
			tC{"posts", "int", 0, false, false, "0"},
			tC{"bigposts", "int", 0, false, false, "0"},
			tC{"megaposts", "int", 0, false, false, "0"},
			tC{"topics", "int", 0, false, false, "0"},
			tC{"liked", "int", 0, false, false, "0"},

			// These two are to bound liked queries with little bits of information we know about the user to reduce the server load
			tC{"oldestItemLikedCreatedAt", "datetime", 0, false, false, ""}, // For internal use only, semantics may change
			tC{"lastLiked", "datetime", 0, false, false, ""},                // For internal use only, semantics may change

			//tC{"penalty_count","int",0,false,false,"0"},
			tC{"temp_group", "int", 0, false, false, "0"}, // For temporary groups, set this to zero when a temporary group isn't in effect
		},
		[]tblKey{
			tblKey{"uid", "primary", "", false},
			tblKey{"name", "unique", "", false},
		},
	)

	createTable("users_groups", mysqlPre, mysqlCol,
		[]tC{
			tC{"gid", "int", 0, false, true, ""},
			tC{"name", "varchar", 100, false, false, ""},
			tC{"permissions", "text", 0, false, false, ""},
			tC{"plugin_perms", "text", 0, false, false, ""},
			tC{"is_mod", "boolean", 0, false, false, "0"},
			tC{"is_admin", "boolean", 0, false, false, "0"},
			tC{"is_banned", "boolean", 0, false, false, "0"},
			tC{"user_count", "int", 0, false, false, "0"}, // TODO: Implement this

			tC{"tag", "varchar", 50, false, false, "''"},
		},
		[]tblKey{
			tblKey{"gid", "primary", "", false},
		},
	)

	createTable("users_groups_promotions", mysqlPre, mysqlCol,
		[]tC{
			tC{"pid", "int", 0, false, true, ""},
			tC{"from_gid", "int", 0, false, false, ""},
			tC{"to_gid", "int", 0, false, false, ""},
			tC{"two_way", "boolean", 0, false, false, "0"}, // If a user no longer meets the requirements for this promotion then they will be demoted if this flag is set

			// Requirements
			tC{"level", "int", 0, false, false, ""},
			tC{"posts", "int", 0, false, false, "0"},
			tC{"minTime", "int", 0, false, false, ""},        // How long someone needs to have been in their current group before being promoted
			tC{"registeredFor", "int", 0, false, false, "0"}, // minutes
		},
		[]tblKey{
			tblKey{"pid", "primary", "", false},
		},
	)

	/*
		createTable("users_groups_promotions_scheduled","","",
			[]tC{
				tC{"prid","int",0,false,false,""},
				tC{"uid","int",0,false,false,""},
				tC{"runAt","datetime",0,false,false,""},
			},
			[]tblKey{
				// TODO: Test to see that the compound primary key works
				tblKey{"prid,uid", "primary", "", false},
			},
		)
	*/

	createTable("users_2fa_keys", mysqlPre, mysqlCol,
		[]tC{
			tC{"uid", "int", 0, false, false, ""},
			tC{"secret", "varchar", 100, false, false, ""},
			tC{"scratch1", "varchar", 50, false, false, ""},
			tC{"scratch2", "varchar", 50, false, false, ""},
			tC{"scratch3", "varchar", 50, false, false, ""},
			tC{"scratch4", "varchar", 50, false, false, ""},
			tC{"scratch5", "varchar", 50, false, false, ""},
			tC{"scratch6", "varchar", 50, false, false, ""},
			tC{"scratch7", "varchar", 50, false, false, ""},
			tC{"scratch8", "varchar", 50, false, false, ""},
			tC{"createdAt", "createdAt", 0, false, false, ""},
		},
		[]tblKey{
			tblKey{"uid", "primary", "", false},
		},
	)

	// What should we do about global penalties? Put them on the users table for speed? Or keep them here?
	// Should we add IP Penalties? No, that's a stupid idea, just implement IP Bans properly. What about shadowbans?
	// TODO: Perm overrides
	// TODO: Add a mod-queue and other basic auto-mod features. This is needed for awaiting activation and the mod_queue penalty flag
	// TODO: Add a penalty type where a user is stopped from creating plugin_guilds social groups
	// TODO: Shadow bans. We will probably have a CanShadowBan permission for this, as we *really* don't want people using this lightly.
	/*createTable("users_penalties","","",
		[]tC{
			tC{"uid","int",0,false,false,""},
			tC{"element_id","int",0,false,false,""},
			tC{"element_type","varchar",50,false,false,""}, //forum, profile?, and social_group. Leave blank for global.
			tC{"overrides","text",0,false,false,"{}"},

			tC{"mod_queue","boolean",0,false,false,"0"},
			tC{"shadow_ban","boolean",0,false,false,"0"},
			tC{"no_avatar","boolean",0,false,false,"0"}, // Coming Soon. Should this be a perm override instead?

			// Do we *really* need rate-limit penalty types? Are we going to be allowing bots or something?
			//tC{"posts_per_hour","int",0,false,false,"0"},
			//tC{"topics_per_hour","int",0,false,false,"0"},
			//tC{"posts_count","int",0,false,false,"0"},
			//tC{"topic_count","int",0,false,false,"0"},
			//tC{"last_hour","int",0,false,false,"0"}, // UNIX Time, as we don't need to do anything too fancy here. When an hour has elapsed since that time, reset the hourly penalty counters.

			tC{"issued_by","int",0,false,false,""},
			tC{"issued_at","createdAt",0,false,false,""},
			tC{"expires_at","datetime",0,false,false,""},
		}, nil,
	)*/

	createTable("users_groups_scheduler", "", "",
		[]tC{
			tC{"uid", "int", 0, false, false, ""},
			tC{"set_group", "int", 0, false, false, ""},

			tC{"issued_by", "int", 0, false, false, ""},
			tC{"issued_at", "createdAt", 0, false, false, ""},
			tC{"revert_at", "datetime", 0, false, false, ""},
			tC{"temporary", "boolean", 0, false, false, ""}, // special case for permanent bans to do the necessary bookkeeping, might be removed in the future
		},
		[]tblKey{
			tblKey{"uid", "primary", "", false},
		},
	)

	// TODO: Can we use a piece of software dedicated to persistent queues for this rather than relying on the database for it?
	createTable("users_avatar_queue", "", "",
		[]tC{
			tC{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
		},
		[]tblKey{
			tblKey{"uid", "primary", "", false},
		},
	)

	// TODO: Should we add a users prefix to this table to fit the "unofficial convention"?
	// TODO: Add an autoincrement key?
	createTable("emails", "", "",
		[]tC{
			tC{"email", "varchar", 200, false, false, ""},
			tC{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"validated", "boolean", 0, false, false, "0"},
			tC{"token", "varchar", 200, false, false, "''"},
		}, nil,
	)

	// TODO: Allow for patterns in domains, if the bots try to shake things up there?
	/*
		createTable("email_domain_blacklist", "", "",
			[]tC{
				tC{"domain", "varchar", 200, false, false, ""},
				tC{"gtld", "boolean", 0, false, false, "0"},
			},
			[]tblKey{
				tblKey{"domain", "primary"},
			},
		)
	*/

	// TODO: Implement password resets
	createTable("password_resets", "", "",
		[]tC{
			tC{"email", "varchar", 200, false, false, ""},
			tC{"uid", "int", 0, false, false, ""},             // TODO: Make this a foreign key
			tC{"validated", "varchar", 200, false, false, ""}, // Token given once the one-use token is consumed, used to prevent multiple people consuming the same one-use token
			tC{"token", "varchar", 200, false, false, ""},
			tC{"createdAt", "createdAt", 0, false, false, ""},
		}, nil,
	)

	createTable("forums", mysqlPre, mysqlCol,
		[]tC{
			tC{"fid", "int", 0, false, true, ""},
			tC{"name", "varchar", 100, false, false, ""},
			tC{"desc", "varchar", 200, false, false, ""},
			tC{"tmpl", "varchar", 200, false, false, "''"},
			tC{"active", "boolean", 0, false, false, "1"},
			tC{"order", "int", 0, false, false, "0"},
			tC{"topicCount", "int", 0, false, false, "0"},
			tC{"preset", "varchar", 100, false, false, "''"},
			tC{"parentID", "int", 0, false, false, "0"},
			tC{"parentType", "varchar", 50, false, false, "''"},
			tC{"lastTopicID", "int", 0, false, false, "0"},
			tC{"lastReplyerID", "int", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"fid", "primary", "", false},
		},
	)

	createTable("forums_permissions", "", "",
		[]tC{
			tC{"fid", "int", 0, false, false, ""},
			tC{"gid", "int", 0, false, false, ""},
			tC{"preset", "varchar", 100, false, false, "''"},
			tC{"permissions", "text", 0, false, false, ""},
		},
		[]tblKey{
			// TODO: Test to see that the compound primary key works
			tblKey{"fid,gid", "primary", "", false},
		},
	)

	createTable("topics", mysqlPre, mysqlCol,
		[]tC{
			tC{"tid", "int", 0, false, true, ""},
			tC{"title", "varchar", 100, false, false, ""}, // TODO: Increase the max length to 200?
			tC{"content", "text", 0, false, false, ""},
			tC{"parsed_content", "text", 0, false, false, ""},
			tC{"createdAt", "createdAt", 0, false, false, ""},
			tC{"lastReplyAt", "datetime", 0, false, false, ""},
			tC{"lastReplyBy", "int", 0, false, false, ""},
			tC{"lastReplyID", "int", 0, false, false, "0"},
			tC{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"is_closed", "boolean", 0, false, false, "0"},
			tC{"sticky", "boolean", 0, false, false, "0"},
			// TODO: Add an index for this
			tC{"parentID", "int", 0, false, false, "2"},
			tC{"ip", "varchar", 200, false, false, "''"},
			tC{"postCount", "int", 0, false, false, "1"},
			tC{"likeCount", "int", 0, false, false, "0"},
			tC{"attachCount", "int", 0, false, false, "0"},
			tC{"words", "int", 0, false, false, "0"},
			tC{"views", "int", 0, false, false, "0"},
			//tC{"dailyViews", "int", 0, false, false, "0"},
			//tC{"weeklyViews", "int", 0, false, false, "0"},
			//tC{"monthlyViews", "int", 0, false, false, "0"},
			// ? - A little hacky, maybe we could do something less likely to bite us with huge numbers of topics?
			// TODO: Add an index for this?
			//tC{"lastMonth", "datetime", 0, false, false, ""},
			tC{"css_class", "varchar", 100, false, false, "''"},
			tC{"poll", "int", 0, false, false, "0"},
			tC{"data", "varchar", 200, false, false, "''"},
		},
		[]tblKey{
			tblKey{"tid", "primary", "", false},
			tblKey{"content", "fulltext", "", false},
		},
	)

	createTable("replies", mysqlPre, mysqlCol,
		[]tC{
			tC{"rid", "int", 0, false, true, ""},  // TODO: Rename to replyID?
			tC{"tid", "int", 0, false, false, ""}, // TODO: Rename to topicID?
			tC{"content", "text", 0, false, false, ""},
			tC{"parsed_content", "text", 0, false, false, ""},
			tC{"createdAt", "createdAt", 0, false, false, ""},
			tC{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"lastEdit", "int", 0, false, false, "0"},
			tC{"lastEditBy", "int", 0, false, false, "0"},
			tC{"lastUpdated", "datetime", 0, false, false, ""},
			tC{"ip", "varchar", 200, false, false, "''"},
			tC{"likeCount", "int", 0, false, false, "0"},
			tC{"attachCount", "int", 0, false, false, "0"},
			tC{"words", "int", 0, false, false, "1"}, // ? - replies has a default of 1 and topics has 0? why?
			tC{"actionType", "varchar", 20, false, false, "''"},
			tC{"poll", "int", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"rid", "primary", "", false},
			tblKey{"content", "fulltext", "", false},
		},
	)

	createTable("attachments", mysqlPre, mysqlCol,
		[]tC{
			tC{"attachID", "int", 0, false, true, ""},
			tC{"sectionID", "int", 0, false, false, "0"},
			tC{"sectionTable", "varchar", 200, false, false, "forums"},
			tC{"originID", "int", 0, false, false, ""},
			tC{"originTable", "varchar", 200, false, false, "replies"},
			tC{"uploadedBy", "int", 0, false, false, ""}, // TODO; Make this a foreign key
			tC{"path", "varchar", 200, false, false, ""},
			tC{"extra", "varchar", 200, false, false, ""},
		},
		[]tblKey{
			tblKey{"attachID", "primary", "", false},
		},
	)

	createTable("revisions", mysqlPre, mysqlCol,
		[]tC{
			tC{"reviseID", "int", 0, false, true, ""},
			tC{"content", "text", 0, false, false, ""},
			tC{"contentID", "int", 0, false, false, ""},
			tC{"contentType", "varchar", 100, false, false, "replies"},
			tC{"createdAt", "createdAt", 0, false, false, ""},
			// TODO: Add a createdBy column?
		},
		[]tblKey{
			tblKey{"reviseID", "primary", "", false},
		},
	)

	createTable("polls", mysqlPre, mysqlCol,
		[]tC{
			tC{"pollID", "int", 0, false, true, ""},
			tC{"parentID", "int", 0, false, false, "0"},
			tC{"parentTable", "varchar", 100, false, false, "topics"}, // topics, replies
			tC{"type", "int", 0, false, false, "0"},
			tC{"options", "json", 0, false, false, ""},
			tC{"votes", "int", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"pollID", "primary", "", false},
		},
	)

	createTable("polls_options", "", "",
		[]tC{
			tC{"pollID", "int", 0, false, false, ""},
			tC{"option", "int", 0, false, false, "0"},
			tC{"votes", "int", 0, false, false, "0"},
		}, nil,
	)

	createTable("polls_votes", mysqlPre, mysqlCol,
		[]tC{
			tC{"pollID", "int", 0, false, false, ""},
			tC{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"option", "int", 0, false, false, "0"},
			tC{"castAt", "createdAt", 0, false, false, ""},
			tC{"ip", "varchar", 200, false, false, "''"},
		}, nil,
	)

	createTable("users_replies", mysqlPre, mysqlCol,
		[]tC{
			tC{"rid", "int", 0, false, true, ""},
			tC{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"content", "text", 0, false, false, ""},
			tC{"parsed_content", "text", 0, false, false, ""},
			tC{"createdAt", "createdAt", 0, false, false, ""},
			tC{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"lastEdit", "int", 0, false, false, "0"},
			tC{"lastEditBy", "int", 0, false, false, "0"},
			tC{"ip", "varchar", 200, false, false, "''"},
		},
		[]tblKey{
			tblKey{"rid", "primary", "", false},
		},
	)

	createTable("likes", "", "",
		[]tC{
			tC{"weight", "tinyint", 0, false, false, "1"},
			tC{"targetItem", "int", 0, false, false, ""},
			tC{"targetType", "varchar", 50, false, false, "replies"},
			tC{"sentBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"createdAt", "createdAt", 0, false, false, ""},
			tC{"recalc", "tinyint", 0, false, false, "0"},
		}, nil,
	)

	//columns("participants, createdBy, createdAt, lastReplyBy, lastReplyAt").Where("cid = ?")
	createTable("conversations", "", "",
		[]tC{
			tC{"cid", "int", 0, false, true, ""},
			tC{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"createdAt", "createdAt", 0, false, false, ""},
			tC{"lastReplyAt", "datetime", 0, false, false, ""},
			tC{"lastReplyBy", "int", 0, false, false, ""},
		},
		[]tblKey{
			tblKey{"cid", "primary", "", false},
		},
	)

	createTable("conversations_posts", "", "",
		[]tC{
			tC{"pid", "int", 0, false, true, ""},
			tC{"cid", "int", 0, false, false, ""},
			tC{"createdBy", "int", 0, false, false, ""},
			tC{"body", "varchar", 50, false, false, ""},
			tC{"post", "varchar", 50, false, false, "''"},
		},
		[]tblKey{
			tblKey{"pid", "primary", "", false},
		},
	)

	createTable("conversations_participants", "", "",
		[]tC{
			tC{"uid", "int", 0, false, false, ""},
			tC{"cid", "int", 0, false, false, ""},
		}, nil,
	)

	/*
		createTable("users_friends", "", "",
			[]tC{
				tC{"uid", "int", 0, false, false, ""},
				tC{"uid2", "int", 0, false, false, ""},
			}, nil,
		)
		createTable("users_friends_invites", "", "",
			[]tC{
				tC{"requester", "int", 0, false, false, ""},
				tC{"target", "int", 0, false, false, ""},
			}, nil,
		)
	*/

	createTable("users_blocks", "", "",
		[]tC{
			tC{"blocker", "int", 0, false, false, ""},
			tC{"blockedUser", "int", 0, false, false, ""},
		}, nil,
	)

	createTable("activity_stream_matches", "", "",
		[]tC{
			tC{"watcher", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"asid", "int", 0, false, false, ""},    // TODO: Make this a foreign key
		},
		[]tblKey{
			tblKey{"asid,asid", "foreign", "activity_stream", true},
		},
	)

	createTable("activity_stream", "", "",
		[]tC{
			tC{"asid", "int", 0, false, true, ""},
			tC{"actor", "int", 0, false, false, ""},            /* the one doing the act */ // TODO: Make this a foreign key
			tC{"targetUser", "int", 0, false, false, ""},       /* the user who created the item the actor is acting on, some items like forums may lack a targetUser field */
			tC{"event", "varchar", 50, false, false, ""},       /* mention, like, reply (as in the act of replying to an item, not the reply item type, you can "reply" to a forum by making a topic in it), friend_invite */
			tC{"elementType", "varchar", 50, false, false, ""}, /* topic, post (calling it post here to differentiate it from the 'reply' event), forum, user */
			tC{"elementID", "int", 0, false, false, ""},        /* the ID of the element being acted upon */
			tC{"createdAt", "createdAt", 0, false, false, ""},
			tC{"extra", "varchar", 200, false, false, "''"},
		},
		[]tblKey{
			tblKey{"asid", "primary", "", false},
		},
	)

	createTable("activity_subscriptions", "", "",
		[]tC{
			tC{"user", "int", 0, false, false, ""},            // TODO: Make this a foreign key
			tC{"targetID", "int", 0, false, false, ""},        /* the ID of the element being acted upon */
			tC{"targetType", "varchar", 50, false, false, ""}, /* topic, post (calling it post here to differentiate it from the 'reply' event), forum, user */
			tC{"level", "int", 0, false, false, "0"},          /* 0: Mentions (aka the global default for any post), 1: Replies To You, 2: All Replies*/
		}, nil,
	)

	/* Due to MySQL's design, we have to drop the unique keys for table settings, plugins, and themes down from 200 to 180 or it will error */
	createTable("settings", "", "",
		[]tC{
			tC{"name", "varchar", 180, false, false, ""},
			tC{"content", "varchar", 250, false, false, ""},
			tC{"type", "varchar", 50, false, false, ""},
			tC{"constraints", "varchar", 200, false, false, "''"},
		},
		[]tblKey{
			tblKey{"name", "unique", "", false},
		},
	)

	createTable("word_filters", "", "",
		[]tC{
			tC{"wfid", "int", 0, false, true, ""},
			tC{"find", "varchar", 200, false, false, ""},
			tC{"replacement", "varchar", 200, false, false, ""},
		},
		[]tblKey{
			tblKey{"wfid", "primary", "", false},
		},
	)

	createTable("plugins", "", "",
		[]tC{
			tC{"uname", "varchar", 180, false, false, ""},
			tC{"active", "boolean", 0, false, false, "0"},
			tC{"installed", "boolean", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"uname", "unique", "", false},
		},
	)

	createTable("themes", "", "",
		[]tC{
			tC{"uname", "varchar", 180, false, false, ""},
			tC{"default", "boolean", 0, false, false, "0"},
			//tC{"profileUserVars", "text", 0, false, false, "''"},
		},
		[]tblKey{
			tblKey{"uname", "unique", "", false},
		},
	)

	createTable("widgets", "", "",
		[]tC{
			tC{"wid", "int", 0, false, true, ""},
			tC{"position", "int", 0, false, false, ""},
			tC{"side", "varchar", 100, false, false, ""},
			tC{"type", "varchar", 100, false, false, ""},
			tC{"active", "boolean", 0, false, false, "0"},
			tC{"location", "varchar", 100, false, false, ""},
			tC{"data", "text", 0, false, false, "''"},
		},
		[]tblKey{
			tblKey{"wid", "primary", "", false},
		},
	)

	createTable("menus", "", "",
		[]tC{
			tC{"mid", "int", 0, false, true, ""},
		},
		[]tblKey{
			tblKey{"mid", "primary", "", false},
		},
	)

	createTable("menu_items", "", "",
		[]tC{
			tC{"miid", "int", 0, false, true, ""},
			tC{"mid", "int", 0, false, false, ""},
			tC{"name", "varchar", 200, false, false, "''"},
			tC{"htmlID", "varchar", 200, false, false, "''"},
			tC{"cssClass", "varchar", 200, false, false, "''"},
			tC{"position", "varchar", 100, false, false, ""},
			tC{"path", "varchar", 200, false, false, "''"},
			tC{"aria", "varchar", 200, false, false, "''"},
			tC{"tooltip", "varchar", 200, false, false, "''"},
			tC{"tmplName", "varchar", 200, false, false, "''"},
			tC{"order", "int", 0, false, false, "0"},

			tC{"guestOnly", "boolean", 0, false, false, "0"},
			tC{"memberOnly", "boolean", 0, false, false, "0"},
			tC{"staffOnly", "boolean", 0, false, false, "0"},
			tC{"adminOnly", "boolean", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"miid", "primary", "", false},
		},
	)

	createTable("pages", mysqlPre, mysqlCol,
		[]tC{
			tC{"pid", "int", 0, false, true, ""},
			//tC{"path", "varchar", 200, false, false, ""},
			tC{"name", "varchar", 200, false, false, ""},
			tC{"title", "varchar", 200, false, false, ""},
			tC{"body", "text", 0, false, false, ""},
			// TODO: Make this a table?
			tC{"allowedGroups", "text", 0, false, false, ""},
			tC{"menuID", "int", 0, false, false, "-1"}, // simple sidebar menu
		},
		[]tblKey{
			tblKey{"pid", "primary", "", false},
		},
	)

	createTable("registration_logs", "", "",
		[]tC{
			tC{"rlid", "int", 0, false, true, ""},
			tC{"username", "varchar", 100, false, false, ""},
			tC{"email", "varchar", 100, false, false, ""},
			tC{"failureReason", "varchar", 100, false, false, ""},
			tC{"success", "bool", 0, false, false, "0"}, // Did this attempt succeed?
			tC{"ipaddress", "varchar", 200, false, false, ""},
			tC{"doneAt", "createdAt", 0, false, false, ""},
		},
		[]tblKey{
			tblKey{"rlid", "primary", "", false},
		},
	)

	createTable("login_logs", "", "",
		[]tC{
			tC{"lid", "int", 0, false, true, ""},
			tC{"uid", "int", 0, false, false, ""},
			tC{"success", "bool", 0, false, false, "0"}, // Did this attempt succeed?
			tC{"ipaddress", "varchar", 200, false, false, ""},
			tC{"doneAt", "createdAt", 0, false, false, ""},
		},
		[]tblKey{
			tblKey{"lid", "primary", "", false},
		},
	)

	createTable("moderation_logs", "", "",
		[]tC{
			tC{"action", "varchar", 100, false, false, ""},
			tC{"elementID", "int", 0, false, false, ""},
			tC{"elementType", "varchar", 100, false, false, ""},
			tC{"ipaddress", "varchar", 200, false, false, ""},
			tC{"actorID", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"doneAt", "datetime", 0, false, false, ""},
			tC{"extra", "text", 0, false, false, ""},
		}, nil,
	)

	createTable("administration_logs", "", "",
		[]tC{
			tC{"action", "varchar", 100, false, false, ""},
			tC{"elementID", "int", 0, false, false, ""},
			tC{"elementType", "varchar", 100, false, false, ""},
			tC{"ipaddress", "varchar", 200, false, false, ""},
			tC{"actorID", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"doneAt", "datetime", 0, false, false, ""},
			tC{"extra", "text", 0, false, false, ""},
		}, nil,
	)

	createTable("viewchunks", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			tC{"route", "varchar", 200, false, false, ""}, // todo: set a default empty here
		}, nil,
	)

	createTable("viewchunks_agents", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			tC{"browser", "varchar", 200, false, false, ""}, // googlebot, firefox, opera, etc.
			//tC{"version","varchar",0,false,false,""}, // the version of the browser or bot
		}, nil,
	)

	createTable("viewchunks_systems", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			tC{"system", "varchar", 200, false, false, ""}, // windows, android, unknown, etc.
		}, nil,
	)

	createTable("viewchunks_langs", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			tC{"lang", "varchar", 200, false, false, ""}, // en, ru, etc.
		}, nil,
	)

	createTable("viewchunks_referrers", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			tC{"domain", "varchar", 200, false, false, ""},
		}, nil,
	)

	createTable("viewchunks_forums", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			tC{"forum", "int", 0, false, false, ""},
		}, nil,
	)

	createTable("topicchunks", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			// TODO: Add a column for the parent forum?
		}, nil,
	)

	createTable("postchunks", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			// TODO: Add a column for the parent topic / profile?
		}, nil,
	)

	createTable("memchunks", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"stack", "int", 0, false, false, "0"},
			tC{"heap", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
		}, nil,
	)

	createTable("sync", "", "",
		[]tC{
			tC{"last_update", "datetime", 0, false, false, ""},
		}, nil,
	)

	createTable("updates", "", "",
		[]tC{
			tC{"dbVersion", "int", 0, false, false, "0"},
		}, nil,
	)

	createTable("meta", "", "",
		[]tC{
			tC{"name", "varchar", 200, false, false, ""},
			tC{"value", "varchar", 200, false, false, ""},
		}, nil,
	)

	return err
}
