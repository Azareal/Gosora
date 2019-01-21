CREATE TABLE [widgets] (
	[wid] int not null IDENTITY,
	[position] int not null,
	[side] nvarchar (100) not null,
	[type] nvarchar (100) not null,
	[active] bit DEFAULT 0 not null,
	[location] nvarchar (100) not null,
	[data] nvarchar (MAX) DEFAULT '' not null,
	primary key([wid])
);