<?php
namespace Imangascans\Storage;

class DatabaseStorage implements IStorage
{
  /**
   * @var \PDO
   */
  private $pdo;
  public function __construct($di)
  {
    $this->pdo = $di['pdo'];
  }

  public function get($table, $id, $class = '\stdClass')
  {
    $table = $key;
    $query = "SELECT * FROM `$table` WHERE `id` = ?";
    $statement = $this->query($query, [$id]);
    return $statement->fetchObject($class);
  }

  public function delete($table, $id)
  {
    $table = $key;
    $query = "DELETE FROM `$table` WHERE `id` = ? LIMIT 1";
    return $this->query($query, [$id]);
  }

  public function save($table, $object)
  {
    $arr = get_object_vars($object);
    unset($arr['id']);
    $cols = array_keys($arr);

    $query = "INSERT INTO `$table` ";
    $query .= '(`' . implode('`, `', $cols) . '`) ';
    $query .= 'VALUES (:' . implode(', :', $cols) . ')';
    $this->query($query, $arr);
    return $this->pdo->lastInsertId();
  }

  public function update($table, $object)
  {
    $arr = get_object_vars($object);
    $id = $arr['id'];
    unset($arr['id']);
    $cols = array_keys($arr);

    $query = "UPDATE `$table` ";
    foreach ($cols as $col) {
      $query .= "SET `$col` = :$col ";
    }
    $query .= 'WHERE `id` = :id LIMIT 1';

    $arr['id'] = $id;

    return $this->query($query, $arr);
  }

  public function query($query, $args = array())
  {
    $statement = $this->pdo->prepare($query);
    $statement->execute($arr);
    return $statement;
  }
}