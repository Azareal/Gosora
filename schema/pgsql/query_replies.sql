CREATE TABLE `replies` (
	`rid` serial not null,
	`tid` int not null,
	`content` text not null,
	`parsed_content` text not null,
	`createdAt` timestamp not null,
	`createdBy` int not null,
	`lastEdit` int DEFAULT 0 not null,
	`lastEditBy` int DEFAULT 0 not null,
	`lastUpdated` timestamp not null,
	`ipaddress` varchar (200) DEFAULT '0.0.0.0.0' not null,
	`likeCount` int DEFAULT 0 not null,
	`words` int DEFAULT 1 not null,
	`actionType` varchar (20) DEFAULT '' not null,
	primary key(`rid`)
);