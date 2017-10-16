CREATE TABLE [revisions] (
	[index] int not null,
	[content] nvarchar (MAX) not null,
	[contentID] int not null,
	[contentType] nvarchar (100) DEFAULT 'replies' not null
);