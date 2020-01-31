CREATE TABLE [activity_stream] (
	[asid] int not null IDENTITY,
	[actor] int not null,
	[targetUser] int not null,
	[event] nvarchar (50) not null,
	[elementType] nvarchar (50) not null,
	[elementID] int not null,
	[createdAt] datetime not null,
	[extra] nvarchar (200) DEFAULT '' not null,
	primary key([asid])
);