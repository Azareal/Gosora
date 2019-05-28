CREATE TABLE `forums` (
	`fid` int not null AUTO_INCREMENT,
	`name` varchar(100) not null,
	`desc` varchar(200) not null,
	`tmpl` varchar(200) DEFAULT '' not null,
	`active` boolean DEFAULT 1 not null,
	`order` int DEFAULT 0 not null,
	`topicCount` int DEFAULT 0 not null,
	`preset` varchar(100) DEFAULT '' not null,
	`parentID` int DEFAULT 0 not null,
	`parentType` varchar(50) DEFAULT '' not null,
	`lastTopicID` int DEFAULT 0 not null,
	`lastReplyerID` int DEFAULT 0 not null,
	primary key(`fid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;