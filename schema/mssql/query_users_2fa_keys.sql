CREATE TABLE [users_2fa_keys] (
	[uid] int not null,
	[secret] nvarchar (100) not null,
	[scratch1] nvarchar (50) not null,
	[scratch2] nvarchar (50) not null,
	[scratch3] nvarchar (50) not null,
	[scratch4] nvarchar (50) not null,
	[scratch5] nvarchar (50) not null,
	[scratch6] nvarchar (50) not null,
	[scratch7] nvarchar (50) not null,
	[scratch8] nvarchar (50) not null,
	[createdAt] datetime not null,
	primary key([uid])
);