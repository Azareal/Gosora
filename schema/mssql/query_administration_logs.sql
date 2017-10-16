CREATE TABLE [administration_logs] (
	[action] nvarchar (100) not null,
	[elementID] int not null,
	[elementType] nvarchar (100) not null,
	[ipaddress] nvarchar (200) not null,
	[actorID] int not null,
	[doneAt] datetime not null
);