CREATE TABLE [password_resets] (
	[email] nvarchar (200) not null,
	[uid] int not null,
	[validated] nvarchar (200) not null,
	[token] nvarchar (200) not null,
	[createdAt] datetime not null
);