CREATE TABLE "conversations_posts" (
	`pid` serial not null,
	`cid` int not null,
	`createdBy` int not null,
	`body` varchar (50) not null,
	`post` varchar (50) DEFAULT '' not null,
	primary key(`pid`)
);