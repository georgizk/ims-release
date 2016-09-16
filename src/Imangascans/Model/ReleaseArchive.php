<?php
namespace Imangascans\Model;

class Release
{
  const TABLE = 'release_archive';

  public $id;
  public $release_id;
  public $size;
  public $contents;
  public $checksum;
}
