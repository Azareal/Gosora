CREATE TABLE "replies" (
	`rid` serial not null,
	`tid` int not null,
	`content` text not null,
	`parsed_content` text not null,
	`createdAt` timestamp not null,
	`createdBy` int not null,
	`lastEdit` int DEFAULT 0 not null,
	`lastEditBy` int DEFAULT 0 not null,
	`lastUpdated` timestamp not null,
	`ip` varchar (200) DEFAULT '' not null,
	`likeCount` int DEFAULT 0 not null,
	`attachCount` int DEFAULT 0 not null,
	`words` int DEFAULT 1 not null,
	`actionType` varchar (20) DEFAULT '' not null,
	`poll` int DEFAULT 0 not null,
	primary key(`rid`),
	fulltext key(`content`)
);