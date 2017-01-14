CREATE TABLE `members` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` TEXT,
  `biography` TEXT,
  `created_at` TIMESTAMP NOT NULL,
PRIMARY KEY(`id`))
ENGINE=InnoDB DEFAULT CHARSET=utf8;