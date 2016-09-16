<?php

use Phinx\Migration\AbstractMigration;

class CreateReleasePageTable extends AbstractMigration
{

  public function up()
  {
    $this->execute("CREATE TABLE `release_archive` (
      `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
      `release_id` INT(10) UNSIGNED NOT NULL,
      `size` INT(10) UNSIGNED NOT NULL,
      `contents` LONGBLOB NOT NULL,
      `checksum` INT(10) UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE FOREIGN KEY `release_id` (`release_id`)
      REFERENCES release(`id`)
      ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4");
  }

  public function down()
  {
    $this->execute("DROP TABLE `release_archive`");
  }
}
