CREATE TABLE `polls_votes` (
	`pollID` int not null,
	`uid` int not null,
	`option` int DEFAULT 0 not null,
	`castAt` datetime not null,
	`ipaddress` varchar(200) DEFAULT '0.0.0.0.0' not null
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;