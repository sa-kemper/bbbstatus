# Compiling the docker image.

This is recommended for most installations, as even though it breaks reproducibility it ensures compatibility with your
current CPU and it's capabilities. Another use-case is in scenarios where your CPU is not supported, As well as in
scenarios where you want to customize bbbstatus.

### 1. Clone your current code repository

```shell
git clone https://git.howaboutno.org/badmin/bbbstatus
```

### 2. Enter the cloned repo

```shell
cd bbbstatus
```

### 3. Make your changes

Change bbbstatus as much as you require, you can use `go build src` & `go run src`, for more details look at
the [Development Guide](DevelopmentGuide.md)

### 4. Compile the docker image

```shell
docker build -t bbbstatus:latest src
```

### 5. Deploy the docker image using the existing [Docker Deployment Guide](DockerDeployment.md)