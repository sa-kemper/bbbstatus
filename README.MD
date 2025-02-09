# bbbstatus

bbbstatus is a utility to safe and view events from [bbb-webhooks](https://github.com/bigbluebutton/bbb-webhooks)

# Configuration

## Overview

bbbstatus uses a configuration system to manage runtime settings. It supports both a default configuration (
`config.toml`) and environment variables to override specific values. If the `config.toml` file is not present, a
default configuration file is generated in the current working directory.

Configuration Sources:

1. Default Configuration: Hardcoded default settings.
2. config.toml: Overrides defaults if present.
3. Environment Variables: Overrides specific settings when provided.

| **Parameter**          | **Default Value**                                    | **Description**                                                  |
|------------------------|------------------------------------------------------|------------------------------------------------------------------|
| `PORT`                 | `8080`                                               | The port on which bbbstatus will run.                            |
| `HOST`                 | `localhost`                                          | The hostname or IP address of bbbstatus.                         |
| `DB_CONNECTION_STRING` | `postgres://bbbstatus:bbbstatus@localhost/bbbstatus` | Database connection string (PostgreSQL).                         |
| `CSV_STRUCTURE`        | `time,user,action,text representation`               | Defines the structure of the CSV format expected by the program. |
| `SERVE_STATIC_CONTENT` | `true`                                               | Whether to serve static content (`true` or `false`).             |

## Configuration Behavior

### 1. Default Values

If `config.toml` is missing or a parameter is not defined, the program falls back to default values.

### 2. Environment Variables

Environment variables can override default values and `config.toml` settings. For example:

- `HOST` set in the environment takes precedence over the `HOST` defined in `config.toml`.

### 3. `config.toml`

If present, `config.toml` settings override the default configuration. If missing, a default `config.toml` is
automatically generated.

## Notes

You may serve static files separately in you environment, therefore there is an option to disable it. bbbstatus will use
the path /static in any case.

# Customization

You may override locales, templates, static content (like CSS) and even database migrations at runtime without
recompiling bbbstatus. Be advised that even though this is possible, it may lead to problems, as only the embedded /
official versions are tested. You may override multiple files at once. to do that you need to override every file
individually, just as described [here](markdown-documentation/Customisation.md).

To read on how to do this customisation read [here](markdown-documentation/Customisation.md)

# Deployment Native

For native deployment, you technically just have to run the binary from any location that you want with either a
configuration file or exported environment variables but here is an example on how deployment is intended for production
use.
Every distro and version **other than debian 12 has limited support.** Use docker or make a pull request for distro
specific documentation.

The Native deployment guide assumes that you are logged in as root.

## install dependencies

This step will be different depending on your operating system. In this deployment example we will assume a debian 12
installation.

Install postgres 15

```shell
sudo apt install postgresql-common -y;
```

## Set up a database and a user with write permissions

In your shell of choice run this command

```shell
su postgres;
psql;
```

you should now be in a SQL Shell. create a database and a user, **please DO NOT USE THE PROVIDED PASSWORD**.

```postgresql
CREATE USER bbbstatus WITH PASSWORD 'Your-Creative-And-Long-Password-Here';
CREATE DATABASE bbbstatus OWNER bbbstatus;
```

## Create a system user that will run the service

```sh
sudo useradd -r -s /bin/false bbbstatus
```

## Install bbbstatus into a directory

```shell
mkdir -m 500 "/opt/bbbstatus";
chown bbbstatus:bbbstatus /opt/bbbstatus
curl --output /tmp/bbbstatus-release.zip # TODO: INSERT RELEASE URL HERE.
unzip /opt/bbbstatus/release.zip -d /opt/bbbstatus/
```

## Set up a service file and enable the service

```shell
systemctl edit --force --full bbbstatus
```

insert this service file template:

```ini
[Unit]
Description=bbbstatus is a utility to safe and view events from bbb-webhooks
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

You may configure more or different configuration options at this point, see [configuration](README.MD#configuration).
You will most definitely need to update the Environment=DB_CONNECTION_STRING line, to the database credentials you've
setup [here](README.MD#set-up-a-database-and-a-user-with-write-permissions).

Save the changed content (on a fresh installation use `ctrl + O`)

Exit the editor (on a fresh installation use `ctrl + C`)

Now you can start, enable, stop, view logs, restart bbbstatus using the systemd utils.

# Deployment Docker (minimal)

I will assume that you have docker setup and running, you have the privileges to use docker compose and run containers.
I'm going to assume that you do not have a requirement to use docker secrets.

Make a directory to hold the docker compose file, the .env file and the data folder.

Create a file named `docker-compose.yml` using your favorite text editor, and insert the contents
of [the example docker compose file](docker-compose.yml). Save the changes, and create a new file called `.env` insert
the contents of the [example file](dotenv) and change the configuration parameters to your needs. you should change the
database username, and the database password.

# Deployment Docker full
