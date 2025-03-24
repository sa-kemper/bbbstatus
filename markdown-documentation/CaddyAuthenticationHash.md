# To make a Caddy authentication hash (in a docker deployment)

use the docker exec command to enter the caddy container:

```shell
docker exec -it caddy sh
```

From this point you have access to the caddy server binary and you can use it to generate a
hash [like documented here](https://caddyserver.com/docs/command-line#caddy-hash-password)

Example:

```shell
caddy hash-password -p PasswordHere
```

This will return a hash that you can use in your caddy file.