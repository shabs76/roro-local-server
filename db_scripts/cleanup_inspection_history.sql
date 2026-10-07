-- Cleanup of vehicle inspection history created by tablet retries.
--
-- Before the fix, every resend of the same inspection (tablet retry loop) moved the
-- saved inspection into the *_history tables. Vehicles therefore show as "multiple
-- inspections" although they were inspected once. The old archive code also copied
-- vehicles_talling rows with SELECT *, which put maker_id, body_id and user_id into
-- the wrong columns of vehicles_talling_history.
--
-- This script:
--   Part 1  repairs shifted vehicles_talling_history rows,
--   Part 2  aligns the archived_at value of each batch across the history tables
--           (the old code took NOW() per statement, so one batch could span two seconds),
--   Part 3  finds retry duplicates: a history batch whose tally fields AND check
--           results equal the next inspection of the same vehicle, archived less than
--           @max_gap_minutes after it was saved,
--   Part 4  deletes those duplicate batches from every history table.
--
-- How to run:
--   1. Apply gendb/migrations/001_inspection_reliability.sql first (as an administrator).
--   2. Back up the history tables:
--        mariadb-dump -u root -p roro_local vehicles_talling_history vehicles_inspection_history \
--          inspection_image_history inspection_remarks_history onboard_packages_history \
--          onboard_packages_media_history vehicle_galllery_history > history_backup.sql
--   3. Dry run (nothing changes, @apply = 0):
--        mariadb -u root -p roro_local < cleanup_inspection_history.sql
--   4. Review the reports. To apply, set @apply = 1 below and run it again.
--   5. Run the apply step again until duplicate_batches_to_delete is 0. A batch that
--      the old code split across two seconds can only be judged after its neighbour
--      is gone, so a second pass may remove a few more.
--
-- The whole script runs in one transaction.

SET @apply = 0;
SET @max_gap_minutes = 30;

START TRANSACTION;

-- ---------------------------------------------------------------------------------
-- Part 1: repair shifted vehicles_talling_history rows.
-- A shifted row holds a maker id in manifest_id and a body id in maker_id.
-- ---------------------------------------------------------------------------------
SELECT COUNT(*) AS shifted_tally_history_rows
FROM vehicles_talling_history h
WHERE h.manifest_id IN (SELECT maker_id FROM vehicle_makers)
  AND h.maker_id IN (SELECT body_id FROM vehicle_bodies);

UPDATE vehicles_talling_history h
INNER JOIN manifest_vehicles mv ON mv.vehicle_id = h.vehicle_id
SET h.user_id = h.body_id,
    h.body_id = h.maker_id,
    h.maker_id = h.manifest_id,
    h.manifest_id = mv.manifest_id
WHERE @apply = 1
  AND h.manifest_id IN (SELECT maker_id FROM vehicle_makers)
  AND h.maker_id IN (SELECT body_id FROM vehicle_bodies);

-- ---------------------------------------------------------------------------------
-- Part 2: give every row of a batch the archived_at of its tally row.
-- The tally row was archived last, so a row whose archived_at matches no tally batch
-- belongs to the nearest tally batch up to 3 seconds later. Rows that already match a
-- tally batch are left alone, so close batches are never merged.
-- ---------------------------------------------------------------------------------
SELECT
  (SELECT COUNT(*) FROM vehicles_inspection_history x
    WHERE NOT EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = x.vehicle_id AND h.archived_at = x.archived_at)) AS orphan_inspection_rows,
  (SELECT COUNT(*) FROM inspection_remarks_history x
    WHERE NOT EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = x.vehicle_id AND h.archived_at = x.archived_at)) AS orphan_remark_rows,
  (SELECT COUNT(*) FROM vehicle_galllery_history x
    WHERE NOT EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = x.vehicle_id AND h.archived_at = x.archived_at)) AS orphan_gallery_rows,
  (SELECT COUNT(*) FROM onboard_packages_history x
    WHERE NOT EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = x.vehicle_id AND h.archived_at = x.archived_at)) AS orphan_package_rows;

-- Effective batch time of every archived check result. Part 3 compares checks by this
-- value, so the dry run sees the same batches as the real run.
DROP TEMPORARY TABLE IF EXISTS insp_batch;
CREATE TEMPORARY TABLE insp_batch AS
SELECT x.inspection_id, x.vehicle_id, x.check_id, x.status,
       COALESCE(
         (SELECT h.archived_at FROM vehicles_talling_history h
           WHERE h.vehicle_id = x.vehicle_id AND h.archived_at = x.archived_at LIMIT 1),
         (SELECT MIN(h.archived_at) FROM vehicles_talling_history h
           WHERE h.vehicle_id = x.vehicle_id AND h.archived_at > x.archived_at
             AND h.archived_at <= x.archived_at + INTERVAL 3 SECOND),
         x.archived_at) AS batch_at
FROM vehicles_inspection_history x;
CREATE INDEX idx_insp_batch ON insp_batch (vehicle_id, batch_at, check_id);

