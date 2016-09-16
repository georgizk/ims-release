<?php
namespace Imangascans\Model;

class Release
{
  const TABLE = 'release';

  public $id;
  public $project_id;
  public $name;
  public $date;
  public $version;
  public $status;
}
