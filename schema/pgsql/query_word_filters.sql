CREATE TABLE `word_filters` (
	`wfid` serial not null,
	`find` varchar (200) not null,
	`replacement` varchar (200) not null,
	primary key(`wfid`)
);