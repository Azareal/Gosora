CREATE TABLE `emails` (
	`email` varchar (200) not null,
	`uid` int not null,
	`validated` boolean DEFAULT 0 not null,
	`token` varchar (200) DEFAULT '' not null
);