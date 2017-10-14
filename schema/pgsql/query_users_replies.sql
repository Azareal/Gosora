CREATE TABLE `users_replies` (
	`rid` serial not null,
	`uid` int not null,
	`content` text not null,
	`parsed_content` text not null,
	`createdAt` timestamp not null,
	`createdBy` int not null,
	`lastEdit` int not null,
	`lastEditBy` int not null,
	`ipaddress` varchar (200) DEFAULT '0.0.0.0.0' not null,
	primary key(`rid`)
);