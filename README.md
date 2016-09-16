# ims-release
Imangascans release manager. This software provides an API to manage 
scanlation releases, as well as the ability to download archives and images
corresponding to the release.

# Deoployment

Install composer.

Install composer dependencies with `composer install`.

Init phinx `vendor/bin/phinx init` and modify the config according to
your database setup. Migrate with `vendor/bin/phinx migrate`.