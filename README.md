# ims-release
Imangascans release manager. This software provides an API to manage 
scanlation releases, as well as the ability to download archives and images
corresponding to the release.

# Deoployment

Install composer dependencies `composer install`
Init phinx `vendor/bin/phinx init` and modify the config according to
your database setup. Migrate `vendor/bin/phinx migrate`