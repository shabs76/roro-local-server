-- Inspection reliability schema changes.
-- Every statement is idempotent (IF [NOT] EXISTS), so the server runs this file on
-- every start (see gendb.RunMigrations). It can also be applied by hand.

-- manifest_vehicles: change tracking for the tablet status sync.
ALTER TABLE manifest_vehicles ADD COLUMN IF NOT EXISTS is_published VARCHAR(10) NOT NULL DEFAULT 'no';
ALTER TABLE manifest_vehicles ADD COLUMN IF NOT EXISTS updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);
ALTER TABLE manifest_vehicles ADD INDEX IF NOT EXISTS idx_mv_manifest_updated (manifest_id, updated_at);

-- vehicles_talling: manifest id and the tablet's submission id (idempotency key).
ALTER TABLE vehicles_talling ADD COLUMN IF NOT EXISTS manifest_id VARCHAR(225) NOT NULL DEFAULT '';
ALTER TABLE vehicles_talling ADD COLUMN IF NOT EXISTS submission_id VARCHAR(64) NULL DEFAULT NULL;
ALTER TABLE vehicles_talling ADD UNIQUE INDEX IF NOT EXISTS uq_vt_submission (submission_id);
UPDATE vehicles_talling vt INNER JOIN manifest_vehicles mv ON mv.vehicle_id = vt.vehicle_id
  SET vt.manifest_id = mv.manifest_id
  WHERE vt.manifest_id = '';

-- History tables. Archiving copies only the columns that exist in both the active and
-- the history table, so both sides must carry the same columns (plus archived_at).
-- CREATE TABLE ... LIKE copies the UNIQUE keys of the active table. Those keys allow
-- only one archived batch per vehicle, so they are dropped here.
CREATE TABLE IF NOT EXISTS vehicles_talling_history LIKE vehicles_talling;
ALTER TABLE vehicles_talling_history ADD COLUMN IF NOT EXISTS archived_at DATETIME NOT NULL DEFAULT '1000-01-01 00:00:00';
ALTER TABLE vehicles_talling_history ADD COLUMN IF NOT EXISTS manifest_id VARCHAR(225) NOT NULL DEFAULT '';
ALTER TABLE vehicles_talling_history ADD COLUMN IF NOT EXISTS user_id VARCHAR(225) NOT NULL DEFAULT '';
ALTER TABLE vehicles_talling_history ADD COLUMN IF NOT EXISTS submission_id VARCHAR(64) NULL DEFAULT NULL;
ALTER TABLE vehicles_talling_history DROP INDEX IF EXISTS vehicle_id;
ALTER TABLE vehicles_talling_history DROP INDEX IF EXISTS uq_vt_submission;
ALTER TABLE vehicles_talling_history ADD INDEX IF NOT EXISTS idx_hist_vehicle_archived (vehicle_id, archived_at);

CREATE TABLE IF NOT EXISTS vehicles_inspection_history LIKE vehicles_inspection;
ALTER TABLE vehicles_inspection_history ADD COLUMN IF NOT EXISTS archived_at DATETIME NOT NULL DEFAULT '1000-01-01 00:00:00';
ALTER TABLE vehicles_inspection_history DROP INDEX IF EXISTS unique_vehicle_check;
ALTER TABLE vehicles_inspection_history ADD INDEX IF NOT EXISTS idx_hist_vehicle_archived (vehicle_id, archived_at);

CREATE TABLE IF NOT EXISTS inspection_image_history LIKE inspection_image;
ALTER TABLE inspection_image_history ADD COLUMN IF NOT EXISTS archived_at DATETIME NOT NULL DEFAULT '1000-01-01 00:00:00';
ALTER TABLE inspection_image_history ADD INDEX IF NOT EXISTS idx_hist_inspection_archived (inspection_id, archived_at);

CREATE TABLE IF NOT EXISTS onboard_packages_history LIKE onboard_packages;
ALTER TABLE onboard_packages_history ADD COLUMN IF NOT EXISTS archived_at DATETIME NOT NULL DEFAULT '1000-01-01 00:00:00';
ALTER TABLE onboard_packages_history ADD INDEX IF NOT EXISTS idx_hist_vehicle_archived (vehicle_id, archived_at);

CREATE TABLE IF NOT EXISTS onboard_packages_media_history LIKE onboard_packages_media;
ALTER TABLE onboard_packages_media_history ADD COLUMN IF NOT EXISTS archived_at DATETIME NOT NULL DEFAULT '1000-01-01 00:00:00';
ALTER TABLE onboard_packages_media_history ADD INDEX IF NOT EXISTS idx_hist_package_archived (package_id, archived_at);

CREATE TABLE IF NOT EXISTS inspection_remarks_history LIKE inspection_remarks;
ALTER TABLE inspection_remarks_history ADD COLUMN IF NOT EXISTS archived_at DATETIME NOT NULL DEFAULT '1000-01-01 00:00:00';
ALTER TABLE inspection_remarks_history DROP INDEX IF EXISTS vehicle_id_2;
ALTER TABLE inspection_remarks_history ADD INDEX IF NOT EXISTS idx_hist_vehicle_archived (vehicle_id, archived_at);

CREATE TABLE IF NOT EXISTS vehicle_galllery_history LIKE vehicle_galllery;
ALTER TABLE vehicle_galllery_history ADD COLUMN IF NOT EXISTS archived_at DATETIME NOT NULL DEFAULT '1000-01-01 00:00:00';
ALTER TABLE vehicle_galllery_history ADD INDEX IF NOT EXISTS idx_hist_vehicle_archived (vehicle_id, archived_at);
