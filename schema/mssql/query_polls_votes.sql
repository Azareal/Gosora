CREATE TABLE [polls_votes] (
	[pollID] int not null,
	[uid] int not null,
	[option] int DEFAULT 0 not null,
	[castAt] datetime not null,
	[ip] nvarchar (200) DEFAULT '' not null
);