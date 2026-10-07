package manifestdataservices

import (
	"database/sql"
	"log/slog"
	"strings"

	"github.com/shabs76/roro-local-server/gendb"
)

// media_remote_map remembers the remote URL of each media file that a publish run
// already uploaded, so a retry does not upload it again. The table comes from a
// migration; without it these functions do nothing.

func mediaRemoteMapReady() bool {
	return gendb.ColumnExists("media_remote_map", "local_path")
}

// SelectMediaRemoteUrl returns the remote URL stored for localPath, if any.
func SelectMediaRemoteUrl(localPath string) (string, bool) {
	if !mediaRemoteMapReady() {
		return "", false
	}
	db, err := gendb.InitDb()
	if err != nil {
		return "", false
	}
	var url string
	if err := db.QueryRow("SELECT remote_url FROM media_remote_map WHERE local_path = ?", localPath).Scan(&url); err != nil {
		return "", false
	}
	return url, url != ""
}

// SaveMediaRemoteUrl stores the remote URL of an uploaded media file.
func SaveMediaRemoteUrl(localPath, remoteUrl string) {
	if !mediaRemoteMapReady() || localPath == "" || remoteUrl == "" {
		return
	}
	db, err := gendb.InitDb()
	if err != nil {
		return
	}
	if _, err := db.Exec(
		"INSERT INTO media_remote_map (local_path, remote_url) VALUES (?,?) ON DUPLICATE KEY UPDATE remote_url = VALUES(remote_url), uploaded_at = NOW()",
		localPath, remoteUrl); err != nil {
		slog.Error("Failed to remember uploaded media", "path", localPath, "error", err)
	}
}

// DeleteMediaRemoteUrls forgets the given files. Publish calls it after a vehicle or
// package is published, so a later full re-publish uploads fresh copies.
func DeleteMediaRemoteUrls(localPaths []string) {
	if !mediaRemoteMapReady() || len(localPaths) == 0 {
		return
	}
	db, err := gendb.InitDb()
	if err != nil {
		return
	}
	args := make([]any, len(localPaths))
	for i, p := range localPaths {
		args[i] = p
	}
	in := strings.TrimSuffix(strings.Repeat("?,", len(args)), ",")
	if _, err := db.Exec("DELETE FROM media_remote_map WHERE local_path IN ("+in+")", args...); err != nil {
		slog.Error("Failed to clear uploaded media records", "error", err)
	}
}

// media_files maps the SHA-256 of an uploaded file to the stored copy, so uploading
// the same content again returns the existing file instead of a new copy. The table
// comes from a migration; without it every upload is stored.

func mediaFilesReady() bool {
	return gendb.ColumnExists("media_files", "sha256")
}

// FindMediaFile returns the stored url of a file with this SHA-256, if any.
func FindMediaFile(sha256 string) (string, bool) {
	if !mediaFilesReady() {
		return "", false
	}
	db, err := gendb.InitDb()
	if err != nil {
		return "", false
	}
	var url string
	if err := db.QueryRow("SELECT url FROM media_files WHERE sha256 = ?", sha256).Scan(&url); err != nil {
		return "", false
	}
	return url, url != ""
}

// SaveMediaFile records the stored url of a newly uploaded file.
func SaveMediaFile(sha256, url string, size int64) {
	if !mediaFilesReady() {
		return
	}
	db, err := gendb.InitDb()
	if err != nil {
		return
	}
	if _, err := db.Exec(
		"INSERT INTO media_files (sha256, url, size) VALUES (?,?,?) ON DUPLICATE KEY UPDATE url = VALUES(url), size = VALUES(size)",
		sha256, url, size); err != nil {
		slog.Error("Failed to record uploaded file", "url", url, "error", err)
	}
}

// SelectTallySubmissionId returns the tablet's submission id of a vehicle's current
// inspection, or "" when there is none or the column does not exist yet. Publishing
// forwards it, so the remote server also recognises a resend of the same inspection.
func SelectTallySubmissionId(vehicleId string) string {
	if !gendb.ColumnExists("vehicles_talling", "submission_id") {
		return ""
	}
	db, err := gendb.InitDb()
	if err != nil {
		return ""
	}
	var id sql.NullString
	if err := db.QueryRow("SELECT submission_id FROM vehicles_talling WHERE vehicle_id = ?", vehicleId).Scan(&id); err != nil {
		return ""
	}
	return id.String
}
