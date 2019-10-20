# Deployer

## Purpose

`Deployer` is a simple server that receives requests and executes predefined commands on the server. This can be used for headless deployments, where you don't want to run whole CI/CD pipeline (I'm thinking of you, Jenkins), and simplest solution should be enough.

## Workflow

`Deployer` runs as a single binary, which starts web server. Runtime configuration should point to a directory with configuration files to run. Server awaits for requests on `/incoming/$name` path. `$name` portion is used as a name of configuration to run. Server checks if it can read given config, and if provided request is correct. Then it executes command from configuration. Any error during that process causes non-200 HTTP response. Succesful execution results in HTTP 200, `OK` response.

## Configuration file

Configuration files should be stored in one directory. `$name` from URL is translated to `$name.conf` config file name. Each configuration should be a TOML file:

```
## list of subprocess command and args
commands = ["/bin/bash",
            "/home/cezio/Projects/deployer-go/script.sh"]

## env for subprocess
env = ["DJANGO_SETTINGS_MODULE=myapp.local_settings"]
## a string with expected body contents, comment to leave it empty
# secret =
## list of allowed methods, comment to leave it empty
allowed-methods= ['POST', 'GET']
```

Following keys are required:

* `commands` - a list of arguments for subcommands (mind the type)
* `env` - a list of enviroment variables to pass to subcommand

Optional keys:
* `secret` - string with secret to be passed in body (leave commented to ignore secret)
* `allowed-method` - list of http methods which are allowed for this config.

## Running

`Deployer` can be run in following way:

`./deployer-go [-port 8081] [-configdir .]`

## License

This software is licensed under MIT license. See `LICENSE` file for details.