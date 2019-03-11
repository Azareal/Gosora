CREATE TABLE `password_resets` (
	`email` varchar(200) not null,
	`uid` int not null,
	`validated` varchar(200) not null,
	`token` varchar(200) not null,
	`createdAt` datetime not null
);