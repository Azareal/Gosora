CREATE TABLE "activity_stream" (
	`asid` serial not null,
	`actor` int not null,
	`targetUser` int not null,
	`event` varchar (50) not null,
	`elementType` varchar (50) not null,
	`elementID` int not null,
	`createdAt` timestamp not null,
	primary key(`asid`)
);