CREATE TABLE `forums_actions` (
	`faid` int not null AUTO_INCREMENT,
	`fid` int not null,
	`runOnTopicCreation` boolean DEFAULT 0 not null,
	`runDaysAfterTopicCreation` int DEFAULT 0 not null,
	`runDaysAfterTopicLastReply` int DEFAULT 0 not null,
	`action` varchar(50) not null,
	`extra` varchar(200) DEFAULT '' not null,
	primary key(`faid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;