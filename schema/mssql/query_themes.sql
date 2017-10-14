DROP TABLE IF EXISTS [themes];
CREATE TABLE [themes] (
	[uname] nvarchar (180) not null,
	[default] bit DEFAULT 0 not null,
	unique([uname])
);