CREATE TABLE [menu_items] (
	[miid] int not null IDENTITY,
	[mid] int not null,
	[name] nvarchar (200) not null,
	[htmlID] nvarchar (200) DEFAULT '' not null,
	[cssClass] nvarchar (200) DEFAULT '' not null,
	[position] nvarchar (100) not null,
	[path] nvarchar (200) DEFAULT '' not null,
	[aria] nvarchar (200) DEFAULT '' not null,
	[tooltip] nvarchar (200) DEFAULT '' not null,
	[tmplName] nvarchar (200) DEFAULT '' not null,
	[order] int DEFAULT 0 not null,
	[guestOnly] bit DEFAULT 0 not null,
	[memberOnly] bit DEFAULT 0 not null,
	[staffOnly] bit DEFAULT 0 not null,
	[adminOnly] bit DEFAULT 0 not null,
	primary key([miid])
);