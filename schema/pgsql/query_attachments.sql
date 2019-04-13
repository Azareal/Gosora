CREATE TABLE "attachments" (
	`attachID` serial not null,
	`sectionID` int DEFAULT 0 not null,
	`sectionTable` varchar (200) DEFAULT 'forums' not null,
	`originID` int not null,
	`originTable` varchar (200) DEFAULT 'replies' not null,
	`uploadedBy` int not null,
	`path` varchar (200) not null,
	`extra` varchar (200) not null,
	primary key(`attachID`)
);