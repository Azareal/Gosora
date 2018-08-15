CREATE TABLE "users_groups_scheduler" (
	`uid` int not null,
	`set_group` int not null,
	`issued_by` int not null,
	`issued_at` timestamp not null,
	`revert_at` timestamp not null,
	`temporary` boolean not null,
	primary key(`uid`)
);