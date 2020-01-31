CREATE TABLE [users_replies] (
	[rid] int not null IDENTITY,
	[uid] int not null,
	[content] nvarchar (MAX) not null,
	[parsed_content] nvarchar (MAX) not null,
	[createdAt] datetime not null,
	[createdBy] int not null,
	[lastEdit] int DEFAULT 0 not null,
	[lastEditBy] int DEFAULT 0 not null,
	[ip] nvarchar (200) DEFAULT '' not null,
	primary key([rid])
);