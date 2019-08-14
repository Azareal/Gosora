CREATE TABLE "conversations" (
	`cid` serial not null,
	`createdBy` int not null,
	`createdAt` timestamp not null,
	`lastReplyAt` timestamp not null,
	`lastReplyBy` int not null,
	primary key(`cid`)
);