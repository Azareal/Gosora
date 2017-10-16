/* WIP Under Construction */
package main

import "./lib"

func createTables(adapter qgen.DB_Adapter) error {
	qgen.Install.CreateTable("users", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"uid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"name", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"password", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"salt", "varchar", 80, false, false, "''"},
			qgen.DB_Table_Column{"group", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"active", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"is_super_admin", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"createdAt", "createdAt", 0, false, false, ""},
			qgen.DB_Table_Column{"lastActiveAt", "datetime", 0, false, false, ""},
			qgen.DB_Table_Column{"session", "varchar", 200, false, false, "''"},
			qgen.DB_Table_Column{"last_ip", "varchar", 200, false, false, "0.0.0.0.0"},
			qgen.DB_Table_Column{"email", "varchar", 200, false, false, "''"},
			qgen.DB_Table_Column{"avatar", "varchar", 100, false, false, "''"},
			qgen.DB_Table_Column{"message", "text", 0, false, false, "''"},
			qgen.DB_Table_Column{"url_prefix", "varchar", 20, false, false, "''"},
			qgen.DB_Table_Column{"url_name", "varchar", 100, false, false, "''"},
			qgen.DB_Table_Column{"level", "smallint", 0, false, false, "0"},
			qgen.DB_Table_Column{"score", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"posts", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"bigposts", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"megaposts", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"topics", "int", 0, false, false, "0"},
			//qgen.DB_Table_Column{"penalty_count","int",0,false,false,"0"},
			qgen.DB_Table_Column{"temp_group", "int", 0, false, false, "0"}, // For temporary groups, set this to zero when a temporary group isn't in effect
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"uid", "primary"},
			qgen.DB_Table_Key{"name", "unique"},
		},
	)

	qgen.Install.CreateTable("users_groups", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"gid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"name", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"permissions", "text", 0, false, false, ""},
			qgen.DB_Table_Column{"plugin_perms", "text", 0, false, false, ""},
			qgen.DB_Table_Column{"is_mod", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"is_admin", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"is_banned", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"tag", "varchar", 50, false, false, "''"},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"gid", "primary"},
		},
	)

	// What should we do about global penalties? Put them on the users table for speed? Or keep them here?
	// Should we add IP Penalties? No, that's a stupid idea, just implement IP Bans properly. What about shadowbans?
	// TODO: Perm overrides
	// TODO: Add a mod-queue and other basic auto-mod features. This is needed for awaiting activation and the mod_queue penalty flag
	// TODO: Add a penalty type where a user is stopped from creating plugin_socialgroups social groups
	// TODO: Shadow bans. We will probably have a CanShadowBan permission for this, as we *really* don't want people using this lightly.
	/*qgen.Install.CreateTable("users_penalties","","",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"uid","int",0,false,false,""},
			qgen.DB_Table_Column{"element_id","int",0,false,false,""},
			qgen.DB_Table_Column{"element_type","varchar",50,false,false,""}, //forum, profile?, and social_group. Leave blank for global.
			qgen.DB_Table_Column{"overrides","text",0,false,false,"{}"},

			qgen.DB_Table_Column{"mod_queue","boolean",0,false,false,"0"},
			qgen.DB_Table_Column{"shadow_ban","boolean",0,false,false,"0"},
			qgen.DB_Table_Column{"no_avatar","boolean",0,false,false,"0"}, // Coming Soon. Should this be a perm override instead?

			// Do we *really* need rate-limit penalty types? Are we going to be allowing bots or something?
			//qgen.DB_Table_Column{"posts_per_hour","int",0,false,false,"0"},
			//qgen.DB_Table_Column{"topics_per_hour","int",0,false,false,"0"},
			//qgen.DB_Table_Column{"posts_count","int",0,false,false,"0"},
			//qgen.DB_Table_Column{"topic_count","int",0,false,false,"0"},
			//qgen.DB_Table_Column{"last_hour","int",0,false,false,"0"}, // UNIX Time, as we don't need to do anything too fancy here. When an hour has elapsed since that time, reset the hourly penalty counters.

			qgen.DB_Table_Column{"issued_by","int",0,false,false,""},
			qgen.DB_Table_Column{"issued_at","createdAt",0,false,false,""},
			qgen.DB_Table_Column{"expires_at","datetime",0,false,false,""},
		},
		[]qgen.DB_Table_Key{},
	)*/

	qgen.Install.CreateTable("users_groups_scheduler", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"uid", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"set_group", "int", 0, false, false, ""},

			qgen.DB_Table_Column{"issued_by", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"issued_at", "createdAt", 0, false, false, ""},
			qgen.DB_Table_Column{"revert_at", "datetime", 0, false, false, ""},
			qgen.DB_Table_Column{"temporary", "boolean", 0, false, false, ""}, // special case for permanent bans to do the necessary bookkeeping, might be removed in the future
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"uid", "primary"},
		},
	)

	qgen.Install.CreateTable("emails", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"email", "varchar", 200, false, false, ""},
			qgen.DB_Table_Column{"uid", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"validated", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"token", "varchar", 200, false, false, "''"},
		},
		[]qgen.DB_Table_Key{},
	)

	qgen.Install.CreateTable("forums", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"fid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"name", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"desc", "varchar", 200, false, false, ""},
			qgen.DB_Table_Column{"active", "boolean", 0, false, false, "1"},
			qgen.DB_Table_Column{"topicCount", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"preset", "varchar", 100, false, false, "''"},
			qgen.DB_Table_Column{"parentID", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"parentType", "varchar", 50, false, false, "''"},
			qgen.DB_Table_Column{"lastTopicID", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"lastReplyerID", "int", 0, false, false, "0"},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"fid", "primary"},
		},
	)

	qgen.Install.CreateTable("forums_permissions", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"fid", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"gid", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"preset", "varchar", 100, false, false, "''"},
			qgen.DB_Table_Column{"permissions", "text", 0, false, false, ""},
		},
		[]qgen.DB_Table_Key{
			// TODO: Test to see that the compound primary key works
			qgen.DB_Table_Key{"fid,gid", "primary"},
		},
	)

	qgen.Install.CreateTable("topics", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"tid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"title", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"content", "text", 0, false, false, ""},
			qgen.DB_Table_Column{"parsed_content", "text", 0, false, false, ""},
			qgen.DB_Table_Column{"createdAt", "createdAt", 0, false, false, ""},
			qgen.DB_Table_Column{"lastReplyAt", "datetime", 0, false, false, ""},
			qgen.DB_Table_Column{"lastReplyBy", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"createdBy", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"is_closed", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"sticky", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"parentID", "int", 0, false, false, "2"},
			qgen.DB_Table_Column{"ipaddress", "varchar", 200, false, false, "0.0.0.0.0"},
			qgen.DB_Table_Column{"postCount", "int", 0, false, false, "1"},
			qgen.DB_Table_Column{"likeCount", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"words", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"css_class", "varchar", 100, false, false, "''"},
			qgen.DB_Table_Column{"data", "varchar", 200, false, false, "''"},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"tid", "primary"},
		},
	)

	qgen.Install.CreateTable("replies", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"rid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"tid", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"content", "text", 0, false, false, ""},
			qgen.DB_Table_Column{"parsed_content", "text", 0, false, false, ""},
			qgen.DB_Table_Column{"createdAt", "createdAt", 0, false, false, ""},
			qgen.DB_Table_Column{"createdBy", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"lastEdit", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"lastEditBy", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"lastUpdated", "datetime", 0, false, false, ""},
			qgen.DB_Table_Column{"ipaddress", "varchar", 200, false, false, "0.0.0.0.0"},
			qgen.DB_Table_Column{"likeCount", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"words", "int", 0, false, false, "1"}, // ? - replies has a default of 1 and topics has 0? why?
			qgen.DB_Table_Column{"actionType", "varchar", 20, false, false, "''"},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"rid", "primary"},
		},
	)

	qgen.Install.CreateTable("attachments", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"attachID", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"sectionID", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"sectionTable", "varchar", 200, false, false, "forums"},
			qgen.DB_Table_Column{"originID", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"originTable", "varchar", 200, false, false, "replies"},
			qgen.DB_Table_Column{"uploadedBy", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"path", "varchar", 200, false, false, ""},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"attachID", "primary"},
		},
	)

	qgen.Install.CreateTable("revisions", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"index", "int", 0, false, false, ""}, // TODO: Replace this with a proper revision ID x.x
			qgen.DB_Table_Column{"content", "text", 0, false, false, ""},
			qgen.DB_Table_Column{"contentID", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"contentType", "varchar", 100, false, false, "replies"},
		},
		[]qgen.DB_Table_Key{},
	)

	qgen.Install.CreateTable("users_replies", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"rid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"uid", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"content", "text", 0, false, false, ""},
			qgen.DB_Table_Column{"parsed_content", "text", 0, false, false, ""},
			qgen.DB_Table_Column{"createdAt", "createdAt", 0, false, false, ""},
			qgen.DB_Table_Column{"createdBy", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"lastEdit", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"lastEditBy", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"ipaddress", "varchar", 200, false, false, "0.0.0.0.0"},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"rid", "primary"},
		},
	)

	qgen.Install.CreateTable("likes", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"weight", "tinyint", 0, false, false, "1"},
			qgen.DB_Table_Column{"targetItem", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"targetType", "varchar", 50, false, false, "replies"},
			qgen.DB_Table_Column{"sentBy", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"recalc", "tinyint", 0, false, false, "0"},
		},
		[]qgen.DB_Table_Key{},
	)

	qgen.Install.CreateTable("activity_stream_matches", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"watcher", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"asid", "int", 0, false, false, ""},
		},
		[]qgen.DB_Table_Key{},
	)

	qgen.Install.CreateTable("activity_stream", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"asid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"actor", "int", 0, false, false, ""},            /* the one doing the act */
			qgen.DB_Table_Column{"targetUser", "int", 0, false, false, ""},       /* the user who created the item the actor is acting on, some items like forums may lack a targetUser field */
			qgen.DB_Table_Column{"event", "varchar", 50, false, false, ""},       /* mention, like, reply (as in the act of replying to an item, not the reply item type, you can "reply" to a forum by making a topic in it), friend_invite */
			qgen.DB_Table_Column{"elementType", "varchar", 50, false, false, ""}, /* topic, post (calling it post here to differentiate it from the 'reply' event), forum, user */
			qgen.DB_Table_Column{"elementID", "int", 0, false, false, ""},        /* the ID of the element being acted upon */
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"asid", "primary"},
		},
	)

	qgen.Install.CreateTable("activity_subscriptions", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"user", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"targetID", "int", 0, false, false, ""},        /* the ID of the element being acted upon */
			qgen.DB_Table_Column{"targetType", "varchar", 50, false, false, ""}, /* topic, post (calling it post here to differentiate it from the 'reply' event), forum, user */
			qgen.DB_Table_Column{"level", "int", 0, false, false, "0"},          /* 0: Mentions (aka the global default for any post), 1: Replies To You, 2: All Replies*/
		},
		[]qgen.DB_Table_Key{},
	)

	/* Due to MySQL's design, we have to drop the unique keys for table settings, plugins, and themes down from 200 to 180 or it will error */
	qgen.Install.CreateTable("settings", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"name", "varchar", 180, false, false, ""},
			qgen.DB_Table_Column{"content", "varchar", 250, false, false, ""},
			qgen.DB_Table_Column{"type", "varchar", 50, false, false, ""},
			qgen.DB_Table_Column{"constraints", "varchar", 200, false, false, "''"},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"name", "unique"},
		},
	)

	qgen.Install.CreateTable("word_filters", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"wfid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"find", "varchar", 200, false, false, ""},
			qgen.DB_Table_Column{"replacement", "varchar", 200, false, false, ""},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"wfid", "primary"},
		},
	)

	qgen.Install.CreateTable("plugins", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"uname", "varchar", 180, false, false, ""},
			qgen.DB_Table_Column{"active", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"installed", "boolean", 0, false, false, "0"},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"uname", "unique"},
		},
	)

	qgen.Install.CreateTable("themes", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"uname", "varchar", 180, false, false, ""},
			qgen.DB_Table_Column{"default", "boolean", 0, false, false, "0"},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"uname", "unique"},
		},
	)

	qgen.Install.CreateTable("widgets", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"position", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"side", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"type", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"active", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"location", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"data", "text", 0, false, false, "''"},
		},
		[]qgen.DB_Table_Key{},
	)

	qgen.Install.CreateTable("moderation_logs", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"action", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"elementID", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"elementType", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"ipaddress", "varchar", 200, false, false, ""},
			qgen.DB_Table_Column{"actorID", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"doneAt", "datetime", 0, false, false, ""},
		},
		[]qgen.DB_Table_Key{},
	)

	qgen.Install.CreateTable("administration_logs", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"action", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"elementID", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"elementType", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"ipaddress", "varchar", 200, false, false, ""},
			qgen.DB_Table_Column{"actorID", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"doneAt", "datetime", 0, false, false, ""},
		},
		[]qgen.DB_Table_Key{},
	)

	qgen.Install.CreateTable("sync", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"last_update", "datetime", 0, false, false, ""},
		},
		[]qgen.DB_Table_Key{},
	)

	return nil
}
