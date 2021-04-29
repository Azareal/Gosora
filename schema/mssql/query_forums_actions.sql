CREATE TABLE [forums_actions] (
	[faid] int not null IDENTITY,
	[fid] int not null,
	[runOnTopicCreation] bit DEFAULT 0 not null,
	[runDaysAfterTopicCreation] int DEFAULT 0 not null,
	[runDaysAfterTopicLastReply] int DEFAULT 0 not null,
	[action] nvarchar (50) not null,
	[extra] nvarchar (200) DEFAULT '' not null,
	primary key([faid])
);