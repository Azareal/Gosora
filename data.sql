CREATE TABLE `users`(
	`uid` int not null AUTO_INCREMENT,
	`name` varchar(100) not null,
	`password` varchar(100) not null,
	`salt` varchar(80) default '' not null,
	`group` int not null,
	`active` tinyint default 0 not null,
	`is_super_admin` tinyint(1) not null,
	`createdAt` datetime not null,
	`lastActiveAt` datetime not null,
	`session` varchar(200) default '' not null,
	`last_ip` varchar(200) default '0.0.0.0.0' not null,
	`email` varchar(200) default '' not null,
	`avatar` varchar(20) default '' not null,
	`message` text not null,
	`url_prefix` varchar(20) default '' not null,
	`url_name` varchar(100) default '' not null,
	`level` tinyint default 0 not null,
	`score` int default 0 not null,
	`posts` int default 0 not null,
	`bigposts` int default 0 not null,
	`megaposts` int default 0 not null,
	`topics` int default 0 not null,
	primary key(`uid`),
	unique(`name`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE `users_groups`(
	`gid` int not null AUTO_INCREMENT,
	`name` varchar(100) not null,
	`permissions` text not null,
	`is_mod` tinyint DEFAULT 0 not null,
	`is_admin` tinyint DEFAULT 0 not null,
	`is_banned` tinyint DEFAULT 0 not null,
	`tag` varchar(50) DEFAULT '' not null,
	primary key(`gid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE `emails`(
	`email` varchar(200) not null,
	`uid` int not null,
	`validated` tinyint DEFAULT 0 not null,
	`token` varchar(200) DEFAULT '' not null
);

CREATE TABLE `forums`(
	`fid` int not null AUTO_INCREMENT,
	`name` varchar(100) not null,
	`active` tinyint DEFAULT 1 not null,
	`topicCount` int DEFAULT 0 not null,
	`preset` varchar(100) DEFAULT '' not null,
	`lastTopic` varchar(100) DEFAULT '' not null,
	`lastTopicID` int DEFAULT 0 not null,
	`lastReplyer` varchar(100) DEFAULT '' not null,
	`lastReplyerID` int DEFAULT 0 not null,
	`lastTopicTime` datetime not null,
	primary key(`fid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE `forums_permissions`(
	`fid` int not null,
	`gid` int not null,
	`preset` varchar(100) DEFAULT '' not null,
	`permissions` text not null
);

CREATE TABLE `topics`(
	`tid` int not null AUTO_INCREMENT,
	`title` varchar(100) not null,
	`content` text not null,
	`parsed_content` text not null,
	`createdAt` datetime not null,
	`lastReplyAt` datetime not null,
	`createdBy` int not null,
	`is_closed` tinyint DEFAULT 0 not null,
	`sticky` tinyint DEFAULT 0 not null,
	`parentID` int DEFAULT 2 not null,
	`ipaddress` varchar(200) DEFAULT '0.0.0.0.0' not null,
	`postCount` int DEFAULT 1 not null,
	`likeCount` int DEFAULT 0 not null,
	`words` int DEFAULT 0 not null,
	`data` varchar(200) DEFAULT '' not null,
	primary key(`tid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE `replies`(
	`rid` int not null AUTO_INCREMENT,
	`tid` int not null,
	`content` text not null,
	`parsed_content` text not null,
	`createdAt` datetime not null,
	`createdBy` int not null,
	`lastEdit` int not null,
	`lastEditBy` int not null,
	`ipaddress` varchar(200) DEFAULT '0.0.0.0.0' not null,
	`likeCount` int DEFAULT 0 not null,
	`words` int DEFAULT 1 not null,
	primary key(`rid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE `revisions`(
	`index` int not null,
	`content` text not null,
	`contentID` int not null,
	`contentType` varchar(100) DEFAULT 'replies' not null
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE `users_replies`(
	`rid` int not null AUTO_INCREMENT,
	`uid` int not null,
	`content` text not null,
	`parsed_content` text not null,
	`createdAt` datetime not null,
	`createdBy` int not null,
	`lastEdit` int not null,
	`lastEditBy` int not null,
	primary key(`rid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE `likes`(
	`weight` tinyint DEFAULT 1 not null,
	/*`type` tinyint not null, /* Regular Post: 1, Big Post: 2, Mega Post: 3, etc.*/
	`targetItem` int not null,
	`targetType` varchar(50) DEFAULT 'replies' not null,
	`sentBy` int not null,
	`recalc` tinyint DEFAULT 0 not null
);

CREATE TABLE `activity_stream_matches`(
	`watcher` int not null,
	`asid` int not null
);

CREATE TABLE `activity_stream`(
	`asid` int not null AUTO_INCREMENT,
	`actor` int not null, /* the one doing the act */
	`targetUser` int not null, /* the user who created the item the actor is acting on, some items like forums may lack a targetUser field */
	`event` varchar(50) not null, /* mention, like, reply (as in the act of replying to an item, not the reply item type, you can "reply" to a forum by making a topic in it), friend_invite */
	`elementType` varchar(50) not null, /* topic, post (calling it post here to differentiate it from the 'reply' event), forum, user */
	`elementID` int not null, /* the ID of the element being acted upon */
	primary key(`asid`)
);

CREATE TABLE `activity_subscriptions`(
	`user` int not null,
	`targetID` int not null, /* the ID of the element being acted upon */
	`targetType` varchar(50) not null, /* topic, post (calling it post here to differentiate it from the 'reply' event), forum, user */
	`level` tinyint DEFAULT 0 not null /* 0: Mentions (aka the global default for any post), 1: Replies To You, 2: All Replies*/
);

CREATE TABLE `settings`(
	`name` varchar(200) not null,
	`content` varchar(250) not null,
	`type` varchar(50) not null,
	`constraints` varchar(200) DEFAULT '' not null,
	unique(`name`)
);

CREATE TABLE `plugins`(
	`uname` varchar(200) not null,
	`active` tinyint DEFAULT 0 not null,
	unique(`uname`)
);

CREATE TABLE `themes`(
	`uname` varchar(200) not null,
	`default` tinyint DEFAULT 0 not null,
	unique(`uname`)
);

INSERT INTO settings(`name`,`content`,`type`) VALUES ('url_tags','1','bool');
INSERT INTO settings(`name`,`content`,`type`,`constraints`) VALUES ('activation_type','1','list','1-3');
INSERT INTO settings(`name`,`content`,`type`) VALUES ('bigpost_min_chars','250','int');
INSERT INTO settings(`name`,`content`,`type`) VALUES ('megapost_min_chars','1000','int');
INSERT INTO themes(`uname`,`default`) VALUES ('tempra-simple',1);

INSERT INTO users(`name`,`password`,`email`,`group`,`is_super_admin`,`createdAt`,`lastActiveAt`,`message`,`last_ip`) 
VALUES ('Admin','password','admin@localhost',1,1,NOW(),NOW(),'','127.0.0.1');
INSERT INTO emails(`email`,`uid`,`validated`) VALUES ('admin@localhost',1,1);

/*
The Permissions:

BanUsers
ActivateUsers
EditUser
EditUserEmail
EditUserPassword
EditUserGroup
EditUserGroupSuperMod
EditUserGroupAdmin
EditGroup
EditGroupLocalPerms
EditGroupGlobalPerms
EditGroupSuperMod
EditGroupAdmin
ManageForums
EditSettings
ManageThemes
ManagePlugins
ViewIPs

ViewTopic
LikeItem
CreateTopic
EditTopic
DeleteTopic
CreateReply
EditReply
DeleteReply
PinTopic
CloseTopic
*/

INSERT INTO users_groups(`name`,`permissions`,`is_mod`,`is_admin`,`tag`) VALUES ('Administrator','{"BanUsers":true,"ActivateUsers":true,"EditUser":true,"EditUserEmail":true,"EditUserPassword":true,"EditUserGroup":true,"EditUserGroupSuperMod":true,"EditUserGroupAdmin":false,"EditGroup":true,"EditGroupLocalPerms":true,"EditGroupGlobalPerms":true,"EditGroupSuperMod":true,"EditGroupAdmin":false,"ManageForums":true,"EditSettings":true,"ManageThemes":true,"ManagePlugins":true,"ViewIPs":true,"ViewTopic":true,"LikeItem":true,"CreateTopic":true,"EditTopic":true,"DeleteTopic":true,"CreateReply":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true}',1,1,"Admin");
INSERT INTO users_groups(`name`,`permissions`,`is_mod`,`tag`) VALUES ('Moderator','{"BanUsers":true,"ActivateUsers":false,"EditUser":true,"EditUserEmail":false,"EditUserGroup":true,"ViewIPs":true,"ViewTopic":true,"LikeItem":true,"CreateTopic":true,"EditTopic":true,"DeleteTopic":true,"CreateReply":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true}',1,"Mod");
INSERT INTO users_groups(`name`,`permissions`) VALUES ('Member','{"ViewTopic":true,"LikeItem":true,"CreateTopic":true,"CreateReply":true}');
INSERT INTO users_groups(`name`,`permissions`,`is_banned`) VALUES ('Banned','{"ViewTopic":true}',1);
INSERT INTO users_groups(`name`,`permissions`) VALUES ('Awaiting Activation','{"ViewTopic":true}');
INSERT INTO users_groups(`name`,`permissions`,`tag`) VALUES ('Not Loggedin','{"ViewTopic":true}','Guest');

INSERT INTO forums(`name`,`active`) VALUES ('Reports',0);
INSERT INTO forums(`name`,`lastTopicTime`) VALUES ('General',NOW());
INSERT INTO forums_permissions(`gid`,`fid`,`permissions`) VALUES (1,1,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"PinTopic":true,"CloseTopic":true}');
INSERT INTO forums_permissions(`gid`,`fid`,`permissions`) VALUES (2,1,'{"ViewTopic":true,"CreateReply":true,"CloseTopic":true}');
INSERT INTO forums_permissions(`gid`,`fid`,`permissions`) VALUES (3,1,'{}');
INSERT INTO forums_permissions(`gid`,`fid`,`permissions`) VALUES (4,1,'{}');
INSERT INTO forums_permissions(`gid`,`fid`,`permissions`) VALUES (5,1,'{}');
INSERT INTO forums_permissions(`gid`,`fid`,`permissions`) VALUES (6,1,'{}');
INSERT INTO topics(`title`,`content`,`createdAt`,`lastReplyAt`,`createdBy`,`parentID`) 
VALUES ('Test Topic','A topic automatically generated by the software.',NOW(),NOW(),1,2);

INSERT INTO replies(`tid`,`content`,`createdAt`,`createdBy`,`lastEdit`,`lastEditBy`) 
VALUES (1,'Reply 1',NOW(),1,0,0);