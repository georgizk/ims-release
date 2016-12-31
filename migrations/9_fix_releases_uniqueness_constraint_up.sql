DROP INDEX `version` ON `releases`;
CREATE UNIQUE INDEX `version` ON `releases` (`project_id`, `identifier`, `version`);