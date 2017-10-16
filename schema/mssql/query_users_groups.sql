CREATE TABLE [users_groups] (
	[gid] int not null IDENTITY,
	[name] nvarchar (100) not null,
	[permissions] nvarchar (MAX) not null,
	[plugin_perms] nvarchar (MAX) not null,
	[is_mod] bit DEFAULT 0 not null,
	[is_admin] bit DEFAULT 0 not null,
	[is_banned] bit DEFAULT 0 not null,
	[tag] nvarchar (50) DEFAULT '' not null,
	primary key([gid])
);