CREATE TABLE "users_replies" (
	`rid` serial not null,
	`uid` int not null,
	`content` text not null,
	`parsed_content` text not null,
	`createdAt` timestamp not null,
	`createdBy` int not null,
	`lastEdit` int DEFAULT 0 not null,
	`lastEditBy` int DEFAULT 0 not null,
	`ip` varchar (200) DEFAULT '' not null,
	primary key(`rid`)
);