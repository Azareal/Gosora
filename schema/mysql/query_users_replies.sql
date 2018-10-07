CREATE TABLE `users_replies` (
	`rid` int not null AUTO_INCREMENT,
	`uid` int not null,
	`content` text not null,
	`parsed_content` text not null,
	`createdAt` datetime DEFAULT UTC_TIMESTAMP() not null,
	`createdBy` int not null,
	`lastEdit` int DEFAULT 0 not null,
	`lastEditBy` int DEFAULT 0 not null,
	`ipaddress` varchar(200) DEFAULT '0.0.0.0.0' not null,
	primary key(`rid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;