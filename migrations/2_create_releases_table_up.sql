CREATE TABLE `releases` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `identifier` BINARY(10) NOT NULL,
  `version` INT NOT NULL,
  `status` INT UNSIGNED NOT NULL,
  `released_on` TIMESTAMP NOT NULL,
  `project_id` INT UNSIGNED NOT NULL,
FOREIGN KEY(`project_id`) REFERENCES `projects`(`id`),
UNIQUE `version` (`identifier`, `version`),
PRIMARY KEY(`id`))
ENGINE=InnoDB DEFAULT CHARSET=utf8;