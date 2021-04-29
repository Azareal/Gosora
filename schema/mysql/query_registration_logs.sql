CREATE TABLE `registration_logs`(
	`rlid` int not null AUTO_INCREMENT,
	`username` varchar(100) not null,
	`email` varchar(100) not null,
	`failureReason` varchar(100) not null,
	`success` boolean DEFAULT 0 not null,
	`ipaddress` varchar(200) not null,
	`doneAt` datetime not null,
	primary key(`rlid`)
);