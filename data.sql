CREATE TABLE `users`(
	`uid` int not null AUTO_INCREMENT,
	`name` varchar(100) not null,
	`password` varchar(100) not null,
	`salt` varchar(80) DEFAULT '' not null,
	`group` int not null,
	`active` tinyint DEFAULT 0 not null,
	`is_super_admin` tinyint(1) not null,
	`createdAt` datetime not null,
	`lastActiveAt` datetime not null,
	`session` varchar(200) DEFAULT '' not null,
	`last_ip` varchar(200) DEFAULT '0.0.0.0.0' not null,
	`email` varchar(200) DEFAULT '' not null,
	`avatar` varchar(20) DEFAULT '' not null,
	`message` text not null,
	`url_prefix` varchar(20) DEFAULT '' not null,
	`url_name` varchar(100) DEFAULT '' not null,
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
	`lastTopic` varchar(100) DEFAULT '' not null,
	`lastTopicID` int DEFAULT 0 not null,
	`lastReplyer` varchar(100) DEFAULT '' not null,
	`lastReplyerID` int DEFAULT 0 not null,
	`lastTopicTime` datetime not null,
	primary key(`fid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;

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
	`parentID` int DEFAULT 1 not null,
	`ipaddress` varchar(200) DEFAULT '0.0.0.0.0' not null,
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
	primary key(`rid`)
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
	`weight` int DEFAULT 1 not null,
	`type` int not null, /* Regular Post = 1, Big Post = 2, Mega Post = 3, etc.*/
	`targetItem` int not null,
	`sentBy` int not null,
	`recalc` tinyint DEFAULT 0 not null
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
INSERT INTO themes(`uname`,`default`) VALUES ('tempra-simple',1);

INSERT INTO users(`name`,`password`,`email`,`group`,`is_super_admin`,`createdAt`,`lastActiveAt`,`message`) 
VALUES ('Admin','password','admin@localhost',1,1,NOW(),NOW(),'');
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
ManageForums
EditSettings
ManageThemes
ManagePlugins
ViewIPs

ViewTopic
CreateTopic
EditTopic
DeleteTopic
CreateReply
EditReply
DeleteReply
PinTopic
CloseTopic
*/

INSERT INTO users_groups(`name`,`permissions`,`is_mod`,`is_admin`,`tag`) VALUES ('Administrator','{"BanUsers":true,"ActivateUsers":true,"EditUser":true,"EditUserEmail":true,"EditUserPassword":true,"EditUserGroup":true,"EditUserGroupSuperMod":true,"EditUserGroupAdmin":false,"ManageForums":true,"EditSettings":true,"ManageThemes":true,"ManagePlugins":true,"ViewIPs":true,"ViewTopic":true,"CreateTopic":true,"EditTopic":true,"DeleteTopic":true,"CreateReply":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true}',1,1,"Admin");
INSERT INTO users_groups(`name`,`permissions`,`is_mod`,`tag`) VALUES ('Moderator','{"BanUsers":true,"ActivateUsers":false,"EditUser":true,"EditUserEmail":false,"EditUserPassword":false,"EditUserGroup":true,"EditUserGroupSuperMod":false,"EditUserGroupAdmin":false,"ManageForums":false,"EditSettings":false,"ManageThemes":false,"ManagePlugins":false,"ViewIPs":true,"ViewTopic":true,"CreateTopic":true,"EditTopic":true,"DeleteTopic":true,"CreateReply":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true}',1,"Mod");
INSERT INTO users_groups(`name`,`permissions`) VALUES ('Member','{"BanUsers":false,"ActivateUsers":false,"EditUser":false,"EditUserEmail":false,"EditUserPassword":false,"EditUserGroup":false,"EditUserGroupSuperMod":false,"EditUserGroupAdmin":false,"ManageForums":false,"EditSettings":false,"ManageThemes":false,"ManagePlugins":false,"ViewIPs":false,"ViewTopic":true,"CreateTopic":true,"EditTopic":false,"DeleteTopic":false,"CreateReply":true,"EditReply":false,"DeleteReply":false,"PinTopic":false,"CloseTopic":false}');
INSERT INTO users_groups(`name`,`permissions`,`is_banned`) VALUES ('Banned','{"BanUsers":false,"ActivateUsers":false,"EditUser":false,"EditUserEmail":false,"EditUserPassword":false,"EditUserGroup":false,"EditUserGroupSuperMod":false,"EditUserGroupAdmin":false,"ManageForums":false,"EditSettings":false,"ManageThemes":false,"ManagePlugins":false,"ViewIPs":false,"ViewTopic":true,"CreateTopic":false,"EditTopic":false,"DeleteTopic":false,"CreateReply":false,"EditReply":false,"DeleteReply":false,"PinTopic":false,"CloseTopic":false}',1);
INSERT INTO users_groups(`name`,`permissions`) VALUES ('Awaiting Activation','{"BanUsers":false,"ActivateUsers":false,"EditUser":false,"EditUserEmail":false,"EditUserPassword":false,"EditUserGroup":false,"EditUserGroupSuperMod":false,"EditUserGroupAdmin":false,"ManageForums":false,"EditSettings":false,"ManageThemes":false,"ManagePlugins":false,"ViewIPs":false,"ViewTopic":true,"CreateTopic":false,"EditTopic":false,"DeleteTopic":false,"CreateReply":false,"EditReply":false,"DeleteReply":false,"PinTopic":false,"CloseTopic":false}');

INSERT INTO forums(`name`,`lastTopicTime`) VALUES ('General',NOW());
INSERT INTO topics(`title`,`content`,`createdAt`,`lastReplyAt`,`createdBy`,`parentID`) 
VALUES ('Test Topic','A topic automatically generated by the software.',NOW(),NOW(),1,1);

INSERT INTO replies(`tid`,`content`,`createdAt`,`createdBy`,`lastEdit`,`lastEditBy`) 
VALUES (1,'Reply 1',NOW(),1,0,0);