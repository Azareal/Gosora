package panel

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	c "github.com/Azareal/Gosora/common"
)

func Backups(w http.ResponseWriter, r *http.Request, user c.User, backupURL string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "backups", "backups")
	if ferr != nil {
		return ferr
	}

	if backupURL != "" {
		// We don't want them trying to break out of this directory, it shouldn't hurt since it's a super admin, but it's always good to practice good security hygiene, especially if this is one of many instances on a managed server not controlled by the superadmin/s
		backupURL = c.Stripslashes(backupURL)

		ext := filepath.Ext("./backups/" + backupURL)
		if ext != ".sql" && ext != ".zip" {
			return c.NotFound(w, r, basePage.Header)
		}
		info, err := os.Stat("./backups/" + backupURL)
		if err != nil {
			return c.NotFound(w, r, basePage.Header)
		}

		h := w.Header()
		h.Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		if ext == ".sql" {
			// TODO: Change the served filename to gosora_backup_%timestamp%.sql, the time the file was generated, not when it was modified aka what the name of it should be
			h.Set("Content-Disposition", "attachment; filename=gosora_backup.sql")
			h.Set("Content-Type", "application/sql")
		} else {
			// TODO: Change the served filename to gosora_backup_%timestamp%.zip, the time the file was generated, not when it was modified aka what the name of it should be
			h.Set("Content-Disposition", "attachment; filename=gosora_backup.zip")
			h.Set("Content-Type", "application/zip")
		}
		// TODO: Fix the problem where non-existent files aren't greeted with custom 404s on ServeFile()'s side
		http.ServeFile(w, r, "./backups/"+backupURL)
		err = c.AdminLogs.Create("download", 0, "backup", user.LastIP, user.ID)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		return nil
	}

	var backupList []c.BackupItem
	backupFiles, err := ioutil.ReadDir("./backups")
	if err != nil {
		return c.InternalError(err, w, r)
	}
	for _, backupFile := range backupFiles {
		ext := filepath.Ext(backupFile.Name())
		if ext != ".sql" {
			continue
		}
		backupList = append(backupList, c.BackupItem{backupFile.Name(), backupFile.ModTime()})
	}

	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_backups", c.PanelBackupPage{basePage, backupList}})
}
