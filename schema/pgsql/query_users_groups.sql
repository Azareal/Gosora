CREATE TABLE "users_groups" (
	`gid` serial not null,
	`name` varchar (100) not null,
	`permissions` text not null,
	`plugin_perms` text not null,
	`is_mod` boolean DEFAULT 0 not null,
	`is_admin` boolean DEFAULT 0 not null,
	`is_banned` boolean DEFAULT 0 not null,
	`user_count` int DEFAULT 0 not null,
	`tag` varchar (50) DEFAULT '' not null,
	primary key(`gid`)
);