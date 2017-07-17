CREATE TABLE `users_penalties` (
	`uid` int not null,
	`element_id` int not null,
	`element_type` varchar(50) not null,
	`overrides` text not null,
	`mod_queue` boolean DEFAULT 0 not null,
	`shadow_ban` boolean DEFAULT 0 not null,
	`no_avatar` boolean DEFAULT 0 not null,
	`issued_by` int not null,
	`issued_at` datetime not null,
	`expiry` duration not null
);