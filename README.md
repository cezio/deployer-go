# Deployer

## Purpose

`Deployer` is a simple http server that receives a request (webhook preassumbly) and executes a predefined command as a subprocess. This can be used to deploy a project from a repository changeset, but you don't want to run a complete CI/CD pipeline (I'm thinking of you, Jenkins), and the simplest solution should be enough.

## Workflow

`Deployer` runs as a single binary, which starts a http server. Runtime configuration should point to a directory with configuration files to run. Server awaits for requests on `/incoming/$name` path. `$name` portion is used as a name of configuration file to run. Server checks if it can read the config, and validates if a provided request is correct. Then it executes the command from the configuration. Any error during that process causes non-200 HTTP response. Succesful execution results in HTTP 200, `OK` response. Response body will contain output from executed command. Internally, deployer will perform simple file locking to ensure just one request is executed at once. Subsequent, locked requests will be executed in nondetermined order.

## Configuration file

Configuration files should be stored in the single directory. `$name` from URL is translated to `$name.conf` config file name. Each configuration should be a TOML file:

```
## subprocess command and args, type: list of strings
commands = ["/bin/bash",
            "/home/cezio/Projects/deployer-go/script.sh"]

## env for subprocess, type: list of strings
env = ["DJANGO_SETTINGS_MODULE=myapp.local_settings"]
## a string with expected body contents, comment to leave it empty
# secret =
## optional name of the header where the secret will be delivered
# secret-header = 
## list of allowed methods, comment to leave it empty, type: list of strings
allowed-methods= ['POST', 'GET']
```

Following keys are required:

* `commands` - a list of arguments for subcommands (mind the list of strings type)
* `env` - a list of enviroment variables to pass to subcommand (mind the list of strings type)

Optional keys:

* `secret` - a string with a secret to be passed in body (leave commented to ignore secret)
* `allowed-method` - a list of http methods which are allowed for this config.

## Running

`Deployer` can be run in the following way:

`./deployer-go [-port 8081] [-configdir .]`

This will execute one server process, which will be attached to the current tty. The process will bind to port 8081 if no argument will be provided with the `-port` parameter.

## License

This software is licensed under MIT license. See `LICENSE` file for details.