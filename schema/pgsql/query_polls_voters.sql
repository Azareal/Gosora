CREATE TABLE `polls_voters` (
	`pollID` int not null,
	`uid` int not null,
	`option` int DEFAULT 0 not null
);