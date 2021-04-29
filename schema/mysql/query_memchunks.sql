CREATE TABLE `memchunks`(
	`count` int DEFAULT 0 not null,
	`stack` int DEFAULT 0 not null,
	`heap` int DEFAULT 0 not null,
	`createdAt` datetime not null
);