DROP TABLE IF EXISTS [attachments];
CREATE TABLE [attachments] (
	[attachID] int not null IDENTITY,
	[sectionID] int DEFAULT 0 not null,
	[sectionTable] nvarchar (200) DEFAULT 'forums' not null,
	[originID] int not null,
	[originTable] nvarchar (200) DEFAULT 'replies' not null,
	[uploadedBy] int not null,
	[path] nvarchar (200) not null,
	primary key([attachID])
);