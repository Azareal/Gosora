CREATE TABLE "polls_votes" (
	`pollID` int not null,
	`uid` int not null,
	`option` int DEFAULT 0 not null,
	`castAt` timestamp not null,
	`ip` varchar (200) DEFAULT '' not null
);