CREATE TABLE "polls_options" (
	`pollID` int not null,
	`option` int DEFAULT 0 not null,
	`votes` int DEFAULT 0 not null
);