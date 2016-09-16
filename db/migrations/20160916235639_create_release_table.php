<?php

use Phinx\Migration\AbstractMigration;

class CreateReleaseTable extends AbstractMigration
{
  public function up()
  {
    $this->execute("CREATE TABLE `release` (
      `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
      `project_id` INT(10) UNSIGNED NOT NULL,
      `name` VARCHAR(10) NOT NULL,
      `date` DATETIME NOT NULL,
      `version` INT(10) UNSIGNED NOT NULL,
      `status` INT(10) UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `name` (`name`),
    INDEX `date` (`date`),
    INDEX `status` (`status`),
    FOREIGN KEY `project_id` (`project_id`)
      REFERENCES project(`id`)
      ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4");
  }

  public function down()
  {
    $this->execute("DROP TABLE `release`");
  }
}
