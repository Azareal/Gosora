CREATE TABLE `pages` (
	`pid` serial not null,
	`name` varchar (200) not null,
	`title` varchar (200) not null,
	`body` text not null,
	`allowedGroups` text not null,
	`menuID` int DEFAULT -1 not null,
	primary key(`pid`)
);