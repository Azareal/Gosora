DROP TABLE IF EXISTS [forums_permissions];
CREATE TABLE [forums_permissions] (
	[fid] int not null,
	[gid] int not null,
	[preset] nvarchar (100) DEFAULT '' not null,
	[permissions] nvarchar (MAX) not null,
	primary key([fid],[gid])
);