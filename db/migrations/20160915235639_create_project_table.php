<?php

use Phinx\Migration\AbstractMigration;

class CreateProjectTable extends AbstractMigration
{
  public function up()
  {
    $this->execute("CREATE TABLE `project` (
      `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
      `name` VARCHAR(128) NOT NULL,
      `filename` VARCHAR(128) NOT NULL,
      `description` TEXT NOT NULL,
      `status` INT(10) UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `name` (`name`),
    UNIQUE KEY `filename` (`filename`),
    INDEX `status` (`status`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4");
  }

  public function down()
  {
    $this->execute("DROP TABLE `project`");
  }
}
