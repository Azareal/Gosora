package main

import "time"

func handleExpiredScheduledGroups() error {
	rows, err := get_expired_scheduled_groups_stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var uid int
	for rows.Next() {
		err := rows.Scan(&uid)
		if err != nil {
			return err
		}
		_, err = replace_schedule_group_stmt.Exec(uid, 0, 0, time.Now(), false)
		if err != nil {
			return err
		}
		_, err = set_temp_group_stmt.Exec(0, uid)
		if err != nil {
			return err
		}
		_ = users.Load(uid)
	}
	return rows.Err()
}
