-- Remote URLs of media files already uploaded by a publish run that did not finish.
-- A retry reuses the URL instead of uploading the file again. Rows are deleted once
-- the vehicle is published.
CREATE TABLE IF NOT EXISTS media_remote_map (
  local_path VARCHAR(300) NOT NULL,
  remote_url VARCHAR(500) NOT NULL,
  uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (local_path)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