UPDATE vehicles_inspection_history x
INNER JOIN insp_batch b ON b.inspection_id = x.inspection_id
SET x.archived_at = b.batch_at
WHERE @apply = 1 AND x.archived_at <> b.batch_at;

UPDATE inspection_image_history i
INNER JOIN vehicles_inspection_history x ON x.inspection_id = i.inspection_id
SET i.archived_at = x.archived_at
WHERE @apply = 1 AND i.archived_at <> x.archived_at;

UPDATE onboard_packages_history x
SET x.archived_at = (SELECT MIN(h.archived_at) FROM vehicles_talling_history h
                      WHERE h.vehicle_id = x.vehicle_id AND h.archived_at > x.archived_at
                        AND h.archived_at <= x.archived_at + INTERVAL 3 SECOND)
WHERE @apply = 1
  AND NOT EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = x.vehicle_id AND h.archived_at = x.archived_at)
  AND EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = x.vehicle_id AND h.archived_at > x.archived_at
                AND h.archived_at <= x.archived_at + INTERVAL 3 SECOND);

UPDATE onboard_packages_media_history m
INNER JOIN onboard_packages_history p ON p.package_id = m.package_id
SET m.archived_at = p.archived_at
WHERE @apply = 1 AND m.archived_at <> p.archived_at;

UPDATE inspection_remarks_history x
SET x.archived_at = (SELECT MIN(h.archived_at) FROM vehicles_talling_history h
                      WHERE h.vehicle_id = x.vehicle_id AND h.archived_at > x.archived_at
                        AND h.archived_at <= x.archived_at + INTERVAL 3 SECOND)
WHERE @apply = 1
  AND NOT EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = x.vehicle_id AND h.archived_at = x.archived_at)
  AND EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = x.vehicle_id AND h.archived_at > x.archived_at
                AND h.archived_at <= x.archived_at + INTERVAL 3 SECOND);

UPDATE vehicle_galllery_history x
SET x.archived_at = (SELECT MIN(h.archived_at) FROM vehicles_talling_history h
                      WHERE h.vehicle_id = x.vehicle_id AND h.archived_at > x.archived_at
                        AND h.archived_at <= x.archived_at + INTERVAL 3 SECOND)
WHERE @apply = 1
  AND NOT EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = x.vehicle_id AND h.archived_at = x.archived_at)
  AND EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = x.vehicle_id AND h.archived_at > x.archived_at
                AND h.archived_at <= x.archived_at + INTERVAL 3 SECOND);

-- ---------------------------------------------------------------------------------
-- Part 3: find retry duplicates.
-- Tally fields of shifted rows are read through the repaired column mapping, so the
-- dry run gives the same answer as the real run.
-- ---------------------------------------------------------------------------------
-- Check results of every inspection: archived ones by batch time, active ones with a
-- NULL batch time. (A temporary table can be read only once per query, so this is a
-- second copy next to insp_batch.)
DROP TEMPORARY TABLE IF EXISTS all_checks;
CREATE TEMPORARY TABLE all_checks AS
SELECT vehicle_id, check_id, status, batch_at FROM insp_batch
UNION ALL
SELECT vehicle_id, check_id, status, NULL FROM vehicles_inspection;
CREATE INDEX idx_all_checks ON all_checks (vehicle_id, check_id);

