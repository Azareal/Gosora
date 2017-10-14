DROP TABLE IF EXISTS [users_groups_scheduler];
CREATE TABLE [users_groups_scheduler] (
	[uid] int not null,
	[set_group] int not null,
	[issued_by] int not null,
	[issued_at] datetime not null,
	[revert_at] datetime not null,
	[temporary] bit not null,
	primary key([uid])
);