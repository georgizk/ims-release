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

/**
 * RESTful API with the following conventions:
 *           Action         Collection(e.g. /projects)  Item(e.g. /projects{id})
 *   POST    Create         201, link to item           404
 *   GET     Read           200, list of ids            200 if found, 404 if not
 *   PUT     Update/Replace 404                         200 or 404
 *   DELETE  Delete         404                         200 or 404
 *
 * Parents hold the children, so for example we can have the following paths:
 *   /projects
 *   /projects/5
 *   /projects/5/releases
 *   /projects/5/releases/2
 *   /projects/5/releases/2/pages
 *   /projects/5/releases/2/archive
 */
$app->group('/api', function() {
  $this->group('/projects', function() {
    $this->get('', function($request, $response) {
      $response->getBody()->write('List of project ids');
      return $response;
    });
    $this->post('', function($request, $response) {
      $response->getBody()->write('Create project');
      return $response;
    });
    $this->group('/{project_id:[0-9]+}', function() {
      $this->get('', function($request, $response, $args) {
        $response->getBody()->write('Project with id ' . $args['project_id']);
        return $response;
      });
      $this->put('', function($request, $response, $args) {
        $response->getBody()->write('Update project with id ' . $args['project_id']);
        return $response;
      });
      $this->delete('', function($request, $response, $args) {
        $response->getBody()->write('Delete project with id ' . $args['project_id']);
        return $response;
      });

      $this->group('/releases', function() {
        $this->get('', function($request, $response, $args) {
          $response->getBody()->write('List of release ids for project ' . $args['project_id']);
          return $response;
        });
        $this->post('', function($request, $response, $args) {
          $response->getBody()->write('Create release under project ' . $args['project_id']);
          return $response;
        });

        $this->group('/{release_id:[0-9]+}', function() {
          $this->get('', function($request, $response, $args) {
            $response->getBody()->write('Release with id ' . $args['release_id']);
            return $response;
          });

          $this->put('', function($request, $response, $args) {
            $response->getBody()->write('Update release with id ' . $args['release_id']);
            return $response;
          });

          $this->delete('', function($request, $response, $args) {
            $response->getBody()->write('Delete release with id ' . $args['release_id']);
            return $response;
          });

          $this->group('/pages', function() {
            $this->get('', function($request, $response, $args) {
              $response->getBody()->write('Page ids for release id ' . $args['release_id']);
              return $response;
            });
            $this->post('', function($request, $response, $args) {
              $response->getBody()->write('Create page under release ' . $args['release_id']);
              return $response;
            });

            $this->group('/{page_id:[0-9]+}', function() {
              $this->get('', function($request, $response, $args) {
                $response->getBody()->write('Get page ' . $args['page_id']);
                return $response;
              });
              $this->put('', function($request, $response, $args) {
                $response->getBody()->write('Update page ' . $args['page_id']);
                return $response;
              });
              $this->delete('', function($request, $response, $args) {
                $response->getBody()->write('Delete page ' . $args['page_id']);
                return $response;
              });
            });
          });

          $this->group('/archive', function() {
            $this->get('', function($request, $response, $args) {
              $response->getBody()->write('Get archive for release ' . $args['release_id']);
              return $response;
            });
            $this->post('', function($request, $response, $args) {
              // this should take a list of page ids, even though implicitly
              // it would be the page ids for the release
              $response->getBody()->write('Create archive for release ' . $args['release_id']);
              return $response;
            });

            $this->delete('', function($request, $response, $args) {
              $response->getBody()->write('Delete archive for release ' . $args['release_id']);
              return $response;
            });
          });
        });
      });
    });
  });
});

$app->run();
