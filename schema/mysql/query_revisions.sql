CREATE TABLE `revisions` (
	`reviseID` int not null AUTO_INCREMENT,
	`content` text not null,
	`contentID` int not null,
	`contentType` varchar(100) DEFAULT 'replies' not null,
	`createdAt` datetime DEFAULT UTC_TIMESTAMP() not null,
	primary key(`reviseID`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;