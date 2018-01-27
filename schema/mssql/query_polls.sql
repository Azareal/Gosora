CREATE TABLE [polls] (
	[pollID] int not null IDENTITY,
	[parentID] int DEFAULT 0 not null,
	[parentTable] nvarchar (100) DEFAULT 'topics' not null,
	[type] int DEFAULT 0 not null,
	[options] nvarchar (MAX) not null,
	[votes] int DEFAULT 0 not null,
	primary key([pollID])
);