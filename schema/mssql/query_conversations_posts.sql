CREATE TABLE [conversations_posts] (
	[pid] int not null IDENTITY,
	[cid] int not null,
	[createdBy] int not null,
	[body] nvarchar (50) not null,
	[post] nvarchar (50) DEFAULT '' not null,
	primary key([pid])
);