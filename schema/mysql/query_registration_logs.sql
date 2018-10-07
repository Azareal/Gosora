CREATE TABLE `registration_logs` (
	`rlid` int not null AUTO_INCREMENT,
	`username` varchar(100) not null,
	`email` varchar(100) not null,
	`failureReason` varchar(100) not null,
	`success` bool DEFAULT 0 not null,
	`ipaddress` varchar(200) not null,
	`doneAt` datetime DEFAULT UTC_TIMESTAMP() not null,
	primary key(`rlid`)
);