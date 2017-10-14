DROP TABLE IF EXISTS [plugins];
CREATE TABLE [plugins] (
	[uname] nvarchar (180) not null,
	[active] bit DEFAULT 0 not null,
	[installed] bit DEFAULT 0 not null,
	unique([uname])
);