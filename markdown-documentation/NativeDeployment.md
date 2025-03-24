# Native deployment

To deploy bbbstatus natively (without docker), you can simply run the binary from any location that you want with either
a
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
