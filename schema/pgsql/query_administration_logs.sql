CREATE TABLE "administration_logs" (
	`action` varchar (100) not null,
	`elementID` int not null,
	`elementType` varchar (100) not null,
	`ipaddress` varchar (200) not null,
	`actorID` int not null,
	`doneAt` timestamp not null
);