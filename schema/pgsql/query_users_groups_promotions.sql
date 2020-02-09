CREATE TABLE "users_groups_promotions" (
	`pid` serial not null,
	`from_gid` int not null,
	`to_gid` int not null,
	`two_way` boolean DEFAULT 0 not null,
	`level` int not null,
	`posts` int DEFAULT 0 not null,
	`minTime` int not null,
	`registeredFor` int DEFAULT 0 not null,
	primary key(`pid`)
);