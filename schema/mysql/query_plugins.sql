CREATE TABLE `plugins`(
	`uname` varchar(180) not null,
	`active` boolean DEFAULT 0 not null,
	`installed` boolean DEFAULT 0 not null,
	unique(`uname`)
);