<?php
namespace Imangascans\Storage;

interface IStorage
{
  /**
   * Return the object stored against the given key and id.
   * @param string $key the key that identifies the type of object
   * @param int $id the id that uniquely identifies the object for the given key
   * @param string $class the type of object to return
   */
  public function get($key, $id, $class = '\stdClass');

  /**
   * Remove the object stored against the given key and id.
   * @param string $key the key that identifies the type of object
   * @param int $id the id that uniquely identifies the object for the given key
   */
  public function delete($key, $id);

  /**
   * Store a new object against the given key.
   * @param string $key the key that identifies the type of object
   * @param unknown $object a PHP object to store
   * @return id the unique identifier of the object
   */
  public function save($key, $object);

  /**
   * Update the object stored against the given key and id.
   * @param string $key the key that identifies the type of object
   * @param unknown $object a PHP object to store
   */
  public function update($key, $object);
}