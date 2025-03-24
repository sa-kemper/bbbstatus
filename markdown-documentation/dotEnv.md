# The dotenv or .env file

This file is used for docker deployments only and is ignored by native deployments.
The docker daemon uses the `.env` file of the current work directory to fill out the `docker-compose.yml` template.
This is different from env_file as this is only used for the `docker-compose.yml` and not inside the containers. You may
add those options to the docker-compose.yml or bind the config.toml into the bbbstatus container.