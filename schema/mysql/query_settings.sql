CREATE TABLE `settings`(
	`name` varchar(180) not null,
	`content` varchar(250) not null,
	`type` varchar(50) not null,
	`constraints` varchar(200) DEFAULT '' not null,
	unique(`name`)
);