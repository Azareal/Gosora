CREATE TABLE `pages`(
	`pid` int not null AUTO_INCREMENT,
	`name` varchar(200) not null,
	`title` varchar(200) not null,
	`body` text not null,
	`allowedGroups` text not null,
	`menuID` int DEFAULT -1 not null,
	primary key(`pid`)
) CHARSET=utf8mb4 COLLATE utf8mb4_general_ci;