CREATE TABLE `pages` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` BINARY(255) NOT NULL,
  `created_at` TIMESTAMP NOT NULL,
  `release_id` INT UNSIGNED NOT NULL,
FOREIGN KEY(`release_id`) REFERENCES `releases`(`id`),
UNIQUE `path` (`release_id`, `name`),
PRIMARY KEY(`id`)) 
ENGINE=InnoDB DEFAULT CHARSET=utf8;