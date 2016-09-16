<?php
define('BASE_ROOT', __DIR__);
require_once BASE_ROOT . '/vendor/autoload.php'; // set up autoloading

$di = new \Slim\Container();

$di['pdo'] = function($di) {
  $dsn = sprintf("mysql:dbname=%s;unix_socket=%s;charset=utf8mb4",
    $di['config']['db.name'],
    $di['config']['db.sock']);

  $pdo = new \PDO(
    $dsn,
    $di['config']['db.user'],
    $di['config']['db.pass'],
    array(
      \PDO::ATTR_PERSISTENT => false,
      \PDO::ATTR_ERRMODE => \PDO::ERRMODE_EXCEPTION
    )
  );
  return $pdo;
};
$app = new \Slim\App($di);
$app->group('/api', function() {
  $this->group('/project', function() {
    $this->get('/list', function($request, $response) {
      $response->getBody()->write('Test');
      return $response;
    });
  });
});

$app->run();
