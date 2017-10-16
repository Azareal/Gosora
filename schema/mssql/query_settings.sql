CREATE TABLE [settings] (
	[name] nvarchar (180) not null,
	[content] nvarchar (250) not null,
	[type] nvarchar (50) not null,
	[constraints] nvarchar (200) DEFAULT '' not null,
	unique([name])
);