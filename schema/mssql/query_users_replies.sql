CREATE TABLE [users_replies] (
	[rid] int not null IDENTITY,
	[uid] int not null,
	[content] nvarchar (MAX) not null,
	[parsed_content] nvarchar (MAX) not null,
	[createdAt] datetime not null,
	[createdBy] int not null,
	[lastEdit] int not null,
	[lastEditBy] int not null,
	[ipaddress] nvarchar (200) DEFAULT '0.0.0.0.0' not null,
	primary key([rid])
);