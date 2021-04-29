CREATE TABLE `menu_items`(
	`miid` int not null AUTO_INCREMENT,
	`mid` int not null,
	`name` varchar(200) DEFAULT '' not null,
	`htmlID` varchar(200) DEFAULT '' not null,
	`cssClass` varchar(200) DEFAULT '' not null,
	`position` varchar(100) not null,
	`path` varchar(200) DEFAULT '' not null,
	`aria` varchar(200) DEFAULT '' not null,
	`tooltip` varchar(200) DEFAULT '' not null,
	`tmplName` varchar(200) DEFAULT '' not null,
	`order` int DEFAULT 0 not null,
	`guestOnly` boolean DEFAULT 0 not null,
	`memberOnly` boolean DEFAULT 0 not null,
	`staffOnly` boolean DEFAULT 0 not null,
	`adminOnly` boolean DEFAULT 0 not null,
	primary key(`miid`)
);