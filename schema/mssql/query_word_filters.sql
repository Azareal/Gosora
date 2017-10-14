DROP TABLE IF EXISTS [word_filters];
CREATE TABLE [word_filters] (
	[wfid] int not null IDENTITY,
	[find] nvarchar (200) not null,
	[replacement] nvarchar (200) not null,
	primary key([wfid])
);