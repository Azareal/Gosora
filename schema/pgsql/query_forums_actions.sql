CREATE TABLE "forums_actions" (
	`faid` serial not null,
	`fid` int not null,
	`runOnTopicCreation` boolean DEFAULT 0 not null,
	`runDaysAfterTopicCreation` int DEFAULT 0 not null,
	`runDaysAfterTopicLastReply` int DEFAULT 0 not null,
	`action` varchar (50) not null,
	`extra` varchar (200) DEFAULT '' not null,
	primary key(`faid`)
);