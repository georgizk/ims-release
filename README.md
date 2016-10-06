# ims-release

A manga release API for the [IMangaScans](https://imangascans.org/) scanlation group.

Its core features are:

* A simple REST API
* Endpoints for downloading release archives and pages
* Endpoints for creating and managing projects
* Endpoints for creating and managing releases under projects
* Endpoints for creating and managing pages under releases

Upcoming features including:

* Support for managing information about staff

## Configuration

To configure the API service, you can edit the `config/config.json` file. It must contain the following fields:

* `address` - The address to bind the server to in the `<ip>:<port>` format. The `ip` should usually be `0.0.0.0`.
* `imageDirectory` - The path to the directory (folder) that pages of manga should be saved to. An absolute pathi is best.
* `database` - The connection string required to connect to the MySQL database.

The format for the `database` parameter is specified in the [SQL driver library](https://github.com/go-sql-driver/mysql#dsn-data-source-name)'s
documentation.

```
[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
```

## Development

To develop IMS-Release, you will need a few things set up. The steps below will guide you through the most simple setup for local development.

### Setup

#### Golang

First, download Go from the [official site](https://golang.org/dl/).  After you've downloaded it, extract and run the installer.  Next, you'll want to set the `GOROOT` environment variable globablly for your user.  The recommended way is to add the following to either your `$HOME/.profile`, `$HOME/.bashrc`, or `$HOME/.zshrc` file.

```
GOROOT = /usr/local/go
```

Assuming you installed Go to the default directory (above). You can verify that your install worked by running

```
go version
```

You should see output like `go version go1.7.1 darwin/amd64` or something else to match your system.

#### MySQL

You should install mysql via whatever package manager you happen to use.  On macOS, this would preferrably be [Homebrew](http://brew.sh/), or else the usual apt/yum/pacman if you are running Linux. Once installed, start the mysql server by running

```
mysql.server start
```

or

```
service mysql start
```

Whichever works for your system.  Then, connect to your database server with root permissions and initialize a test database.

```
sudo mysql # This will bring you to the mysql prompt, prefixed by "mysql>"
mysql> create database testing;
mysql> grant all privileges on testing.* to 'tester'@'localhost' identified by 'password1';
```

#### Config

The last thing to do is to make sure your `config/config.json` file has the right data.  For this local setup, it should contain the following.

```json
{
  "address": "0.0.0.0:3000",
  "imageDirectory": "./images",
  "database": "tester:password1@/testing"
}
```

Lastly, make sure you create a the folder that images will be saved to.

```
mkdir images
```

### Running

Once your setup is complete, you can get the API server running by first building and then executing the server binary with the following commands, run from the `ims-release/` base directory.

```
go build
./ims-release
```

You should see output like `Listening on 0.0.0.0:3000`.