CREATE TABLE "registration_logs" (
	`rlid` serial not null,
	`username` varchar (100) not null,
	`email` varchar (100) not null,
	`failureReason` varchar (100) not null,
	`success` boolean DEFAULT 0 not null,
	`ipaddress` varchar (200) not null,
	`doneAt` timestamp not null,
	primary key(`rlid`)
);