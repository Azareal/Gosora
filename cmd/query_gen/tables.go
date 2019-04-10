package main

import "github.com/Azareal/Gosora/query_gen"

var mysqlPre = "utf8mb4"
var mysqlCol = "utf8mb4_general_ci"

type tblColumn = qgen.DBTableColumn
type tblKey = qgen.DBTableKey

func createTables(adapter qgen.Adapter) error {
	qgen.Install.CreateTable("users", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"uid", "int", 0, false, true, ""},
			tblColumn{"name", "varchar", 100, false, false, ""},
			tblColumn{"password", "varchar", 100, false, false, ""},

			tblColumn{"salt", "varchar", 80, false, false, "''"},
			tblColumn{"group", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"active", "boolean", 0, false, false, "0"},
			tblColumn{"is_super_admin", "boolean", 0, false, false, "0"},
			tblColumn{"createdAt", "createdAt", 0, false, false, ""},
			tblColumn{"lastActiveAt", "datetime", 0, false, false, ""},
			tblColumn{"session", "varchar", 200, false, false, "''"},
			//tblColumn{"authToken", "varchar", 200, false, false, "''"},
			tblColumn{"last_ip", "varchar", 200, false, false, "0.0.0.0.0"},
			tblColumn{"email", "varchar", 200, false, false, "''"},
			tblColumn{"avatar", "varchar", 100, false, false, "''"},
			tblColumn{"message", "text", 0, false, false, "''"},
			tblColumn{"url_prefix", "varchar", 20, false, false, "''"},
			tblColumn{"url_name", "varchar", 100, false, false, "''"},
			tblColumn{"level", "smallint", 0, false, false, "0"},
			tblColumn{"score", "int", 0, false, false, "0"},
			tblColumn{"posts", "int", 0, false, false, "0"},
			tblColumn{"bigposts", "int", 0, false, false, "0"},
			tblColumn{"megaposts", "int", 0, false, false, "0"},
			tblColumn{"topics", "int", 0, false, false, "0"},
			tblColumn{"liked", "int", 0, false, false, "0"},

			// These two are to bound liked queries with little bits of information we know about the user to reduce the server load
			tblColumn{"oldestItemLikedCreatedAt", "datetime", 0, false, false, ""}, // For internal use only, semantics may change
			tblColumn{"lastLiked", "datetime", 0, false, false, ""},                // For internal use only, semantics may change

			//tblColumn{"penalty_count","int",0,false,false,"0"},
			tblColumn{"temp_group", "int", 0, false, false, "0"}, // For temporary groups, set this to zero when a temporary group isn't in effect
		},
		[]tblKey{
			tblKey{"uid", "primary"},
			tblKey{"name", "unique"},
		},
	)

	qgen.Install.CreateTable("users_groups", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"gid", "int", 0, false, true, ""},
			tblColumn{"name", "varchar", 100, false, false, ""},
			tblColumn{"permissions", "text", 0, false, false, ""},
			tblColumn{"plugin_perms", "text", 0, false, false, ""},
			tblColumn{"is_mod", "boolean", 0, false, false, "0"},
			tblColumn{"is_admin", "boolean", 0, false, false, "0"},
			tblColumn{"is_banned", "boolean", 0, false, false, "0"},
			tblColumn{"user_count", "int", 0, false, false, "0"}, // TODO: Implement this

			tblColumn{"tag", "varchar", 50, false, false, "''"},
		},
		[]tblKey{
			tblKey{"gid", "primary"},
		},
	)

	qgen.Install.CreateTable("users_2fa_keys", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"uid", "int", 0, false, false, ""},
			tblColumn{"secret", "varchar", 100, false, false, ""},
			tblColumn{"scratch1", "varchar", 50, false, false, ""},
			tblColumn{"scratch2", "varchar", 50, false, false, ""},
			tblColumn{"scratch3", "varchar", 50, false, false, ""},
			tblColumn{"scratch4", "varchar", 50, false, false, ""},
			tblColumn{"scratch5", "varchar", 50, false, false, ""},
			tblColumn{"scratch6", "varchar", 50, false, false, ""},
			tblColumn{"scratch7", "varchar", 50, false, false, ""},
			tblColumn{"scratch8", "varchar", 50, false, false, ""},
			tblColumn{"createdAt", "createdAt", 0, false, false, ""},
		},
		[]tblKey{
			tblKey{"uid", "primary"},
		},
	)

	// What should we do about global penalties? Put them on the users table for speed? Or keep them here?
	// Should we add IP Penalties? No, that's a stupid idea, just implement IP Bans properly. What about shadowbans?
	// TODO: Perm overrides
	// TODO: Add a mod-queue and other basic auto-mod features. This is needed for awaiting activation and the mod_queue penalty flag
	// TODO: Add a penalty type where a user is stopped from creating plugin_guilds social groups
	// TODO: Shadow bans. We will probably have a CanShadowBan permission for this, as we *really* don't want people using this lightly.
	/*qgen.Install.CreateTable("users_penalties","","",
		[]tblColumn{
			tblColumn{"uid","int",0,false,false,""},
			tblColumn{"element_id","int",0,false,false,""},
			tblColumn{"element_type","varchar",50,false,false,""}, //forum, profile?, and social_group. Leave blank for global.
			tblColumn{"overrides","text",0,false,false,"{}"},

			tblColumn{"mod_queue","boolean",0,false,false,"0"},
			tblColumn{"shadow_ban","boolean",0,false,false,"0"},
			tblColumn{"no_avatar","boolean",0,false,false,"0"}, // Coming Soon. Should this be a perm override instead?

			// Do we *really* need rate-limit penalty types? Are we going to be allowing bots or something?
			//tblColumn{"posts_per_hour","int",0,false,false,"0"},
			//tblColumn{"topics_per_hour","int",0,false,false,"0"},
			//tblColumn{"posts_count","int",0,false,false,"0"},
			//tblColumn{"topic_count","int",0,false,false,"0"},
			//tblColumn{"last_hour","int",0,false,false,"0"}, // UNIX Time, as we don't need to do anything too fancy here. When an hour has elapsed since that time, reset the hourly penalty counters.

			tblColumn{"issued_by","int",0,false,false,""},
			tblColumn{"issued_at","createdAt",0,false,false,""},
			tblColumn{"expires_at","datetime",0,false,false,""},
		}, nil,
	)*/

	qgen.Install.CreateTable("users_groups_scheduler", "", "",
		[]tblColumn{
			tblColumn{"uid", "int", 0, false, false, ""},
			tblColumn{"set_group", "int", 0, false, false, ""},

			tblColumn{"issued_by", "int", 0, false, false, ""},
			tblColumn{"issued_at", "createdAt", 0, false, false, ""},
			tblColumn{"revert_at", "datetime", 0, false, false, ""},
			tblColumn{"temporary", "boolean", 0, false, false, ""}, // special case for permanent bans to do the necessary bookkeeping, might be removed in the future
		},
		[]tblKey{
			tblKey{"uid", "primary"},
		},
	)

	// TODO: Can we use a piece of software dedicated to persistent queues for this rather than relying on the database for it?
	qgen.Install.CreateTable("users_avatar_queue", "", "",
		[]tblColumn{
			tblColumn{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
		},
		[]tblKey{
			tblKey{"uid", "primary"},
		},
	)

	// TODO: Should we add a users prefix to this table to fit the "unofficial convention"?
	qgen.Install.CreateTable("emails", "", "",
		[]tblColumn{
			tblColumn{"email", "varchar", 200, false, false, ""},
			tblColumn{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"validated", "boolean", 0, false, false, "0"},
			tblColumn{"token", "varchar", 200, false, false, "''"},
		}, nil,
	)

	// TODO: Allow for patterns in domains, if the bots try to shake things up there?
	/*
		qgen.Install.CreateTable("email_domain_blacklist", "", "",
			[]tblColumn{
				tblColumn{"domain", "varchar", 200, false, false, ""},
				tblColumn{"gtld", "boolean", 0, false, false, "0"},
			},
			[]tblKey{
				tblKey{"domain", "primary"},
			},
		)
	*/

	// TODO: Implement password resets
	qgen.Install.CreateTable("password_resets", "", "",
		[]tblColumn{
			tblColumn{"email", "varchar", 200, false, false, ""},
			tblColumn{"uid", "int", 0, false, false, ""},             // TODO: Make this a foreign key
			tblColumn{"validated", "varchar", 200, false, false, ""}, // Token given once the one-use token is consumed, used to prevent multiple people consuming the same one-use token
			tblColumn{"token", "varchar", 200, false, false, ""},
			tblColumn{"createdAt", "createdAt", 0, false, false, ""},
		}, nil,
	)

	qgen.Install.CreateTable("forums", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"fid", "int", 0, false, true, ""},
			tblColumn{"name", "varchar", 100, false, false, ""},
			tblColumn{"desc", "varchar", 200, false, false, ""},
			tblColumn{"active", "boolean", 0, false, false, "1"},
			tblColumn{"topicCount", "int", 0, false, false, "0"},
			tblColumn{"preset", "varchar", 100, false, false, "''"},
			tblColumn{"parentID", "int", 0, false, false, "0"},
			tblColumn{"parentType", "varchar", 50, false, false, "''"},
			tblColumn{"lastTopicID", "int", 0, false, false, "0"},
			tblColumn{"lastReplyerID", "int", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"fid", "primary"},
		},
	)

	qgen.Install.CreateTable("forums_permissions", "", "",
		[]tblColumn{
			tblColumn{"fid", "int", 0, false, false, ""},
			tblColumn{"gid", "int", 0, false, false, ""},
			tblColumn{"preset", "varchar", 100, false, false, "''"},
			tblColumn{"permissions", "text", 0, false, false, ""},
		},
		[]tblKey{
			// TODO: Test to see that the compound primary key works
			tblKey{"fid,gid", "primary"},
		},
	)

	qgen.Install.CreateTable("topics", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"tid", "int", 0, false, true, ""},
			tblColumn{"title", "varchar", 100, false, false, ""}, // TODO: Increase the max length to 200?
			tblColumn{"content", "text", 0, false, false, ""},
			tblColumn{"parsed_content", "text", 0, false, false, ""},
			tblColumn{"createdAt", "createdAt", 0, false, false, ""},
			tblColumn{"lastReplyAt", "datetime", 0, false, false, ""},
			tblColumn{"lastReplyBy", "int", 0, false, false, ""},
			tblColumn{"lastReplyID", "int", 0, false, false, "0"},
			tblColumn{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"is_closed", "boolean", 0, false, false, "0"},
			tblColumn{"sticky", "boolean", 0, false, false, "0"},
			// TODO: Add an index for this
			tblColumn{"parentID", "int", 0, false, false, "2"},
			tblColumn{"ipaddress", "varchar", 200, false, false, "0.0.0.0.0"},
			tblColumn{"postCount", "int", 0, false, false, "1"},
			tblColumn{"likeCount", "int", 0, false, false, "0"},
			tblColumn{"attachCount", "int", 0, false, false, "0"},
			tblColumn{"words", "int", 0, false, false, "0"},
			tblColumn{"views", "int", 0, false, false, "0"},
			//tblColumn{"dailyViews", "int", 0, false, false, "0"},
			//tblColumn{"weeklyViews", "int", 0, false, false, "0"},
			//tblColumn{"monthlyViews", "int", 0, false, false, "0"},
			// ? - A little hacky, maybe we could do something less likely to bite us with huge numbers of topics?
			// TODO: Add an index for this?
			//tblColumn{"lastMonth", "datetime", 0, false, false, ""},
			tblColumn{"css_class", "varchar", 100, false, false, "''"},
			tblColumn{"poll", "int", 0, false, false, "0"},
			tblColumn{"data", "varchar", 200, false, false, "''"},
		},
		[]tblKey{
			tblKey{"tid", "primary"},
			tblKey{"content", "fulltext"},
		},
	)

	qgen.Install.CreateTable("replies", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"rid", "int", 0, false, true, ""},  // TODO: Rename to replyID?
			tblColumn{"tid", "int", 0, false, false, ""}, // TODO: Rename to topicID?
			tblColumn{"content", "text", 0, false, false, ""},
			tblColumn{"parsed_content", "text", 0, false, false, ""},
			tblColumn{"createdAt", "createdAt", 0, false, false, ""},
			tblColumn{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"lastEdit", "int", 0, false, false, "0"},
			tblColumn{"lastEditBy", "int", 0, false, false, "0"},
			tblColumn{"lastUpdated", "datetime", 0, false, false, ""},
			tblColumn{"ipaddress", "varchar", 200, false, false, "0.0.0.0.0"},
			tblColumn{"likeCount", "int", 0, false, false, "0"},
			tblColumn{"attachCount", "int", 0, false, false, "0"},
			tblColumn{"words", "int", 0, false, false, "1"}, // ? - replies has a default of 1 and topics has 0? why?
			tblColumn{"actionType", "varchar", 20, false, false, "''"},
			tblColumn{"poll", "int", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"rid", "primary"},
			tblKey{"content", "fulltext"},
		},
	)

	qgen.Install.CreateTable("attachments", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"attachID", "int", 0, false, true, ""},
			tblColumn{"sectionID", "int", 0, false, false, "0"},
			tblColumn{"sectionTable", "varchar", 200, false, false, "forums"},
			tblColumn{"originID", "int", 0, false, false, ""},
			tblColumn{"originTable", "varchar", 200, false, false, "replies"},
			tblColumn{"uploadedBy", "int", 0, false, false, ""}, // TODO; Make this a foreign key
			tblColumn{"path", "varchar", 200, false, false, ""},
			//tblColumn{"extra", "varchar", 200, false, false, ""},
		},
		[]tblKey{
			tblKey{"attachID", "primary"},
		},
	)

	qgen.Install.CreateTable("revisions", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"reviseID", "int", 0, false, true, ""},
			tblColumn{"content", "text", 0, false, false, ""},
			tblColumn{"contentID", "int", 0, false, false, ""},
			tblColumn{"contentType", "varchar", 100, false, false, "replies"},
			tblColumn{"createdAt", "createdAt", 0, false, false, ""},
			// TODO: Add a createdBy column?
		},
		[]tblKey{
			tblKey{"reviseID", "primary"},
		},
	)

	qgen.Install.CreateTable("polls", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"pollID", "int", 0, false, true, ""},
			tblColumn{"parentID", "int", 0, false, false, "0"},
			tblColumn{"parentTable", "varchar", 100, false, false, "topics"}, // topics, replies
			tblColumn{"type", "int", 0, false, false, "0"},
			tblColumn{"options", "json", 0, false, false, ""},
			tblColumn{"votes", "int", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"pollID", "primary"},
		},
	)

	qgen.Install.CreateTable("polls_options", "", "",
		[]tblColumn{
			tblColumn{"pollID", "int", 0, false, false, ""},
			tblColumn{"option", "int", 0, false, false, "0"},
			tblColumn{"votes", "int", 0, false, false, "0"},
		}, nil,
	)

	qgen.Install.CreateTable("polls_votes", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"pollID", "int", 0, false, false, ""},
			tblColumn{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"option", "int", 0, false, false, "0"},
			tblColumn{"castAt", "createdAt", 0, false, false, ""},
			tblColumn{"ipaddress", "varchar", 200, false, false, "0.0.0.0.0"},
		}, nil,
	)

	qgen.Install.CreateTable("users_replies", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"rid", "int", 0, false, true, ""},
			tblColumn{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"content", "text", 0, false, false, ""},
			tblColumn{"parsed_content", "text", 0, false, false, ""},
			tblColumn{"createdAt", "createdAt", 0, false, false, ""},
			tblColumn{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"lastEdit", "int", 0, false, false, "0"},
			tblColumn{"lastEditBy", "int", 0, false, false, "0"},
			tblColumn{"ipaddress", "varchar", 200, false, false, "0.0.0.0.0"},
		},
		[]tblKey{
			tblKey{"rid", "primary"},
		},
	)

	qgen.Install.CreateTable("likes", "", "",
		[]tblColumn{
			tblColumn{"weight", "tinyint", 0, false, false, "1"},
			tblColumn{"targetItem", "int", 0, false, false, ""},
			tblColumn{"targetType", "varchar", 50, false, false, "replies"},
			tblColumn{"sentBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"createdAt", "createdAt", 0, false, false, ""},
			tblColumn{"recalc", "tinyint", 0, false, false, "0"},
		}, nil,
	)

	qgen.Install.CreateTable("activity_stream_matches", "", "",
		[]tblColumn{
			tblColumn{"watcher", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"asid", "int", 0, false, false, ""},    // TODO: Make this a foreign key
		}, nil,
	)

	qgen.Install.CreateTable("activity_stream", "", "",
		[]tblColumn{
			tblColumn{"asid", "int", 0, false, true, ""},
			tblColumn{"actor", "int", 0, false, false, ""},            /* the one doing the act */ // TODO: Make this a foreign key
			tblColumn{"targetUser", "int", 0, false, false, ""},       /* the user who created the item the actor is acting on, some items like forums may lack a targetUser field */
			tblColumn{"event", "varchar", 50, false, false, ""},       /* mention, like, reply (as in the act of replying to an item, not the reply item type, you can "reply" to a forum by making a topic in it), friend_invite */
			tblColumn{"elementType", "varchar", 50, false, false, ""}, /* topic, post (calling it post here to differentiate it from the 'reply' event), forum, user */
			tblColumn{"elementID", "int", 0, false, false, ""},        /* the ID of the element being acted upon */
		},
		[]tblKey{
			tblKey{"asid", "primary"},
		},
	)

	qgen.Install.CreateTable("activity_subscriptions", "", "",
		[]tblColumn{
			tblColumn{"user", "int", 0, false, false, ""},            // TODO: Make this a foreign key
			tblColumn{"targetID", "int", 0, false, false, ""},        /* the ID of the element being acted upon */
			tblColumn{"targetType", "varchar", 50, false, false, ""}, /* topic, post (calling it post here to differentiate it from the 'reply' event), forum, user */
			tblColumn{"level", "int", 0, false, false, "0"},          /* 0: Mentions (aka the global default for any post), 1: Replies To You, 2: All Replies*/
		}, nil,
	)

	/* Due to MySQL's design, we have to drop the unique keys for table settings, plugins, and themes down from 200 to 180 or it will error */
	qgen.Install.CreateTable("settings", "", "",
		[]tblColumn{
			tblColumn{"name", "varchar", 180, false, false, ""},
			tblColumn{"content", "varchar", 250, false, false, ""},
			tblColumn{"type", "varchar", 50, false, false, ""},
			tblColumn{"constraints", "varchar", 200, false, false, "''"},
		},
		[]tblKey{
			tblKey{"name", "unique"},
		},
	)

	qgen.Install.CreateTable("word_filters", "", "",
		[]tblColumn{
			tblColumn{"wfid", "int", 0, false, true, ""},
			tblColumn{"find", "varchar", 200, false, false, ""},
			tblColumn{"replacement", "varchar", 200, false, false, ""},
		},
		[]tblKey{
			tblKey{"wfid", "primary"},
		},
	)

	qgen.Install.CreateTable("plugins", "", "",
		[]tblColumn{
			tblColumn{"uname", "varchar", 180, false, false, ""},
			tblColumn{"active", "boolean", 0, false, false, "0"},
			tblColumn{"installed", "boolean", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"uname", "unique"},
		},
	)

	qgen.Install.CreateTable("themes", "", "",
		[]tblColumn{
			tblColumn{"uname", "varchar", 180, false, false, ""},
			tblColumn{"default", "boolean", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"uname", "unique"},
		},
	)

	qgen.Install.CreateTable("widgets", "", "",
		[]tblColumn{
			tblColumn{"wid", "int", 0, false, true, ""},
			tblColumn{"position", "int", 0, false, false, ""},
			tblColumn{"side", "varchar", 100, false, false, ""},
			tblColumn{"type", "varchar", 100, false, false, ""},
			tblColumn{"active", "boolean", 0, false, false, "0"},
			tblColumn{"location", "varchar", 100, false, false, ""},
			tblColumn{"data", "text", 0, false, false, "''"},
		},
		[]tblKey{
			tblKey{"wid", "primary"},
		},
	)

	qgen.Install.CreateTable("menus", "", "",
		[]tblColumn{
			tblColumn{"mid", "int", 0, false, true, ""},
		},
		[]tblKey{
			tblKey{"mid", "primary"},
		},
	)

	qgen.Install.CreateTable("menu_items", "", "",
		[]tblColumn{
			tblColumn{"miid", "int", 0, false, true, ""},
			tblColumn{"mid", "int", 0, false, false, ""},
			tblColumn{"name", "varchar", 200, false, false, "''"},
			tblColumn{"htmlID", "varchar", 200, false, false, "''"},
			tblColumn{"cssClass", "varchar", 200, false, false, "''"},
			tblColumn{"position", "varchar", 100, false, false, ""},
			tblColumn{"path", "varchar", 200, false, false, "''"},
			tblColumn{"aria", "varchar", 200, false, false, "''"},
			tblColumn{"tooltip", "varchar", 200, false, false, "''"},
			tblColumn{"tmplName", "varchar", 200, false, false, "''"},
			tblColumn{"order", "int", 0, false, false, "0"},

			tblColumn{"guestOnly", "boolean", 0, false, false, "0"},
			tblColumn{"memberOnly", "boolean", 0, false, false, "0"},
			tblColumn{"staffOnly", "boolean", 0, false, false, "0"},
			tblColumn{"adminOnly", "boolean", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"miid", "primary"},
		},
	)

	qgen.Install.CreateTable("pages", mysqlPre, mysqlCol,
		[]tblColumn{
			tblColumn{"pid", "int", 0, false, true, ""},
			//tblColumn{"path", "varchar", 200, false, false, ""},
			tblColumn{"name", "varchar", 200, false, false, ""},
			tblColumn{"title", "varchar", 200, false, false, ""},
			tblColumn{"body", "text", 0, false, false, ""},
			// TODO: Make this a table?
			tblColumn{"allowedGroups", "text", 0, false, false, ""},
			tblColumn{"menuID", "int", 0, false, false, "-1"}, // simple sidebar menu
		},
		[]tblKey{
			tblKey{"pid", "primary"},
		},
	)

	qgen.Install.CreateTable("registration_logs", "", "",
		[]tblColumn{
			tblColumn{"rlid", "int", 0, false, true, ""},
			tblColumn{"username", "varchar", 100, false, false, ""},
			tblColumn{"email", "varchar", 100, false, false, ""},
			tblColumn{"failureReason", "varchar", 100, false, false, ""},
			tblColumn{"success", "bool", 0, false, false, "0"}, // Did this attempt succeed?
			tblColumn{"ipaddress", "varchar", 200, false, false, ""},
			tblColumn{"doneAt", "createdAt", 0, false, false, ""},
		},
		[]tblKey{
			tblKey{"rlid", "primary"},
		},
	)

	qgen.Install.CreateTable("login_logs", "", "",
		[]tblColumn{
			tblColumn{"lid", "int", 0, false, true, ""},
			tblColumn{"uid", "int", 0, false, false, ""},
			tblColumn{"success", "bool", 0, false, false, "0"}, // Did this attempt succeed?
			tblColumn{"ipaddress", "varchar", 200, false, false, ""},
			tblColumn{"doneAt", "createdAt", 0, false, false, ""},
		},
		[]tblKey{
			tblKey{"lid", "primary"},
		},
	)

	qgen.Install.CreateTable("moderation_logs", "", "",
		[]tblColumn{
			tblColumn{"action", "varchar", 100, false, false, ""},
			tblColumn{"elementID", "int", 0, false, false, ""},
			tblColumn{"elementType", "varchar", 100, false, false, ""},
			tblColumn{"ipaddress", "varchar", 200, false, false, ""},
			tblColumn{"actorID", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"doneAt", "datetime", 0, false, false, ""},
		}, nil,
	)

	qgen.Install.CreateTable("administration_logs", "", "",
		[]tblColumn{
			tblColumn{"action", "varchar", 100, false, false, ""},
			tblColumn{"elementID", "int", 0, false, false, ""},
			tblColumn{"elementType", "varchar", 100, false, false, ""},
			tblColumn{"ipaddress", "varchar", 200, false, false, ""},
			tblColumn{"actorID", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			tblColumn{"doneAt", "datetime", 0, false, false, ""},
		}, nil,
	)

	qgen.Install.CreateTable("viewchunks", "", "",
		[]tblColumn{
			tblColumn{"count", "int", 0, false, false, "0"},
			tblColumn{"createdAt", "datetime", 0, false, false, ""},
			tblColumn{"route", "varchar", 200, false, false, ""},
		}, nil,
	)

	qgen.Install.CreateTable("viewchunks_agents", "", "",
		[]tblColumn{
			tblColumn{"count", "int", 0, false, false, "0"},
			tblColumn{"createdAt", "datetime", 0, false, false, ""},
			tblColumn{"browser", "varchar", 200, false, false, ""}, // googlebot, firefox, opera, etc.
			//tblColumn{"version","varchar",0,false,false,""}, // the version of the browser or bot
		}, nil,
	)

	qgen.Install.CreateTable("viewchunks_systems", "", "",
		[]tblColumn{
			tblColumn{"count", "int", 0, false, false, "0"},
			tblColumn{"createdAt", "datetime", 0, false, false, ""},
			tblColumn{"system", "varchar", 200, false, false, ""}, // windows, android, unknown, etc.
		}, nil,
	)

	qgen.Install.CreateTable("viewchunks_langs", "", "",
		[]tblColumn{
			tblColumn{"count", "int", 0, false, false, "0"},
			tblColumn{"createdAt", "datetime", 0, false, false, ""},
			tblColumn{"lang", "varchar", 200, false, false, ""}, // en, ru, etc.
		}, nil,
	)

	qgen.Install.CreateTable("viewchunks_referrers", "", "",
		[]tblColumn{
			tblColumn{"count", "int", 0, false, false, "0"},
			tblColumn{"createdAt", "datetime", 0, false, false, ""},
			tblColumn{"domain", "varchar", 200, false, false, ""},
		}, nil,
	)

	qgen.Install.CreateTable("viewchunks_forums", "", "",
		[]tblColumn{
			tblColumn{"count", "int", 0, false, false, "0"},
			tblColumn{"createdAt", "datetime", 0, false, false, ""},
			tblColumn{"forum", "int", 0, false, false, ""},
		}, nil,
	)

	qgen.Install.CreateTable("topicchunks", "", "",
		[]tblColumn{
			tblColumn{"count", "int", 0, false, false, "0"},
			tblColumn{"createdAt", "datetime", 0, false, false, ""},
			// TODO: Add a column for the parent forum?
		}, nil,
	)

	qgen.Install.CreateTable("postchunks", "", "",
		[]tblColumn{
			tblColumn{"count", "int", 0, false, false, "0"},
			tblColumn{"createdAt", "datetime", 0, false, false, ""},
			// TODO: Add a column for the parent topic / profile?
		}, nil,
	)

	qgen.Install.CreateTable("sync", "", "",
		[]tblColumn{
			tblColumn{"last_update", "datetime", 0, false, false, ""},
		}, nil,
	)

	qgen.Install.CreateTable("updates", "", "",
		[]tblColumn{
			tblColumn{"dbVersion", "int", 0, false, false, "0"},
		}, nil,
	)

	return nil
}
