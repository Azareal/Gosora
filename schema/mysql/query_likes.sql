CREATE TABLE `likes` (
	`weight` tinyint DEFAULT 1 not null,
	`targetItem` int not null,
	`targetType` varchar(50) DEFAULT 'replies' not null,
	`sentBy` int not null,
	`recalc` tinyint DEFAULT 0 not null
);