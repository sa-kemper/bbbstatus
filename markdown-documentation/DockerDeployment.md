# Docker Deployment

## See also: [Docker Compile  Guide](DockerCompile.md)

Make a directory to hold the docker compose file, the .env file, the Caddyfile and the data folder from bbbstatus and
Caddy (the default webserver).


---

# Deployment Instructions
- Create a file named `docker-compose.yml` using your favorite text editor, and insert the contents
  of [the example docker compose file](docker-compose.yml).
  Or just simply download the given example of this repository.
- create a new file called `.env`
- Insert the contents of the [example file](dotenv) and change the configuration parameters to your needs.
  You should change at least the database password.
- Copy the provided Caddyfile.example to Caddyfile and change at least the authentication hash (It's just a example.).
  You can generate a caddy hash using the caddy binary
  as [documented here](https://caddyserver.com/docs/command-line#caddy-hash-password)

## Notes:

- For some installations (ARM), formatting issues and verbose configuration of the Caddyfile result in strange bugs such
  as ip resolution being broken. (Fix: use
  the [caddy container in interactive mode](https://stackoverflow.com/a/71943760)
  to [format the Caddyfile](https://caddyserver.com/docs/command-line#caddy-fmt) with the --override option)

## IPV4 & IPV6

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
