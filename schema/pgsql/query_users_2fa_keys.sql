CREATE TABLE `users_2fa_keys` (
	`uid` int not null,
	`secret` varchar (100) not null,
	`scratch1` varchar (50) not null,
	`scratch2` varchar (50) not null,
	`scratch3` varchar (50) not null,
	`scratch4` varchar (50) not null,
	`scratch5` varchar (50) not null,
	`scratch6` varchar (50) not null,
	`scratch7` varchar (50) not null,
	`scratch8` varchar (50) not null,
	`createdAt` timestamp not null,
	primary key(`uid`)
);