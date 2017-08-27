CREATE TABLE `word_filters` (
	`wfid` int not null AUTO_INCREMENT,
	`find` varchar(200) not null,
	`replacement` varchar(200) not null,
	primary key(`wfid`)
);