CREATE TABLE `activity_subscriptions`(
	`user` int not null,
	`targetID` int not null,
	`targetType` varchar(50) not null,
	`level` int DEFAULT 0 not null
);