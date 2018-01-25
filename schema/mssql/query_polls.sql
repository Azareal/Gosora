CREATE TABLE [polls] (
	[pollID] int not null IDENTITY,
	[type] int DEFAULT 0 not null,
	[options] nvarchar (MAX) not null,
	[votes] int DEFAULT 0 not null,
	primary key([pollID])
);