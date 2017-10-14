CREATE TABLE `revisions` (
	`index` int not null,
	`content` text not null,
	`contentID` int not null,
	`contentType` varchar(100) DEFAULT 'replies' not null
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;