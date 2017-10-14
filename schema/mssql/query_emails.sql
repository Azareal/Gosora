DROP TABLE IF EXISTS [emails];
CREATE TABLE [emails] (
	[email] nvarchar (200) not null,
	[uid] int not null,
	[validated] bit DEFAULT 0 not null,
	[token] nvarchar (200) DEFAULT '' not null
);