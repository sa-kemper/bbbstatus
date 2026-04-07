![banner.png](banner.png)

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/sa-kemper/bbbstatus?style=for-the-badge)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/sa-kemper/bbbstatus?style=for-the-badge)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/sa-kemper/bbbstatus/StaticAnalisys.yml?style=for-the-badge)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/sa-kemper/bbbstatus/UnitTests.yml?style=for-the-badge&label=unit%20test)
![GitHub License](https://img.shields.io/github/license/sa-kemper/bbbstatus?style=for-the-badge)

bbbstatus is a utility that allows you to view and save events and their details
from [bbb-webhooks](https://github.com/bigbluebutton/bbb-webhooks).
bbbstatus was developed as my first Go project. The code quality and design choices reflect this. For a more up-to-date
representation of my skills, please take a look at my GitHub.

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
| `HOST`                 | `0.0.0.0`                                            | Hostname or IP address of bbbstatus.                                                                                                                                       | BaseConfig     |
| `DB_CONNECTION_STRING` | `postgres://bbbstatus:bbbstatus@localhost/bbbstatus` | Database connection string (PostgreSQL).                                                                                                                                   | DatabaseConfig |
| `CSV_STRUCTURE`        | `time,user,action,text representation`               | Defines the structure of the CSV format for exports.                                                                                                                       | ReportConfig   |
| `SERVE_STATIC_CONTENT` | `true`                                               | Serving static content (`true` or `false`).                                                                                                                                | BaseConfig     |
| `BBB_SERVERS`          | *None*                                               | A list of Big Blue Button servers used for webhook validation, API functions and Report configuration                                                                      | BBBServers     |
| `TRUSTED_PROXIES`      | `127.0.0.1/8,172.17.0.1/16,::1/128`                  | A comma seperated list of CIDR notated network ranges, used to identify the *real ip* of the client, configure this value to be the list of proxies in your infrastructure | BaseConfig     | 
| `SERVER_LANG`          | `en`                                                 | A setting defining the language the server uses for anonymising users if GDPR is enabled                                                                                   | BaseConfig     |

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

To add another BBB server, repeat this block for each server in your list of servers.

Through Environment variables / `.env`

If you configure bbbstatus using the .env file or using environment variables, configure BBB_SERVERS like this:
`<Hostname>[:apiPort]`

```
BBB_SERVERS="bbb.example.com:443,bbb2.example.com"
```

The comma is only required when configuring multiple bbb servers.


## Notes

- When configuring BBB servers in this manner, BBBStatus attempts to automatically retrieve the 'SHARED_SECRET' provided
  with each BBB-Webhook request. One limitation is that this is not persistent, meaning that you have to wait for
  bbb-webhooks after each restart in order to use API features, such as getting recordings.

  To be absolutely clear: if you restart bbbstatus, you cannot access recordings until the corresponding bbb server has
  sent a new event.


- You may serve static files separately in your environment; therefore, there is an option to disable this feature.
  Bbbstatus will use the path /static in any case.

# Customization

You can override locales, templates, static content (such as CSS) and even database migrations without recompiling
bbbstatus. However, be aware that, although this is possible, it may cause problems as only the embedded and official
versions are tested.
You can also override multiple files at once. To do this, you need to override every file individually, as
described [here](markdown-documentation/Customisation.md).

# Deployment preparation:

Prepare your bbb server as described in [BBBServerPreperation.md](markdown-documentation/BBBServerPreperation.md)

# [Native Deployment ](markdown-documentation/NativeDeployment.md) installation guide

# [Docker Deployment](markdown-documentation/DockerDeployment.md) installation guide

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
- github.com/lithammer/fuzzysearch
