ALTER TABLE `releases` DROP FOREIGN KEY `releases_ibfk_1`;
DROP INDEX `version` ON `releases`;
CREATE UNIQUE INDEX `version` ON `releases` (`identifier`, `version`);
ALTER TABLE `releases` ADD CONSTRAINT `releases_ibfk_1` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`);