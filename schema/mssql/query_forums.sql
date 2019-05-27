CREATE TABLE [forums] (
	[fid] int not null IDENTITY,
	[name] nvarchar (100) not null,
	[desc] nvarchar (200) not null,
	[tmpl] nvarchar (200) not null,
	[active] bit DEFAULT 1 not null,
	[order] int DEFAULT 0 not null,
	[topicCount] int DEFAULT 0 not null,
	[preset] nvarchar (100) DEFAULT '' not null,
	[parentID] int DEFAULT 0 not null,
	[parentType] nvarchar (50) DEFAULT '' not null,
	[lastTopicID] int DEFAULT 0 not null,
	[lastReplyerID] int DEFAULT 0 not null,
	primary key([fid])
);