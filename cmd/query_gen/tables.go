package main

import qgen "github.com/Azareal/Gosora/query_gen"

var mysqlPre = "utf8mb4"
var mysqlCol = "utf8mb4_general_ci"

var tables []string

type tblColumn = qgen.DBTableColumn
type tC = tblColumn
type tblKey = qgen.DBTableKey

func createTables(a qgen.Adapter) error {
	tables = nil
	f := func(table, charset, collation string, cols []tC, keys []tblKey) error {
		tables = append(tables, table)
		return qgen.Install.CreateTable(table, charset, collation, cols, keys)
	}
	return createTables2(a, f)
}

func createTables2(a qgen.Adapter, f func(table, charset, collation string, columns []tC, keys []tblKey) error) (err error) {
	createTable := func(table, charset, collation string, cols []tC, keys []tblKey) {
		if err != nil {
			return
		}
		err = f(table, charset, collation, cols, keys)
	}
	bcol := func(col string, val bool) qgen.DBTableColumn {
		if val {
			return tC{col, "boolean", 0, false, false, "1"}
		}
		return tC{col, "boolean", 0, false, false, "0"}
	}
	ccol := func(col string, size int, sdefault string) qgen.DBTableColumn {
		return tC{col, "varchar", size, false, false, sdefault}
	}
	text := func(params ...string) qgen.DBTableColumn {
		if len(params) == 0 {
			return tC{"", "text", 0, false, false, ""}
		}
		col, sdefault := params[0], ""
		if len(params) > 1 {
			sdefault = params[1]
			if sdefault == "" {
				sdefault = "''"
			}
		}
		return tC{col, "text", 0, false, false, sdefault}
	}
	createdAt := func(coll ...string) qgen.DBTableColumn {
		var col string
		if len(coll) > 0 {
			col = coll[0]
		}
		if col == "" {
			col = "createdAt"
		}
		return tC{col, "createdAt", 0, false, false, ""}
	}

	createTable("users", mysqlPre, mysqlCol,
		[]tC{
			tC{"uid", "int", 0, false, true, ""},
			ccol("name", 100, ""),
			ccol("password", 100, ""),

			ccol("salt", 80, "''"),
			tC{"group", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			bcol("active", false),
			bcol("is_super_admin", false),
			createdAt(),
			tC{"lastActiveAt", "datetime", 0, false, false, ""},
			ccol("session", 200, "''"),
			//ccol("authToken", 200, "''"),
			ccol("last_ip", 200, "''"),
			tC{"profile_comments", "int", 0, false, false, "0"},
			tC{"who_can_convo", "int", 0, false, false, "0"},
			tC{"enable_embeds", "int", 0, false, false, "-1"},
			ccol("email", 200, "''"),
			ccol("avatar", 100, "''"),
			text("message"),

			// TODO: Drop these columns?
			ccol("url_prefix", 20, "''"),
			ccol("url_name", 100, "''"),
			//text("pub_key"),

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
		[]tK{
			tK{"uid", "primary", "", false},
			tK{"name", "unique", "", false},
		},
	)

	createTable("users_groups", mysqlPre, mysqlCol,
		[]tC{
			tC{"gid", "int", 0, false, true, ""},
			ccol("name", 100, ""),
			text("permissions"),
			text("plugin_perms"),
			bcol("is_mod", false),
			bcol("is_admin", false),
			bcol("is_banned", false),
			tC{"user_count", "int", 0, false, false, "0"}, // TODO: Implement this

			ccol("tag", 50, "''"),
		},
		[]tK{
			tK{"gid", "primary", "", false},
		},
	)

	createTable("users_groups_promotions", mysqlPre, mysqlCol,
		[]tC{
			tC{"pid", "int", 0, false, true, ""},
			tC{"from_gid", "int", 0, false, false, ""},
			tC{"to_gid", "int", 0, false, false, ""},
			bcol("two_way", false), // If a user no longer meets the requirements for this promotion then they will be demoted if this flag is set

			// Requirements
			tC{"level", "int", 0, false, false, ""},
			tC{"posts", "int", 0, false, false, "0"},
			tC{"minTime", "int", 0, false, false, ""},        // How long someone needs to have been in their current group before being promoted
			tC{"registeredFor", "int", 0, false, false, "0"}, // minutes
		},
		[]tK{
			tK{"pid", "primary", "", false},
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
			ccol("secret", 100, ""),
			ccol("scratch1", 50, ""),
			ccol("scratch2", 50, ""),
			ccol("scratch3", 50, ""),
			ccol("scratch4", 50, ""),
			ccol("scratch5", 50, ""),
			ccol("scratch6", 50, ""),
			ccol("scratch7", 50, ""),
			ccol("scratch8", 50, ""),
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
			ccol("element_type",50,""), //forum, profile?, and social_group. Leave blank for global.
			text("overrides","{}"),

			bcol("mod_queue",false),
			bcol("shadow_ban",false),
			bcol("no_avatar",false), // Coming Soon. Should this be a perm override instead?

			// Do we *really* need rate-limit penalty types? Are we going to be allowing bots or something?
			//tC{"posts_per_hour","int",0,false,false,"0"},
			//tC{"topics_per_hour","int",0,false,false,"0"},
			//tC{"posts_count","int",0,false,false,"0"},
			//tC{"topic_count","int",0,false,false,"0"},
			//tC{"last_hour","int",0,false,false,"0"}, // UNIX Time, as we don't need to do anything too fancy here. When an hour has elapsed since that time, reset the hourly penalty counters.

			tC{"issued_by","int",0,false,false,""},
			createdAt("issued_at"),
			tC{"expires_at","datetime",0,false,false,""},
		}, nil,
	)*/

	createTable("users_groups_scheduler", "", "",
		[]tC{
			tC{"uid", "int", 0, false, false, ""},
			tC{"set_group", "int", 0, false, false, ""},

			tC{"issued_by", "int", 0, false, false, ""},
			createdAt("issued_at"),
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
			ccol("email", 200, ""),
			tC{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			bcol("validated", false),
			ccol("token", 200, "''"),
		}, nil,
	)

	// TODO: Allow for patterns in domains, if the bots try to shake things up there?
	/*
		createTable("email_domain_blacklist", "", "",
			[]tC{
				ccol("domain", 200, ""),
				bcol("gtld", false),
			},
			[]tK{
				tK{"domain", "primary"},
			},
		)
	*/

	// TODO: Implement password resets
	createTable("password_resets", "", "",
		[]tC{
			ccol("email", 200, ""),
			tC{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			ccol("validated", 200, ""),            // Token given once the one-use token is consumed, used to prevent multiple people consuming the same one-use token
			ccol("token", 200, ""),
			createdAt(),
		}, nil,
	)

	createTable("forums", mysqlPre, mysqlCol,
		[]tC{
			tC{"fid", "int", 0, false, true, ""},
			ccol("name", 100, ""),
			ccol("desc", 200, ""),
			ccol("tmpl", 200, "''"),
			bcol("active", true),
			tC{"order", "int", 0, false, false, "0"},
			tC{"topicCount", "int", 0, false, false, "0"},
			ccol("preset", 100, "''"),
			tC{"parentID", "int", 0, false, false, "0"},
			ccol("parentType", 50, "''"),
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
			ccol("preset", 100, "''"),
			text("permissions", "{}"),
		},
		[]tblKey{
			// TODO: Test to see that the compound primary key works
			tblKey{"fid,gid", "primary", "", false},
		},
	)

	createTable("topics", mysqlPre, mysqlCol,
		[]tC{
			tC{"tid", "int", 0, false, true, ""},
			ccol("title", 100, ""), // TODO: Increase the max length to 200?
			text("content"),
			text("parsed_content"),
			createdAt(),
			tC{"lastReplyAt", "datetime", 0, false, false, ""},
			tC{"lastReplyBy", "int", 0, false, false, ""},
			tC{"lastReplyID", "int", 0, false, false, "0"},
			tC{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			bcol("is_closed", false),
			bcol("sticky", false),
			// TODO: Add an index for this
			tC{"parentID", "int", 0, false, false, "2"},
			ccol("ip", 200, "''"),
			tC{"postCount", "int", 0, false, false, "1"},
			tC{"likeCount", "int", 0, false, false, "0"},
			tC{"attachCount", "int", 0, false, false, "0"},
			tC{"words", "int", 0, false, false, "0"},
			tC{"views", "int", 0, false, false, "0"},
			//tC{"dayViews", "int", 0, false, false, "0"},
			tC{"weekEvenViews", "int", 0, false, false, "0"},
			tC{"weekOddViews", "int", 0, false, false, "0"},
			///tC{"weekViews", "int", 0, false, false, "0"},
			///tC{"lastWeekViews", "int", 0, false, false, "0"},
			//tC{"monthViews", "int", 0, false, false, "0"},
			// ? - A little hacky, maybe we could do something less likely to bite us with huge numbers of topics?
			// TODO: Add an index for this?
			//tC{"lastMonth", "datetime", 0, false, false, ""},
			ccol("css_class", 100, "''"),
			tC{"poll", "int", 0, false, false, "0"},
			ccol("data", 200, "''"),
		},
		[]tK{
			tK{"tid", "primary", "", false},
			tK{"title", "fulltext", "", false},
			tK{"content", "fulltext", "", false},
		},
	)

	createTable("replies", mysqlPre, mysqlCol,
		[]tC{
			tC{"rid", "int", 0, false, true, ""},  // TODO: Rename to replyID?
			tC{"tid", "int", 0, false, false, ""}, // TODO: Rename to topicID?
			text("content"),
			text("parsed_content"),
			createdAt(),
			tC{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"lastEdit", "int", 0, false, false, "0"},
			tC{"lastEditBy", "int", 0, false, false, "0"},
			tC{"lastUpdated", "datetime", 0, false, false, ""},
			ccol("ip", 200, "''"),
			tC{"likeCount", "int", 0, false, false, "0"},
			tC{"attachCount", "int", 0, false, false, "0"},
			tC{"words", "int", 0, false, false, "1"}, // ? - replies has a default of 1 and topics has 0? why?
			ccol("actionType", 20, "''"),
			tC{"poll", "int", 0, false, false, "0"},
		},
		[]tK{
			tK{"rid", "primary", "", false},
			tK{"content", "fulltext", "", false},
		},
	)

	createTable("attachments", mysqlPre, mysqlCol,
		[]tC{
			tC{"attachID", "int", 0, false, true, ""},
			tC{"sectionID", "int", 0, false, false, "0"},
			ccol("sectionTable", 200, "forums"),
			tC{"originID", "int", 0, false, false, ""},
			ccol("originTable", 200, "replies"),
			tC{"uploadedBy", "int", 0, false, false, ""}, // TODO; Make this a foreign key
			ccol("path", 200, ""),
			ccol("extra", 200, ""),
		},
		[]tblKey{
			tblKey{"attachID", "primary", "", false},
		},
	)

	createTable("revisions", mysqlPre, mysqlCol,
		[]tC{
			tC{"reviseID", "int", 0, false, true, ""},
			text("content"),
			tC{"contentID", "int", 0, false, false, ""},
			ccol("contentType", 100, "replies"),
			createdAt(),
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
			ccol("parentTable", 100, "topics"), // topics, replies
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
			createdAt("castAt"),
			ccol("ip", 200, "''"),
		}, nil,
	)

	createTable("users_replies", mysqlPre, mysqlCol,
		[]tC{
			tC{"rid", "int", 0, false, true, ""},
			tC{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			text("content"),
			text("parsed_content"),
			createdAt(),
			tC{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"lastEdit", "int", 0, false, false, "0"},
			tC{"lastEditBy", "int", 0, false, false, "0"},
			ccol("ip", 200, "''"),
		},
		[]tblKey{
			tblKey{"rid", "primary", "", false},
		},
	)

	createTable("likes", "", "",
		[]tC{
			tC{"weight", "tinyint", 0, false, false, "1"},
			tC{"targetItem", "int", 0, false, false, ""},
			ccol("targetType", 50, "replies"),
			tC{"sentBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			createdAt(),
			tC{"recalc", "tinyint", 0, false, false, "0"},
		}, nil,
	)

	//columns("participants, createdBy, createdAt, lastReplyBy, lastReplyAt").Where("cid = ?")
	createTable("conversations", "", "",
		[]tC{
			tC{"cid", "int", 0, false, true, ""},
			tC{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			createdAt(),
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
			ccol("body", 50, ""),
			ccol("post", 50, "''"),
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
			tC{"actor", "int", 0, false, false, ""},      /* the one doing the act */ // TODO: Make this a foreign key
			tC{"targetUser", "int", 0, false, false, ""}, /* the user who created the item the actor is acting on, some items like forums may lack a targetUser field */
			ccol("event", 50, ""),                        /* mention, like, reply (as in the act of replying to an item, not the reply item type, you can "reply" to a forum by making a topic in it), friend_invite */
			ccol("elementType", 50, ""),                  /* topic, post (calling it post here to differentiate it from the 'reply' event), forum, user */

			// replacement for elementType
			tC{"elementTable", "int", 0, false, false, "0"},

			tC{"elementID", "int", 0, false, false, ""}, /* the ID of the element being acted upon */
			createdAt(),
			ccol("extra", 200, "''"),
		},
		[]tK{
			tK{"asid", "primary", "", false},
		},
	)

	createTable("activity_subscriptions", "", "",
		[]tC{
			tC{"user", "int", 0, false, false, ""},     // TODO: Make this a foreign key
			tC{"targetID", "int", 0, false, false, ""}, /* the ID of the element being acted upon */
			ccol("targetType", 50, ""),                 /* topic, post (calling it post here to differentiate it from the 'reply' event), forum, user */
			tC{"level", "int", 0, false, false, "0"},   /* 0: Mentions (aka the global default for any post), 1: Replies To You, 2: All Replies*/
		}, nil,
	)

	/* Due to MySQL's design, we have to drop the unique keys for table settings, plugins, and themes down from 200 to 180 or it will error */
	createTable("settings", "", "",
		[]tC{
			ccol("name", 180, ""),
			ccol("content", 250, ""),
			ccol("type", 50, ""),
			ccol("constraints", 200, "''"),
		},
		[]tblKey{
			tblKey{"name", "unique", "", false},
		},
	)

	createTable("word_filters", "", "",
		[]tC{
			tC{"wfid", "int", 0, false, true, ""},
			ccol("find", 200, ""),
			ccol("replacement", 200, ""),
		},
		[]tblKey{
			tblKey{"wfid", "primary", "", false},
		},
	)

	createTable("plugins", "", "",
		[]tC{
			ccol("uname", 180, ""),
			bcol("active", false),
			bcol("installed", false),
		},
		[]tblKey{
			tblKey{"uname", "unique", "", false},
		},
	)

	createTable("themes", "", "",
		[]tC{
			ccol("uname", 180, ""),
			bcol("default", false),
			//text("profileUserVars"),
		},
		[]tblKey{
			tblKey{"uname", "unique", "", false},
		},
	)

	createTable("widgets", "", "",
		[]tC{
			tC{"wid", "int", 0, false, true, ""},
			tC{"position", "int", 0, false, false, ""},
			ccol("side", 100, ""),
			ccol("type", 100, ""),
			bcol("active", false),
			ccol("location", 100, ""),
			text("data"),
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
			ccol("name", 200, "''"),
			ccol("htmlID", 200, "''"),
			ccol("cssClass", 200, "''"),
			ccol("position", 100, ""),
			ccol("path", 200, "''"),
			ccol("aria", 200, "''"),
			ccol("tooltip", 200, "''"),
			ccol("tmplName", 200, "''"),
			tC{"order", "int", 0, false, false, "0"},

			bcol("guestOnly", false),
			bcol("memberOnly", false),
			bcol("staffOnly", false),
			bcol("adminOnly", false),
		},
		[]tK{
			tK{"miid", "primary", "", false},
		},
	)

	createTable("pages", mysqlPre, mysqlCol,
		[]tC{
			tC{"pid", "int", 0, false, true, ""},
			//ccol("path", 200, ""),
			ccol("name", 200, ""),
			ccol("title", 200, ""),
			text("body"),
			// TODO: Make this a table?
			text("allowedGroups"),
			tC{"menuID", "int", 0, false, false, "-1"}, // simple sidebar menu
		},
		[]tK{
			tK{"pid", "primary", "", false},
		},
	)

	createTable("registration_logs", "", "",
		[]tC{
			tC{"rlid", "int", 0, false, true, ""},
			ccol("username", 100, ""),
			tC{"email", "varchar", 100, false, false, ""},
			ccol("failureReason", 100, ""),
			bcol("success", false), // Did this attempt succeed?
			ccol("ipaddress", 200, ""),
			createdAt("doneAt"),
		},
		[]tblKey{
			tblKey{"rlid", "primary", "", false},
		},
	)

	createTable("login_logs", "", "",
		[]tC{
			tC{"lid", "int", 0, false, true, ""},
			tC{"uid", "int", 0, false, false, ""},
			bcol("success", false), // Did this attempt succeed?
			ccol("ipaddress", 200, ""),
			createdAt("doneAt"),
		},
		[]tblKey{
			tblKey{"lid", "primary", "", false},
		},
	)

	createTable("moderation_logs", "", "",
		[]tC{
			ccol("action", 100, ""),
			tC{"elementID", "int", 0, false, false, ""},
			ccol("elementType", 100, ""),
			ccol("ipaddress", 200, ""),
			tC{"actorID", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"doneAt", "datetime", 0, false, false, ""},
			text("extra"),
		}, nil,
	)

	createTable("administration_logs", "", "",
		[]tC{
			ccol("action", 100, ""),
			tC{"elementID", "int", 0, false, false, ""},
			ccol("elementType", 100, ""),
			ccol("ipaddress", 200, ""),
			tC{"actorID", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tC{"doneAt", "datetime", 0, false, false, ""},
			text("extra"),
		}, nil,
	)

	createTable("viewchunks", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"avg", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			ccol("route", 200, ""), // TODO: set a default empty here
		}, nil,
	)

	createTable("viewchunks_agents", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			ccol("browser", 200, ""), // googlebot, firefox, opera, etc.
			//ccol("version",0,""), // the version of the browser or bot
		}, nil,
	)

	createTable("viewchunks_systems", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			ccol("system", 200, ""), // windows, android, unknown, etc.
		}, nil,
	)

	createTable("viewchunks_langs", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			ccol("lang", 200, ""), // en, ru, etc.
		}, nil,
	)

	createTable("viewchunks_referrers", "", "",
		[]tC{
			tC{"count", "int", 0, false, false, "0"},
			tC{"createdAt", "datetime", 0, false, false, ""},
			ccol("domain", 200, ""),
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

	createTable("perfchunks", "", "",
		[]tC{
			tC{"low", "int", 0, false, false, "0"},
			tC{"high", "int", 0, false, false, "0"},
			tC{"avg", "int", 0, false, false, "0"},
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
			ccol("name", 200, ""),
			ccol("value", 200, ""),
		}, nil,
	)

	/*createTable("tables", "", "",
		[]tC{
			tC{"id", "int", 0, false, true, ""},
			ccol("name", 200, ""),
		},
		[]tK{
			tK{"id", "primary", "", false},
			tK{"name", "unique", "", false},
		},
	)*/

	return err
}
