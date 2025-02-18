# bbbstatus

bbbstatus is an utility to view and save events (and their details) from [bbb-webhooks](https://github.com/bigbluebutton/bbb-webhooks)

# Configuration

## Overview

bbbstatus uses a configuration system to manage runtime settings. It supports a default configuration (
`config.toml`) and environment variables to override specific values. If the `config.toml` file is not present, a
default configuration file is generated in the current working directory for later modifications.

Configuration sources:

1. default configuration: hardcoded default settings.
2. config.toml: overrides existing defaults if correctly set.
3. environment variables: overrides specific settings by runtime when provided.

| **Parameter**          | **default value**                                    | **description**                                                  |
|------------------------|------------------------------------------------------|------------------------------------------------------------------|
| `PORT`                 | `8080`                                               | Port on which bbbstatus will run (default 8080).                 |
| `HOST`                 | `localhost`                                          | Hostname or IP address of bbbstatus.                             |
| `DB_CONNECTION_STRING` | `postgres://bbbstatus:bbbstatus@localhost/bbbstatus` | Database connection string (PostgreSQL).                         |
| `CSV_STRUCTURE`        | `time,user,action,text representation`               | Defines the structure of the CSV format for exports.             |
| `SERVE_STATIC_CONTENT` | `true`                                               | Serving static content (`true` or `false`).                      |

## Configuration behavior

### 1. Environment variables

Environment variables will always override default values and `config.toml` settings by runtime.
For example:

- `HOST` set in the environment overwrites the set hostname or ip defined in `config.toml`.

### 2. `config.toml`

If present, `config.toml` settings override the default configuration. If missing, a default `config.toml` is
automatically generated, and you can later modify it to your needs.

### 3. Default values
If `config.toml` is missing or a parameter is not defined, the program falls back to the hardcoded default values.


## Notes

You may serve static files separately in your environment, therefore there is an option to disable it.
bbbstatus will use the path /static in any case.

# Customization

You may override locales, templates, static content (like CSS) and even database migrations at runtime without
recompiling bbbstatus. But be aware that even though this is possible, it may lead to problems, as only the embedded /
official versions are tested.
You may even override multiple files at once. To do that you need to override every file individually, just as described [here](markdown-documentation/Customisation.md).

# Native deployment

To deploy bbbstatus natively (without docker), you can simply run the binary from any location that you want with either a
configuration file or exported environment variables.
Here is an example on how deployment is intended for production use.
Every distro and version **other than debian 12** has limited support.
It is advised to use docker or make a pull request for distro specific requests/modifications.

The native deployment guide assumes that you are logged in as root.

## Install dependencies

This step is depending on your operating system. The example will be given for a Debian 12 installation.

### Install postgres 15

```shell
sudo apt install postgresql-common -y;
```

### Set up a database and a user with write permissions

In your shell of choice run this command

```shell
su postgres;
psql;
```

You should now be in a SQL Shell.
Now create a database and an user, **please DO NOT USE THE PROVIDED PASSWORD**!

```postgresql
CREATE USER bbbstatus WITH PASSWORD 'Your-Creative-And-Long-Password-Here';
CREATE DATABASE bbbstatus OWNER bbbstatus;
```

### Create a system user that will run the service

```sh
useradd -r -s /bin/false bbbstatus
```

### Download bbbstatus into a directory

```shell
mkdir -m 500 "/opt/bbbstatus";
chown bbbstatus:bbbstatus /opt/bbbstatus
curl --output /tmp/bbbstatus-release.zip # TODO: INSERT RELEASE URL HERE.
unzip /opt/bbbstatus/release.zip -d /opt/bbbstatus/
```

### Setting up and enabling a service

```shell
systemctl edit --force --full bbbstatus
```

Copy the following content to the now opened service file:

```ini
[Unit]
Description=bbbstatus is an utility to view and save events from bbb-webhooks
After=network.target

[Service]
Type=simple
User=bbbstatus
Group=bbbstatus
WorkingDirectory=/opt/bbbstatus
ExecStart=/opt/bbbstatus/bbbstatus
Restart=on-failure
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=bbbstatus
Environment=DB_CONNECTION_STRING=postgres://bbbstatus:bbbstatus@localhost/bbbstatus
Environment=HOST=0.0.0.0
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
```

Please adapt the given sample configuration to your needs and values, see [configuration](README.MD#configuration).
You will most definitely need to update the Environment=DB_CONNECTION_STRING line to the database credentials you have setup [here](README.MD#set-up-a-database-and-a-user-with-write-permissions).

Save the changed content (on a fresh installation use `ctrl + O`)

Exit the editor (on a fresh installation use `ctrl + C`)

Now you can start, enable, stop, view logs, restart bbbstatus using the common systemd commands.

# Deployment Docker (minimal)

Make a directory to hold the docker compose file, the .env file and the data folder from bbbstatus and Caddy (the default webserver).

Create a file named `docker-compose.yml` using your favorite text editor, and insert the contents of [the example docker compose file](docker-compose.yml).
Or just simply download the given example of this repository.
Save the changes, and create a new file called `.env`
Insert the contents of the [example file](dotenv) and change the configuration parameters to your needs.
You should change at least the database username, and the database password.
