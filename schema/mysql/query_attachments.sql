CREATE TABLE `attachments`(
	`attachID` int not null AUTO_INCREMENT,
	`sectionID` int DEFAULT 0 not null,
	`sectionTable` varchar(200) DEFAULT 'forums' not null,
	`originID` int not null,
	`originTable` varchar(200) DEFAULT 'replies' not null,
	`uploadedBy` int not null,
	`path` varchar(200) not null,
	`extra` varchar(200) not null,
	primary key(`attachID`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;