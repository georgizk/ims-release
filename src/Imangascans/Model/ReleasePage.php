<?php
namespace Imangascans\Model;

class Release
{
  const TABLE = 'release_page';

  public $id;
  public $release_id;
  public $name;
  public $size;
  public $type;
  public $contents;
  public $checksum;
}
