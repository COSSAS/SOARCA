# Logging

SOARCA support extensive logging. Logging is based on the [logrus](https://github.com/sirupsen/logrus) framework. 

## Format
Logging can be done in different formats suitable for your application. The following formats are available:

* JSON `default`
* Plain text


## Destination

* std::out `default` (terminal)
* To file (to log file path)

Later:

* syslog (NOT IMPLEMENTED)

## Log levels
SOARCA supports the following log levels. Also is indicated how they are used.

* PANIC (non fixable error system crash)
* FATAL (non fixable error, restart would fix)
* ERROR (operation went wrong but can be caught by other higher component)
* WARNING (let the user know some operation might not have the expected result but execution can continue on normal path)
* INFO `default` (let the user know that a major event has occurred)
* DEBUG (add some extra detail to normal execution paths)
* TRACE (get some fine grained detail from the logging)

## Types of logging
SOARCA will log different information, these will be combined in the same output. 

### Runtime logging
Runtime logging wil include the running state of SOARCA, errors encountered when registering modules etc.


### Security event logging
Will log the status of the execution of an playbook, database updates of playbooks

## Using the logger (developer)

To use SOARCA logging you can add the following to your module.

```golang

type YourModule struct {
}

var component = reflect.TypeOf(YourModule{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

```

## Changing log level

To change logging for your SOARCA instance you can use use the following environment variables 


|variable |content |description
|---|---|---|
|LOG_GLOBAL_LEVEL |[Log levels]  |One of the specified log levels. Defaults to `info`
|LOG_MODE |development \| production  |If production is chosen the `LOG_GLOBAL_LEVEL` is used for all modules defaults to `production`
|LOG_FILE_PATH |filepath  |Path to the logfile you want to use for all logging. Defaults to `""` (empty string)
|LOG_FORMAT |text \| json  |The logging can be in plain text format or in JSON format. Defaults to `json`


This can be set as environment variables or loaded through the `.env`

```bash

LOG_GLOBAL_LEVEL: "info"
LOG_MODE: "production"
LOG_FILE_PATH: ""
LOG_FORMAT: "json"
```