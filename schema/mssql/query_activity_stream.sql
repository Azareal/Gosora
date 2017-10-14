DROP TABLE IF EXISTS [activity_stream];
CREATE TABLE [activity_stream] (
	[asid] int not null IDENTITY,
	[actor] int not null,
	[targetUser] int not null,
	[event] nvarchar (50) not null,
	[elementType] nvarchar (50) not null,
	[elementID] int not null,
	primary key([asid])
);