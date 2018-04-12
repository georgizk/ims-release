# ims-release

The [Imangascans](https://imangascans.org/) manga release API.

Its core features are:

* A simple REST API
* Endpoints for downloading release archives and pages
* Endpoints for creating and managing projects
* Endpoints for creating and managing releases under projects
* Endpoints for creating and managing pages under releases

Upcoming features including:

* Support for managing information about staff

## Usage
```
go get
go install
$GOBIN/ims-release config.json
```
## Setup

### Golang

1. Install Go version 1.6.3 or later
2. Set the GOPATH environment variable
3. Set the GOBIN environment variable to $GOPATH/bin
4. Place ims-release inside $GOPATH/src

Example:
```
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
mkdir -p $GOBIN
mkdir -p $GOPATH/src
cd $GOPATH/src
git clone https://github.com/georgizk/ims-release.git
cd ims-release
go get
go install
```

### MySQL

1. Install MySQL or MariaDB server
2. Start the MySQL server
3. Create and configure a new database

Example:
```
mysql -u root -p # This will bring you to the mysql prompt, prefixed by "mysql>"
mysql> create database ims_release;
mysql> grant all privileges on ims_release.* to 'ims'@'localhost' identified by 'password1';
mysql> set global sql_mode = 'NO_ENGINE_SUBSTITUTION';
```

### Configuration

Configuration is done via a json config file. It must contain the following fields:

* `bindAddress` - tcp bind address. The format is `<host>:<port>`.
* `imageDirectory` - path where images are stored.
* `dbProtocol` - database connection protocol - "tcp" or "unix".
* `dbAddress` - database connection address e.g. "127.0.0.1:3306" or "/var/run/mysqld/mysqld.sock".
* `dbName` - database name.
* `dbUser` - database user.
* `dbPassword` - database password.
* `authToken` - the secret authentication token used to authenticate `POST`, `PUT` and `DELETE` requests.

Refer to `config.json.example`.
