-- Upload dedupe: one stored file per distinct content. Older app builds upload the
-- same photo again on every retry; the server now answers with the existing file.
CREATE TABLE IF NOT EXISTS media_files (
  sha256 CHAR(64) NOT NULL,
  url VARCHAR(300) NOT NULL,
  size BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (sha256)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

-- Tablet-vs-server comparison reports, one row per comparison a tablet runs.
CREATE TABLE IF NOT EXISTS tablet_compare_reports (
  report_id VARCHAR(64) NOT NULL,
  manifest_id VARCHAR(225) NOT NULL,
  device_id VARCHAR(200) NOT NULL,
  device_name VARCHAR(200) NOT NULL DEFAULT '',
  app_version VARCHAR(50) NOT NULL DEFAULT '',
  user_id VARCHAR(225) NOT NULL DEFAULT '',
  summary_json TEXT NOT NULL,
  items_json MEDIUMTEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (report_id),
  KEY idx_tcr_manifest_device (manifest_id, device_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
