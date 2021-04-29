CREATE TABLE `polls`(
	`pollID` int not null AUTO_INCREMENT,
	`parentID` int DEFAULT 0 not null,
	`parentTable` varchar(100) DEFAULT 'topics' not null,
	`type` int DEFAULT 0 not null,
	`options` text not null,
	`votes` int DEFAULT 0 not null,
	primary key(`pollID`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;