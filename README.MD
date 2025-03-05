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

| **Parameter**          | **default value**                                    | **description**                                                                                                                                                            | Section        |
|------------------------|------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------|
| `PORT`                 | `8080`                                               | Port on which bbbstatus will run (default 8080).                                                                                                                           | BaseConfig     |
| `HOST`                 | `localhost`                                          | Hostname or IP address of bbbstatus.                                                                                                                                       | BaseConfig     |
| `DB_CONNECTION_STRING` | `postgres://bbbstatus:bbbstatus@localhost/bbbstatus` | Database connection string (PostgreSQL).                                                                                                                                   | DatabaseConfig |
| `CSV_STRUCTURE`        | `time,user,action,text representation`               | Defines the structure of the CSV format for exports.                                                                                                                       | ReportConfig   |
| `SERVE_STATIC_CONTENT` | `true`                                               | Serving static content (`true` or `false`).                                                                                                                                | BaseConfig     |
| `BBB_SERVERS`          | *None*                                               | A list of Big Blue Button servers used for webhook validation, API functions and Report configuration                                                                      | BBBServers     |
| `TRUSTED_PROXIES`      | `127.0.0.1/8,172.17.0.1/16,::1/128`                  | A comma seperated list of CIDR notated network ranges, used to identify the *real ip* of the client, configure this value to be the list of proxies in your infrastructure | BaseConfig     | 

## Configuration behavior

### 1. Environment variables

Environment variables will always override default values and `config.toml` settings by runtime.
For example:

- `HOST` set in the environment overwrites the set hostname or ip defined in `config.toml`.

### 2. `config.toml`

If present, `config.toml` settings override the default configuration. If missing, a default `config.toml` is
automatically generated, and you can later modify it to your needs.
If Environment variables are set, they override both the default configuration **and** the config.toml configuration

### 3. Default values
If `config.toml` is missing or a parameter is not defined, the program falls back to the hardcoded default values.

### 4. BBBServers

BBB_SERVERS are used for authenticating bbb servers to bbbstatus. This has to be configured differently depending on the
configuration methode:

Through `config.toml`

If you configure bbbstatus using the config.toml, configure BBB_SERVERS like this:

```yaml
[ [ BBBServers ] ]
   HOSTNAME = "bbb.example.com"
   SHARED_SECRET = "BBB Shared Secret"
   FRIENDLY_NAME = "Example Big Blue Button Server"
```

To add another BBBServer, repeat this block for each server in your list of servers.

Through Environment variables / `.env`

If you configure bbbstatus using the .env file or using environment variables, configure BBB_SERVERS like this:
`<Hostname>[:apiPort]`

```
BBB_SERVERS="bbb.example.com:443,bbb2.example.com"
```

The comma is only required when configuring multiple bbb servers.


## Notes

- When configuring bbb servers this way, bbbstatus tries to automatically obtain the `SHARED_SECRET` as it's
  provided with each bbb-webhook request. One limitation is that this is not persistent, meaning that you have to wait
  for
  bbb-webhooks after each restart in order to use API features (like get recordings).


- You may serve static files separately in your environment, therefore there is an option to disable it.
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

[Install]
WantedBy=multi-user.target
```

Please adapt the given sample configuration to your needs and values, see [configuration](README.MD#configuration).
You will most definitely need to update the DB_CONNECTION_STRING line to the database credentials you have
setup [here](README.MD#set-up-a-database-and-a-user-with-write-permissions).

Save the changed content (on a fresh installation use `ctrl + O`)

Exit the editor (on a fresh installation use `ctrl + C`)

Now you can start, enable, stop, view logs, restart bbbstatus using the common systemd commands.
After the first start of the service, bbbstatus will write a config.toml file in it's WorkingDirectory, please use it as
your main way of configuring, using environment variables as a way of overwriting the config values.

# Deployment Docker

Make a directory to hold the docker compose file, the .env file, the Caddyfile and the data folder from bbbstatus and
Caddy (the default webserver).

- Create a file named `docker-compose.yml` using your favorite text editor, and insert the contents
  of [the example docker compose file](docker-compose.yml).
Or just simply download the given example of this repository.
- create a new file called `.env`
- Insert the contents of the [example file](dotenv) and change the configuration parameters to your needs.
  You should change at least the database username, and the database password.
- Copy the provided Caddyfile.example to Caddyfile and change at least the authentication hash (It's just a example.).
  You can generate a caddy hash using the caddy binary
  as [documented here](https://caddyserver.com/docs/command-line#caddy-hash-password)

bbbstatus requires either full ipv6 support or fully disabled ipv6 support.
If your bbb server sends webhooks using ipv6, bbbstatus **MUST** receive it as ipv6 as well.

### For the next step Choose **ONE** of the two config options

1. Fully enable IPV6 for the docker host

   On your docker host edit `/etc/docker/daemon.json`
   and paste this configuration:
    ```json
    {
        "ipv6": true,
        "fixed-cidr-v6": "fd00::/80"
    }
   ```

2. Fully disable IPV6 for the docker host

   On your docker host edit `/etc/sysctl.d/01-disable-ipv6.conf`
   and paste this configuration:
   `net.ipv6.conf.all.disable_ipv6 = 1`

## Dependencies

Frontend:

- [font-awesome 6.0.0](https://fontawesome.com/)
- [charts.css 1.1.0](https://chartscss.org/)

Backend:

- github.com/BurntSushi/toml
- github.com/jackc/pgx/v5
- github.com/labstack/echo/v4
- github.com/labstack/gommon
- github.com/nicksnyder/go-i18n/v2
- github.com/pressly/goose/v3
- golang.org/x/text