CREATE TABLE [pages] (
	[pid] int not null IDENTITY,
	[name] nvarchar (200) not null,
	[title] nvarchar (200) not null,
	[body] nvarchar (MAX) not null,
	[allowedGroups] nvarchar (MAX) not null,
	[menuID] int DEFAULT -1 not null,
	primary key([pid])
);