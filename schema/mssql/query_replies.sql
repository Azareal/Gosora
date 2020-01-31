CREATE TABLE [replies] (
	[rid] int not null IDENTITY,
	[tid] int not null,
	[content] nvarchar (MAX) not null,
	[parsed_content] nvarchar (MAX) not null,
	[createdAt] datetime not null,
	[createdBy] int not null,
	[lastEdit] int DEFAULT 0 not null,
	[lastEditBy] int DEFAULT 0 not null,
	[lastUpdated] datetime not null,
	[ip] nvarchar (200) DEFAULT '' not null,
	[likeCount] int DEFAULT 0 not null,
	[attachCount] int DEFAULT 0 not null,
	[words] int DEFAULT 1 not null,
	[actionType] nvarchar (20) DEFAULT '' not null,
	[poll] int DEFAULT 0 not null,
	primary key([rid]),
	fulltext key([content])
);