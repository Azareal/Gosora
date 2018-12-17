CREATE TABLE [login_logs] (
	[lid] int not null IDENTITY,
	[uid] int not null,
	[success] bool DEFAULT 0 not null,
	[ipaddress] nvarchar (200) not null,
	[doneAt] datetime not null,
	primary key([lid])
);