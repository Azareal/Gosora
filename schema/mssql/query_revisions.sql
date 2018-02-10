CREATE TABLE [revisions] (
	[reviseID] int not null IDENTITY,
	[content] nvarchar (MAX) not null,
	[contentID] int not null,
	[contentType] nvarchar (100) DEFAULT 'replies' not null,
	[createdAt] datetime not null,
	primary key([reviseID])
);