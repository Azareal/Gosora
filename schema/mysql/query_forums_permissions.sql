CREATE TABLE `forums_permissions`(
	`fid` int not null,
	`gid` int not null,
	`preset` varchar(100) DEFAULT '' not null,
	`permissions` text not null,
	primary key(`fid`,`gid`)
);