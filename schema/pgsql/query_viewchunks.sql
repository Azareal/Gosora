CREATE TABLE "viewchunks" (
	`count` int DEFAULT 0 not null,
	`avg` int DEFAULT 0 not null,
	`createdAt` timestamp not null,
	`route` varchar (200) not null
);