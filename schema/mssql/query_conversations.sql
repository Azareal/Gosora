CREATE TABLE [conversations] (
	[cid] int not null IDENTITY,
	[createdBy] int not null,
	[createdAt] datetime not null,
	[lastReplyAt] datetime not null,
	[lastReplyBy] int not null,
	primary key([cid])
);