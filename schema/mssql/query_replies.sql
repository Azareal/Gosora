DROP TABLE IF EXISTS [replies];
CREATE TABLE [replies] (
	[rid] int not null IDENTITY,
	[tid] int not null,
	[content] nvarchar (MAX) not null,
	[parsed_content] nvarchar (MAX) not null,
	[createdAt] datetime not null,
	[createdBy] int not null,
	[lastEdit] int not null,
	[lastEditBy] int not null,
	[lastUpdated] datetime not null,
	[ipaddress] nvarchar (200) DEFAULT '0.0.0.0.0' not null,
	[likeCount] int DEFAULT 0 not null,
	[words] int DEFAULT 1 not null,
	[actionType] nvarchar (20) DEFAULT '' not null,
	primary key([rid])
);