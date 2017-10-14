CREATE TABLE `themes` (
	`uname` varchar (180) not null,
	`default` boolean DEFAULT 0 not null,
	unique(`uname`)
);