CREATE TABLE `login_logs` (
	`lid` int not null AUTO_INCREMENT,
	`uid` int not null,
	`success` boolean DEFAULT 0 not null,
	`ipaddress` varchar(200) not null,
	`doneAt` datetime not null,
	primary key(`lid`)
);