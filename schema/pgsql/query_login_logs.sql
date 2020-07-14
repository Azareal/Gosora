CREATE TABLE "login_logs" (
	`lid` serial not null,
	`uid` int not null,
	`success` boolean DEFAULT 0 not null,
	`ipaddress` varchar (200) not null,
	`doneAt` timestamp not null,
	primary key(`lid`)
);