CREATE TABLE [registration_logs] (
	[rlid] int not null IDENTITY,
	[username] nvarchar (100) not null,
	[email] nvarchar (100) not null,
	[failureReason] nvarchar (100) not null,
	[success] bool DEFAULT 0 not null,
	[ipaddress] nvarchar (200) not null,
	[doneAt] datetime not null,
	primary key([rlid])
);