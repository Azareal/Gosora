CREATE TABLE `activity_stream_matches`(
	`watcher` int not null,
	`asid` int not null,
	foreign key(`asid`) REFERENCES `activity_stream`(`asid`) ON DELETE CASCADE
);