DROP TEMPORARY TABLE IF EXISTS retry_batches;
CREATE TEMPORARY TABLE retry_batches AS
WITH batches AS (
    SELECT h.vehicle_id,
           h.archived_at,
           h.tallied_time,
           CASE WHEN shifted THEN h.manifest_id ELSE h.maker_id END AS maker_id,
           CASE WHEN shifted THEN h.maker_id ELSE h.body_id END AS body_id,
           CASE WHEN shifted THEN h.body_id ELSE h.user_id END AS user_id,
           h.deck_number,
           h.number_of_keys,
           h.key_type,
           LEAD(h.archived_at) OVER (PARTITION BY h.vehicle_id ORDER BY h.archived_at) AS next_archived_at
    FROM (
        SELECT t.*,
               (t.manifest_id IN (SELECT maker_id FROM vehicle_makers)
                AND t.maker_id IN (SELECT body_id FROM vehicle_bodies)) AS shifted
        FROM vehicles_talling_history t
    ) h
),
next_tally AS (
    -- The record that replaced each batch: the next history batch, or the active tally.
    SELECT b.*,
           COALESCE(CASE WHEN nh.shifted THEN nh.manifest_id ELSE nh.maker_id END, vt.maker_id) AS n_maker_id,
           COALESCE(CASE WHEN nh.shifted THEN nh.maker_id ELSE nh.body_id END, vt.body_id) AS n_body_id,
           COALESCE(CASE WHEN nh.shifted THEN nh.body_id ELSE nh.user_id END, vt.user_id) AS n_user_id,
           COALESCE(nh.deck_number, vt.deck_number) AS n_deck_number,
           COALESCE(nh.number_of_keys, vt.number_of_keys) AS n_number_of_keys,
           COALESCE(nh.key_type, vt.key_type) AS n_key_type
    FROM batches b
    LEFT JOIN (
        SELECT t.*,
               (t.manifest_id IN (SELECT maker_id FROM vehicle_makers)
                AND t.maker_id IN (SELECT body_id FROM vehicle_bodies)) AS shifted
        FROM vehicles_talling_history t
    ) nh ON nh.vehicle_id = b.vehicle_id AND nh.archived_at = b.next_archived_at
    LEFT JOIN vehicles_talling vt ON vt.vehicle_id = b.vehicle_id AND b.next_archived_at IS NULL
)
SELECT DISTINCT c.vehicle_id, c.archived_at, c.next_archived_at
FROM next_tally c
WHERE c.n_maker_id IS NOT NULL
  AND c.maker_id = c.n_maker_id
  AND c.body_id = c.n_body_id
  AND (c.user_id = '' OR c.user_id = c.n_user_id)
  AND c.deck_number = c.n_deck_number
  AND c.number_of_keys = c.n_number_of_keys
  AND c.key_type = c.n_key_type
  AND TIMESTAMPDIFF(MINUTE, c.tallied_time, c.archived_at) < @max_gap_minutes
  -- every archived check result also appears, unchanged, in the next inspection
  AND NOT EXISTS (
      SELECT 1 FROM insp_batch hi
      WHERE hi.vehicle_id = c.vehicle_id
        AND hi.batch_at = c.archived_at
        AND NOT EXISTS (
            SELECT 1 FROM all_checks ni
            WHERE ni.vehicle_id = c.vehicle_id
              AND ni.batch_at <=> c.next_archived_at
              AND ni.check_id = hi.check_id AND ni.status = hi.status
        )
  );

SELECT mv.manifest_id,
       COUNT(*) AS duplicate_batches,
       COUNT(DISTINCT r.vehicle_id) AS vehicles_affected
FROM retry_batches r
INNER JOIN manifest_vehicles mv ON mv.vehicle_id = r.vehicle_id
GROUP BY mv.manifest_id
ORDER BY duplicate_batches DESC;

-- (A temporary table can be read only once per query, hence separate SELECTs.)
SELECT COUNT(DISTINCT vehicle_id, archived_at) AS history_batches_total,
       COUNT(DISTINCT vehicle_id) AS vehicles_with_history_now
FROM vehicles_talling_history;

SELECT COUNT(*) AS duplicate_batches_to_delete FROM retry_batches;

SELECT COUNT(DISTINCT h.vehicle_id) AS vehicles_with_history_after
FROM vehicles_talling_history h
LEFT JOIN retry_batches r ON r.vehicle_id = h.vehicle_id AND r.archived_at = h.archived_at
WHERE r.vehicle_id IS NULL;

-- ---------------------------------------------------------------------------------
-- Part 4: delete the duplicate batches, children before parents. This runs only
-- with @apply = 1, after Part 2 aligned every row to its batch time, so rows are
-- matched on the exact archived_at.
-- ---------------------------------------------------------------------------------
DELETE i FROM inspection_image_history i
INNER JOIN vehicles_inspection_history x ON x.inspection_id = i.inspection_id
INNER JOIN retry_batches r
  ON r.vehicle_id = x.vehicle_id
 AND x.archived_at = r.archived_at
WHERE @apply = 1;

DELETE x FROM vehicles_inspection_history x
INNER JOIN retry_batches r
  ON r.vehicle_id = x.vehicle_id
 AND x.archived_at = r.archived_at
WHERE @apply = 1;

DELETE m FROM onboard_packages_media_history m
INNER JOIN onboard_packages_history p ON p.package_id = m.package_id
INNER JOIN retry_batches r
  ON r.vehicle_id = p.vehicle_id
 AND p.archived_at = r.archived_at
WHERE @apply = 1;

DELETE p FROM onboard_packages_history p
INNER JOIN retry_batches r
  ON r.vehicle_id = p.vehicle_id
 AND p.archived_at = r.archived_at
WHERE @apply = 1;

DELETE x FROM inspection_remarks_history x
INNER JOIN retry_batches r
  ON r.vehicle_id = x.vehicle_id
 AND x.archived_at = r.archived_at
WHERE @apply = 1;

DELETE x FROM vehicle_galllery_history x
INNER JOIN retry_batches r
  ON r.vehicle_id = x.vehicle_id
 AND x.archived_at = r.archived_at
WHERE @apply = 1;

DELETE x FROM vehicles_talling_history x
INNER JOIN retry_batches r
  ON r.vehicle_id = x.vehicle_id
 AND x.archived_at = r.archived_at
WHERE @apply = 1;

SELECT IF(@apply = 1, 'APPLIED', 'DRY RUN - nothing changed') AS mode;

COMMIT;
DROP TEMPORARY TABLE IF EXISTS retry_batches;
DROP TEMPORARY TABLE IF EXISTS all_checks;
DROP TEMPORARY TABLE IF EXISTS insp_batch;
