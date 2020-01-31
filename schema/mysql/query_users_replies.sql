CREATE TABLE `users_replies` (
	`rid` int not null AUTO_INCREMENT,
	`uid` int not null,
	`content` text not null,
	`parsed_content` text not null,
	`createdAt` datetime not null,
	`createdBy` int not null,
	`lastEdit` int DEFAULT 0 not null,
	`lastEditBy` int DEFAULT 0 not null,
	`ip` varchar(200) DEFAULT '' not null,
	primary key(`rid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;