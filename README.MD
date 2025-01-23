# BBBStatus

BBBStatus is a utility to safe and view events from [bbb-webhooks](https://github.com/bigbluebutton/bbb-webhooks)

# Configuration

## Overview

BBBStatus uses a configuration system to manage runtime settings. It supports both a default configuration (
`config.toml`) and environment variables to override specific values. If the `config.toml` file is not present, a
default configuration file is generated in the current working directory.

Configuration Sources:

1. Default Configuration: Hardcoded default settings.
2. config.toml: Overrides defaults if present.
3. Environment Variables: Overrides specific settings when provided.

| **Parameter**          | **Default Value**                                    | **Description**                                                  |
|------------------------|------------------------------------------------------|------------------------------------------------------------------|
| `PORT`                 | `8080`                                               | The port on which the application will run.                      |
| `HOST`                 | `localhost`                                          | The hostname or IP address of the application.                   |
| `DB_CONNECTION_STRING` | `postgres://myapp:coolAppPassword!1@localhost/appdb` | Database connection string (PostgreSQL).                         |
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

You may serve static files separately in you environment, therefore there is a option to disable it. BBBStatus will use
the path /static in any case.

# Customisation

You may override locales, templates, static content (like CSS) and even database migrations at runtime without
recompiling BBBStatus. Be advised that even though this is possible, it may lead to problems, as only the embedded /
offical versions are tested. You may override multiple files at once. to do that you need to override every file
individually, just as described [here](markdown-documentation/Customisation.md).

To read on how to do this customisation read [here](markdown-documentation/Customisation.md)

# Deployment Native

For native deployment, you technically just have to run the binary from any location that you want with either a
configuration file or exported environment variables but here is a example on how deployment is intended for production
use.

## install dependencies

This step will be different depending on your operating system. In this deployment example we will assume a debian 12
install.

Install postgres 15

```shell
sudo apt install postgresql-common -y;
```

## Setup a database and a user with write permissions

In your shell of choice run this command

```shell
sudo -u postgres psql;
```

you should now be in a SQL Shell. create a database and a user, **please DO NOT USE THE PROVIDED PASSWORD**.

```postgresql
CREATE DATABASE bbbstatus;
CREATE USER bbbstatus_user WITH PASSWORD 'Your-Creative-And-Long-Password-Here';
GRANT ALL PRIVILEGES ON DATABASE bbbstatus TO bbbstatus_user;
```

## Create a system user that will run the service

```sh
sudo useradd -r -s /bin/false bbbstatus
```

## Install BBBStatus into a directory

```shell
sudo mkdir -m 500 "/opt/bbbstatus";
sudo chown bbbstatus:bbbstatus /opt/bbbstatus
sudo curl --output /tmp/bbbstatus-release.zip # TODO: INSERT RELEASE URL HERE.
sudo unzip /opt/bbbstatus/release.zip -d /opt/bbbstatus/
```

## Setup a service file and enable the service

```shell
systemctl edit --force --full bbbstatus
```

insert this service file template:

```ini
[Unit]
Description=BBBStatus is a utility to safe and view events from bbb-webhooks
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
Environment=DB_CONNECTION_STRING=postgres://bbbstatus_user:securepassword@localhost/bbbstatus
Environment=HOST=0.0.0.0
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
```

You may configure more or different configuration options at this point, see [configuration](README.MD#configuration).
You will most defenetly need to update the Environment=DB_CONNECTION_STRING line, to the database credentials you've
setup [here](README.MD#setup-a-database-and-a-user-with-write-permissions).

Save the changed content (on a fresh installation use `ctrl + O`)

Exit the editor (on a fresh installation use `ctrl + C`)

Now you can start, enable, stop, view logs, restart BBBStatus using the systemd utils.

# Deployment Docker (minimal)

# Deployment Docker full
