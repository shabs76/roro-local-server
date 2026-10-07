-- RORO local server: complete database schema for a fresh installation.
--
-- Structure only, no data. Roles, users and the reference lists (vehicle makers,
-- body types, models, inspection checklist, package types and statuses, clients)
-- are pulled from the remote server when the server starts, as the initial user set
-- in INIT_USER_EMAIL / INIT_USER_PASSWORD (see src/.env.example). Manifests and
-- their vehicles and packages are pulled with the manifest sync.
--
-- Load it into an empty database:
--   mariadb -u root -p roro_local < roro_local.sql
-- With docker compose it is loaded automatically the first time the database
-- volume is created.
--
-- Every table uses CREATE TABLE IF NOT EXISTS and nothing is dropped, so running it
-- against an existing database cannot remove data. The server also applies the
-- files in src/gendb/migrations at start; on this schema they change nothing.
--
-- Generated from the development database plus migrations 001-004 on 2026-10-07.


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `clients` (
  `client_id` varchar(225) NOT NULL,
  `client_name` varchar(100) NOT NULL,
  `logo` varchar(300) NOT NULL DEFAULT 'notset',
  `cover` varchar(300) NOT NULL DEFAULT 'notset',
  `principal` varchar(100) NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'active',
  `client_location` varchar(100) NOT NULL,
  `creation_date` datetime NOT NULL,
  PRIMARY KEY (`client_id`),
  UNIQUE KEY `client_name` (`client_name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `decks_numbers` (
  `deck_id` varchar(225) NOT NULL,
  `manifest_id` varchar(225) NOT NULL,
  `deck_name` varchar(20) NOT NULL,
  `units` int(11) NOT NULL,
  PRIMARY KEY (`deck_id`),
  KEY `manifest_id` (`manifest_id`),
  CONSTRAINT `decks_numbers_ibfk_1` FOREIGN KEY (`manifest_id`) REFERENCES `manifest` (`manifest_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `drivers` (
  `driver_id` varchar(225) NOT NULL,
  `first_name` varchar(100) NOT NULL,
  `last_name` varchar(100) NOT NULL,
  `id_number` varchar(20) NOT NULL,
  `passport` varchar(300) NOT NULL,
  `licence_number` varchar(50) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'active',
  `creation_time` datetime NOT NULL,
  PRIMARY KEY (`driver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `inspection_checklist` (
  `check_id` varchar(225) NOT NULL,
  `check_name` varchar(500) NOT NULL,
  `updated_date` datetime NOT NULL,
  `registered_date` datetime NOT NULL,
  PRIMARY KEY (`check_id`),
  UNIQUE KEY `check_name` (`check_name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `inspection_image` (
  `image_id` varchar(225) NOT NULL,
  `image_link` varchar(300) NOT NULL,
  `inspection_id` varchar(225) NOT NULL,
  `creation_time` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`image_id`),
  KEY `inspection_id` (`inspection_id`),
  CONSTRAINT `inspection_image_ibfk_1` FOREIGN KEY (`inspection_id`) REFERENCES `vehicles_inspection` (`inspection_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `inspection_image_history` (
  `image_id` varchar(225) NOT NULL,
  `image_link` varchar(300) NOT NULL,
  `inspection_id` varchar(225) NOT NULL,
  `creation_time` datetime NOT NULL,
  `archived_at` datetime NOT NULL,
  PRIMARY KEY (`image_id`),
  KEY `inspection_id` (`inspection_id`),
  KEY `idx_hist_inspection_archived` (`inspection_id`,`archived_at`),
  CONSTRAINT `inspection_image_history_ibfk_1` FOREIGN KEY (`inspection_id`) REFERENCES `vehicles_inspection_history` (`inspection_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `inspection_remarks` (
  `remark_id` varchar(225) NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  `remark` text NOT NULL,
  `remark_type` enum('damage','info') NOT NULL,
  `image_link` varchar(300) NOT NULL,
  `remark_time` datetime NOT NULL,
  PRIMARY KEY (`remark_id`),
  UNIQUE KEY `vehicle_id_2` (`vehicle_id`,`remark`) USING HASH,
  KEY `vehicle_id` (`vehicle_id`),
  CONSTRAINT `inspection_remarks_ibfk_1` FOREIGN KEY (`vehicle_id`) REFERENCES `manifest_vehicles` (`vehicle_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `inspection_remarks_history` (
  `remark_id` varchar(225) NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  `remark` text NOT NULL,
  `remark_type` enum('damage','info') NOT NULL,
  `image_link` varchar(300) NOT NULL,
  `remark_time` datetime NOT NULL,
  `archived_at` datetime NOT NULL,
  PRIMARY KEY (`remark_id`),
  KEY `idx_hist_vehicle_archived` (`vehicle_id`,`archived_at`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `logins` (
  `log_id` varchar(225) NOT NULL,
  `login_id` varchar(200) NOT NULL,
  `login_key` varchar(400) NOT NULL,
  `user_id` varchar(225) NOT NULL,
  `status` varchar(20) NOT NULL,
  `expire_date` datetime NOT NULL,
  `login_date` datetime NOT NULL,
  PRIMARY KEY (`log_id`),
  UNIQUE KEY `login_id` (`login_id`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `logins_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `manifest` (
  `manifest_id` varchar(225) NOT NULL,
  `manifest_name` varchar(200) NOT NULL,
  `client_id` varchar(225) NOT NULL,
  `vessel_name` varchar(200) NOT NULL,
  `voyage_no` varchar(50) NOT NULL,
  `berth_no` varchar(20) NOT NULL DEFAULT 'notset',
  `arrival_date` datetime NOT NULL,
  `received_date` date NOT NULL,
  `uploaded_date` datetime NOT NULL,
  PRIMARY KEY (`manifest_id`),
  KEY `client_id` (`client_id`),
  CONSTRAINT `manifest_ibfk_1` FOREIGN KEY (`client_id`) REFERENCES `clients` (`client_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `manifest_packages` (
  `package_id` varchar(225) NOT NULL,
  `package_number` varchar(100) NOT NULL,
  `bl_no` varchar(100) NOT NULL,
  `manifest_id` varchar(225) NOT NULL,
  `description` text NOT NULL,
  `is_inspected` varchar(10) NOT NULL DEFAULT 'no',
  `is_added_later` enum('no','yes') NOT NULL DEFAULT 'no',
  `creation_time` datetime NOT NULL,
  `is_published` enum('no','yes') NOT NULL DEFAULT 'no',
  PRIMARY KEY (`package_id`),
  UNIQUE KEY `package_number` (`package_number`,`manifest_id`),
  KEY `manifest_id` (`manifest_id`),
  CONSTRAINT `manifest_packages_ibfk_1` FOREIGN KEY (`manifest_id`) REFERENCES `manifest` (`manifest_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `manifest_vehicles` (
  `vehicle_id` varchar(225) NOT NULL,
  `manifest_id` varchar(225) NOT NULL,
  `chasis_number` varchar(300) NOT NULL,
  `model` varchar(100) NOT NULL DEFAULT 'notset',
  `description` text NOT NULL,
  `weight` float NOT NULL,
  `bl_no` varchar(100) NOT NULL,
  `creation_date` datetime NOT NULL DEFAULT current_timestamp(),
  `inspection_status` varchar(10) NOT NULL DEFAULT 'no',
  `tallied_status` varchar(10) NOT NULL DEFAULT 'no',
  `discharged_status` varchar(10) NOT NULL DEFAULT 'no',
  `is_overland` varchar(10) NOT NULL DEFAULT 'no',
  `is_added_later` enum('no','yes') NOT NULL DEFAULT 'no',
  `inspection_time` datetime NOT NULL DEFAULT '1000-01-01 00:00:00',
  `tallied_time` datetime NOT NULL DEFAULT '1000-01-01 00:00:00',
  `discharge_time` datetime NOT NULL DEFAULT '1000-01-01 00:00:00',
  `is_published` enum('no','yes') NOT NULL DEFAULT 'no',
  `updated_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) ON UPDATE current_timestamp(3),
  PRIMARY KEY (`vehicle_id`),
  UNIQUE KEY `unique_manifest_chasiss_number` (`manifest_id`,`chasis_number`),
  KEY `idx_mv_manifest_updated` (`manifest_id`,`updated_at`),
  CONSTRAINT `manifest_vehicles_ibfk_1` FOREIGN KEY (`manifest_id`) REFERENCES `manifest` (`manifest_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `media_files` (
  `sha256` char(64) NOT NULL,
  `url` varchar(300) NOT NULL,
  `size` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`sha256`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `media_remote_map` (
  `local_path` varchar(300) NOT NULL,
  `remote_url` varchar(500) NOT NULL,
  `uploaded_at` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`local_path`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `onboard_packages` (
  `package_id` varchar(225) NOT NULL,
  `title` varchar(100) NOT NULL,
  `remark` text NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  PRIMARY KEY (`package_id`),
  KEY `vehicle_id` (`vehicle_id`),
  CONSTRAINT `onboard_packages_ibfk_1` FOREIGN KEY (`vehicle_id`) REFERENCES `manifest_vehicles` (`vehicle_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `onboard_packages_history` (
  `package_id` varchar(225) NOT NULL,
  `title` varchar(100) NOT NULL,
  `remark` text NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  `archived_at` datetime NOT NULL,
  PRIMARY KEY (`package_id`),
  KEY `vehicle_id` (`vehicle_id`),
  KEY `idx_hist_vehicle_archived` (`vehicle_id`,`archived_at`),
  CONSTRAINT `onboard_packages_history_ibfk_1` FOREIGN KEY (`vehicle_id`) REFERENCES `manifest_vehicles` (`vehicle_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `onboard_packages_media` (
  `media_id` varchar(225) NOT NULL,
  `media_type` enum('image','video') NOT NULL,
  `media_link` varchar(150) NOT NULL,
  `package_id` varchar(225) NOT NULL,
  PRIMARY KEY (`media_id`),
  KEY `package_id` (`package_id`),
  CONSTRAINT `onboard_packages_media_ibfk_1` FOREIGN KEY (`package_id`) REFERENCES `onboard_packages` (`package_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `onboard_packages_media_history` (
  `media_id` varchar(225) NOT NULL,
  `media_type` enum('image','video','pdf') NOT NULL,
  `media_link` varchar(150) NOT NULL,
  `package_id` varchar(225) NOT NULL,
  `archived_at` datetime NOT NULL,
  PRIMARY KEY (`media_id`),
  KEY `package_id` (`package_id`),
  KEY `idx_hist_package_archived` (`package_id`,`archived_at`),
  CONSTRAINT `onboard_packages_media_history_ibfk_1` FOREIGN KEY (`package_id`) REFERENCES `onboard_packages_history` (`package_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `package_gallery` (
  `media_id` varchar(225) NOT NULL,
  `media_link` text NOT NULL,
  `media_type` enum('image','video','pdf') NOT NULL,
  `remark` text NOT NULL,
  `package_id` varchar(225) NOT NULL,
  `status` enum('active','deleted','blocked') NOT NULL,
  PRIMARY KEY (`media_id`),
  KEY `package_id` (`package_id`),
  CONSTRAINT `package_gallery_ibfk_1` FOREIGN KEY (`package_id`) REFERENCES `manifest_packages` (`package_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `package_types` (
  `type_id` varchar(225) NOT NULL,
  `type_name` varchar(100) NOT NULL,
  `status` varchar(50) NOT NULL,
  `created_time` datetime NOT NULL,
  PRIMARY KEY (`type_id`),
  UNIQUE KEY `type_name` (`type_name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `packages_inspection` (
  `inspection_id` varchar(225) NOT NULL,
  `package_id` varchar(225) NOT NULL,
  `type_id` varchar(225) NOT NULL,
  `picture` varchar(300) NOT NULL,
  `inspection_status` varchar(225) NOT NULL,
  `user_id` varchar(225) NOT NULL,
  `inspection_time` datetime NOT NULL,
  `creation_time` datetime NOT NULL,
  PRIMARY KEY (`inspection_id`),
  UNIQUE KEY `package_id` (`package_id`),
  KEY `type_id` (`type_id`),
  KEY `user_id` (`user_id`),
  KEY `inspection_status` (`inspection_status`),
  CONSTRAINT `packages_inspection_ibfk_1` FOREIGN KEY (`package_id`) REFERENCES `manifest_packages` (`package_id`) ON UPDATE CASCADE,
  CONSTRAINT `packages_inspection_ibfk_2` FOREIGN KEY (`type_id`) REFERENCES `package_types` (`type_id`) ON UPDATE CASCADE,
  CONSTRAINT `packages_inspection_ibfk_3` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON UPDATE CASCADE,
  CONSTRAINT `packages_inspection_ibfk_4` FOREIGN KEY (`inspection_status`) REFERENCES `packages_inspection_status` (`status_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `packages_inspection_status` (
  `status_id` varchar(225) NOT NULL,
  `status_name` varchar(100) NOT NULL,
  `description` text NOT NULL,
  `status_number` int(6) NOT NULL,
  PRIMARY KEY (`status_id`),
  UNIQUE KEY `status_name` (`status_name`),
  UNIQUE KEY `status_number` (`status_number`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `ramp_vehicle_count` (
  `count_id` varchar(50) NOT NULL,
  `manifest_id` varchar(225) NOT NULL,
  `number` int(11) NOT NULL,
  `count_time` datetime NOT NULL,
  PRIMARY KEY (`count_id`),
  KEY `manifest_id` (`manifest_id`),
  CONSTRAINT `ramp_vehicle_count_ibfk_1` FOREIGN KEY (`manifest_id`) REFERENCES `manifest` (`manifest_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `roles` (
  `role_id` varchar(225) NOT NULL,
  `role_name` varchar(50) NOT NULL,
  `role_number` smallint(6) NOT NULL,
  `role_date` datetime NOT NULL,
  PRIMARY KEY (`role_id`),
  UNIQUE KEY `role_name` (`role_name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `statement_of_fact_events` (
  `event_id` varchar(225) NOT NULL,
  `shift_id` varchar(225) NOT NULL,
  `event_name` varchar(300) NOT NULL,
  `event_remarks` text NOT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  PRIMARY KEY (`event_id`),
  KEY `shift_id` (`shift_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `statement_of_fact_shift` (
  `shift_id` varchar(225) NOT NULL,
  `manifest_id` varchar(225) NOT NULL,
  `shift_name` varchar(300) NOT NULL,
  `number_of_gangs` int(5) NOT NULL,
  `shift_start_time` datetime NOT NULL,
  `shift_end_time` datetime NOT NULL,
  PRIMARY KEY (`shift_id`),
  KEY `manifest_id` (`manifest_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `subscribers` (
  `subscription_id` varchar(225) NOT NULL,
  `email` varchar(200) NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'active',
  `subscription_time` datetime NOT NULL,
  PRIMARY KEY (`subscription_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `sync_contribution` (
  `contribution_id` varchar(50) NOT NULL,
  `member_id` varchar(50) NOT NULL,
  `chassis_number` varchar(50) NOT NULL,
  `deck_id` varchar(50) NOT NULL,
  `is_damaged` varchar(20) NOT NULL DEFAULT 'no',
  `sync_time` datetime NOT NULL,
  PRIMARY KEY (`contribution_id`),
  UNIQUE KEY `chassis_number` (`chassis_number`),
  KEY `deck_id` (`deck_id`),
  KEY `member_id` (`member_id`),
  CONSTRAINT `sync_contribution_ibfk_1` FOREIGN KEY (`deck_id`) REFERENCES `decks_numbers` (`deck_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `sync_contribution_ibfk_2` FOREIGN KEY (`member_id`) REFERENCES `sync_member` (`member_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `sync_member` (
  `member_id` varchar(50) NOT NULL,
  `user_id` varchar(225) NOT NULL,
  `manifest_id` varchar(225) NOT NULL,
  `last_seen` datetime NOT NULL,
  `creation_time` datetime NOT NULL,
  PRIMARY KEY (`member_id`),
  UNIQUE KEY `user_id` (`user_id`,`manifest_id`),
  KEY `manifest_id` (`manifest_id`),
  CONSTRAINT `sync_member_ibfk_1` FOREIGN KEY (`manifest_id`) REFERENCES `manifest` (`manifest_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `sync_member_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `tablet_compare_reports` (
  `report_id` varchar(64) NOT NULL,
  `manifest_id` varchar(225) NOT NULL,
  `device_id` varchar(200) NOT NULL,
  `device_name` varchar(200) NOT NULL DEFAULT '',
  `app_version` varchar(50) NOT NULL DEFAULT '',
  `user_id` varchar(225) NOT NULL DEFAULT '',
  `summary_json` text NOT NULL,
  `items_json` mediumtext NOT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`report_id`),
  KEY `idx_tcr_manifest_device` (`manifest_id`,`device_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `tally_inspection_gate` (
  `gate_id` varchar(225) NOT NULL COMMENT 'vehicle_id as recorded on local database to prevent double upload',
  `manifest_id` varchar(225) NOT NULL,
  `chassis_number` varchar(300) NOT NULL,
  `creation_time` datetime NOT NULL,
  PRIMARY KEY (`gate_id`),
  KEY `manifest_id` (`manifest_id`),
  KEY `chassis_number` (`chassis_number`),
  CONSTRAINT `tally_inspection_gate_ibfk_1` FOREIGN KEY (`manifest_id`) REFERENCES `manifest` (`manifest_id`) ON UPDATE CASCADE,
  CONSTRAINT `tally_inspection_gate_ibfk_2` FOREIGN KEY (`chassis_number`) REFERENCES `manifest_vehicles` (`vehicle_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `users` (
  `user_id` varchar(225) NOT NULL,
  `fname` varchar(200) NOT NULL,
  `lname` varchar(200) NOT NULL,
  `email` varchar(200) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `password` varchar(400) NOT NULL,
  `role` varchar(225) NOT NULL,
  `status` varchar(20) NOT NULL,
  `creation_date` datetime NOT NULL,
  PRIMARY KEY (`user_id`),
  KEY `role` (`role`),
  CONSTRAINT `users_ibfk_1` FOREIGN KEY (`role`) REFERENCES `roles` (`role_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `vehicle_bodies` (
  `body_id` varchar(225) NOT NULL,
  `body_name` varchar(200) NOT NULL,
  `creation_time` datetime NOT NULL,
  PRIMARY KEY (`body_id`),
  UNIQUE KEY `body_name` (`body_name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `vehicle_discharge_tally` (
  `discharge_id` varchar(225) NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  `discharge_image` varchar(300) NOT NULL,
  `driver_id` varchar(225) NOT NULL,
  `user_id` varchar(225) NOT NULL,
  `discharge_time` datetime NOT NULL,
  PRIMARY KEY (`discharge_id`),
  UNIQUE KEY `vehicle_id` (`vehicle_id`),
  KEY `user_id` (`user_id`),
  KEY `vehicle_discharge_tally_ibfk_1` (`vehicle_id`),
  KEY `driver_id` (`driver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `vehicle_galllery` (
  `media_id` varchar(225) NOT NULL,
  `media_link` text NOT NULL,
  `media_type` enum('image','pdf','video') NOT NULL,
  `remark` text NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  `status` enum('active','blocked','deleted') NOT NULL,
  PRIMARY KEY (`media_id`),
  KEY `vehicle_id` (`vehicle_id`),
  CONSTRAINT `vehicle_galllery_ibfk_1` FOREIGN KEY (`vehicle_id`) REFERENCES `manifest_vehicles` (`vehicle_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `vehicle_galllery_history` (
  `media_id` varchar(225) NOT NULL,
  `media_link` text NOT NULL,
  `media_type` enum('image','video','pdf') NOT NULL,
  `remark` text NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  `status` enum('active','deleted','blocked') NOT NULL,
  `archived_at` datetime NOT NULL,
  PRIMARY KEY (`media_id`),
  KEY `idx_hist_vehicle_archived` (`vehicle_id`,`archived_at`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `vehicle_makers` (
  `maker_id` varchar(225) NOT NULL,
  `maker_name` varchar(200) NOT NULL,
  `creation_time` datetime NOT NULL,
  PRIMARY KEY (`maker_id`),
  UNIQUE KEY `maker_name` (`maker_name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `vehicle_models` (
  `model_id` varchar(225) NOT NULL,
  `model_name` varchar(200) NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'active',
  `created_time` datetime NOT NULL,
  PRIMARY KEY (`model_id`),
  UNIQUE KEY `model_name` (`model_name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `vehicles_inspection` (
  `inspection_id` varchar(225) NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  `check_id` varchar(225) NOT NULL,
  `status` varchar(50) NOT NULL,
  `check_time` datetime NOT NULL,
  PRIMARY KEY (`inspection_id`),
  UNIQUE KEY `unique_vehicle_check` (`vehicle_id`,`check_id`),
  KEY `check_id` (`check_id`),
  CONSTRAINT `vehicles_inspection_ibfk_1` FOREIGN KEY (`check_id`) REFERENCES `inspection_checklist` (`check_id`) ON UPDATE CASCADE,
  CONSTRAINT `vehicles_inspection_ibfk_2` FOREIGN KEY (`vehicle_id`) REFERENCES `manifest_vehicles` (`vehicle_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `vehicles_inspection_history` (
  `inspection_id` varchar(225) NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  `check_id` varchar(225) NOT NULL,
  `status` varchar(50) NOT NULL,
  `check_time` datetime NOT NULL,
  `archived_at` datetime NOT NULL,
  PRIMARY KEY (`inspection_id`),
  KEY `vehicle_id` (`vehicle_id`),
  KEY `check_id` (`check_id`),
  KEY `idx_hist_vehicle_archived` (`vehicle_id`,`archived_at`),
  CONSTRAINT `vehicles_inspection_history_ibfk_1` FOREIGN KEY (`vehicle_id`) REFERENCES `manifest_vehicles` (`vehicle_id`) ON UPDATE CASCADE,
  CONSTRAINT `vehicles_inspection_history_ibfk_2` FOREIGN KEY (`check_id`) REFERENCES `inspection_checklist` (`check_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `vehicles_talling` (
  `tally_id` varchar(225) NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  `maker_id` varchar(225) NOT NULL,
  `body_id` varchar(225) NOT NULL,
  `user_id` varchar(225) NOT NULL,
  `image_link` varchar(300) NOT NULL,
  `deck_number` varchar(10) NOT NULL,
  `number_of_keys` int(4) NOT NULL DEFAULT 0,
  `key_type` varchar(50) NOT NULL DEFAULT 'Not specified',
  `tallied_time` datetime NOT NULL,
  `manifest_id` varchar(225) NOT NULL DEFAULT '',
  `submission_id` varchar(64) DEFAULT NULL,
  PRIMARY KEY (`tally_id`),
  UNIQUE KEY `vehicle_id` (`vehicle_id`),
  UNIQUE KEY `uq_vt_submission` (`submission_id`),
  KEY `body_id` (`body_id`),
  KEY `maker_id` (`maker_id`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `vehicles_talling_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON UPDATE CASCADE,
  CONSTRAINT `vehicles_talling_ibfk_2` FOREIGN KEY (`vehicle_id`) REFERENCES `manifest_vehicles` (`vehicle_id`) ON UPDATE CASCADE,
  CONSTRAINT `vehicles_talling_ibfk_3` FOREIGN KEY (`maker_id`) REFERENCES `vehicle_makers` (`maker_id`) ON UPDATE CASCADE,
  CONSTRAINT `vehicles_talling_ibfk_4` FOREIGN KEY (`body_id`) REFERENCES `vehicle_bodies` (`body_id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE IF NOT EXISTS `vehicles_talling_history` (
  `tally_id` varchar(225) NOT NULL,
  `vehicle_id` varchar(225) NOT NULL,
  `manifest_id` varchar(225) NOT NULL,
  `maker_id` varchar(225) NOT NULL,
  `body_id` varchar(225) NOT NULL,
  `image_link` varchar(300) NOT NULL,
  `deck_number` varchar(10) NOT NULL,
  `number_of_keys` int(4) NOT NULL DEFAULT 0,
  `key_type` varchar(50) NOT NULL,
  `tallied_time` datetime NOT NULL,
  `archived_at` datetime NOT NULL,
  `user_id` varchar(225) NOT NULL DEFAULT '',
  `submission_id` varchar(64) DEFAULT NULL,
  PRIMARY KEY (`tally_id`),
  KEY `idx_hist_vehicle_archived` (`vehicle_id`,`archived_at`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

