CREATE TABLE `polls` (
	`pollID` serial not null,
	`type` int DEFAULT 0 not null,
	`options` json not null,
	`votes` int DEFAULT 0 not null,
	primary key(`pollID`)
);