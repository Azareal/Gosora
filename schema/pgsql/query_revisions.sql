CREATE TABLE "revisions" (
	`reviseID` serial not null,
	`content` text not null,
	`contentID` int not null,
	`contentType` varchar (100) DEFAULT 'replies' not null,
	`createdAt` timestamp not null,
	primary key(`reviseID`)
);