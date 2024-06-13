# Stack Definitions

### Complete example

```toml

stack = "stack-name" # must be globally unique

root_app = "app-name" # app that handles HTTP requests to stack-name.ark-root-domain.tld

[apps.app-name]
    type = "web" # web, pserv, worker, or cron
    repo_url = "github.com/..."

    cpu = 1
    mem = 256

    schedule = "* * * *" # cron schedule if this app is of type cron. only valid if type = "cron"

    [apps.app-name.build]
        dockerfile = "Dockerfile" # path to dockerfile (within repo)
        ignorefile = "/path/.dockerignore" # should look for .dockerignore by default
        build-target = "app" # for multi-stage dockerfiles

        [apps.app-name.build.args]
            ENV = "preview"

    [apps.app-name.deploy]
        command = "bin/rails server"
        release_command = "bin/rails db:prepare"

    [apps.app-name.env]
        LOG_LEVEL = "debug"

    [apps.app-name.http_service] # only valid if type = "web" or type = "pserv"
        container_port = 8080
        keep_alive = false

    [apps.app-name.health_check]
        grace_period = "10s"
        intervale = "30s"
        timeout = "5s"
        command = "bin/health_check" # run on docker container via docker exec
        request = "http GET /" # defaults to an http GET request

    [apps.app-name.disks.log-data]
        mount_path = "/var/app-name/logs"
        size = 5 # in GB

[services.service-name]
    image = "postgis/postgis"
    repo_url = "github.com/..."
    dockerfile = "Dockerfile"

```
