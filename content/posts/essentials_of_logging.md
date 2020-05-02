---
title: Visibility is Mission Critical 
published: 2019-02-12T00:26:27Z
intro: Quality logs make issue investigation, exploration, and iterative system improvements much easier. Without logs, it's pretty much impossible to reason about what your service is doing or has done.
---
When building services, one of the best developer ergonomics improvements is logging. The value of well-structured logs cannot be understated. Quality logs make issue investigation, exploration, and iterative system improvements much easier. Without logs, it's pretty much impossible to reason about what your service is doing or has done.

## Essentials

Logs are journals that services keep about what it's doing. On the first cut, a developer usually writes unstructured text to [stdout & stderr](https://en.wikipedia.org/wiki/Standard_streams) or a file. After a while, a logging library is introduced to the project and slowly the writes to stdout/stderr are replaced with calls to the logging library. Save yourself the time and use a logging library from the start. 

The contents of the log vary depending on the purpose of the log. It's not uncommon for services to keep multiple logs, writing them out to different files or a centralized logging system. Start-up details (ports, DB connection status, values of crucial environment variables), errors, stack traces, scheduled tasks, etc... go in the service log. Requests can go in the service log, but most services write them to a separate file. Writing slow performance information (slow operations, queries, requests) is also common. 

What entries are written into the log is often configurable. The configuration varies depending on the environment (dev, testing, staging, production, etc...). A person working in their development environment wants to see detailed information relevant to their current task. However, someone investigating a problem in production is looking mostly for errors. To support the various use cases logging libraries provide a way of specifying a severity level. A minimum severity level can be set, and only entrie of that level or higher are written to the logs. 

Another key feature of logging libraries is tagging additional context information to the log entry. It's common to include time, hostname, PID (process id), environment, and more. Different kinds of logs will include different contextual details. Fi

Logging libraries also support writing entries in one or more formats. When writing to stdout, the format should be human-readable (possibly with color). However, when writing to a central logging system the format should be structured for indexing of key details/tags. 

Below are logs from a project written in Go using [Logrus](https://github.com/sirupsen/logrus). Note the level, seconds from start, message, and tag (port). These logs are written to stdout and are intended to be human-readable.

```
INFO[0000] Starting...
WARN[0000] PG_URL not provided, using default
INFO[0000] Listening... port=8080
INFO[0001] Metric request
```

Below are logs from this blog. It's the same library as above (Logrus), but it's been configured to write structured logs in JSON with more details. Some of the logs entries are from requests and contain an Apache formatted HTTP request log entry.

``` json
{"env":"local","host":"desktop-gkdctup","level":"info","msg":"starting server","time":"2019-02-05t20:03:29-08:00"}
{"env":"local","host":"desktop-gkdctup","level":"info","msg":"::1 - - [05/feb/2019:20:03:32 -0800] \"get / http/1.1\" 200 5682","time":"2019-02-05t20:03:32-08:00"}
{"env":"local","host":"desktop-gkdctup","level":"info","msg":"::1 - - [05/feb/2019:20:03:32 -0800] \"get /static/style.css?m=c9c82ed84e35f71f9533b81494a6f2a6 http/1.1\" 200 3931","time":"2019-02-05t20:03:32-08:00"}
{"env":"local","host":"desktop-gkdctup","level":"info","msg":"::1 - - [05/feb/2019:20:03:32 -0800] \"get /static/logo.png?m=30fac5d7c5602071f356c220903432f4 http/1.1\" 200 12525","time":"2019-02-05t20:03:32-08:00"}
{"env":"local","host":"desktop-gkdctup","level":"info","msg":"::1 - - [05/feb/2019:20:03:32 -0800] \"get /static/ryanolds.jpg?m=213e216ee736fcf35e6216e901f2947f http/1.1\" 200 158955","time":"2019-02-05t20:03:32-08:00"}
```

## Implementations

Let's look over logging implementations for Go ([Logrus](https://github.com/sirupsen/logrus)), Node.js ([Winston](https://github.com/winstonjs/winston)), and Python ([Logmatic](https://github.com/logmatic/logmatic-python)). Each example will write the message, environment, hostname, and time to stdout. They all look pretty similar, get the logger, set the logging level, set a new formatter that outputs JSON, add some metadata (hostname and environment name), and then write an `info` level log.

### Go w/ Logrus

``` go
package main

import (
  "os"
  "github.com/sirupsen/logrus"
)

func main() {
  log := logrus.New()
  log.SetLevel(logrus.InfoLevel)
  log.SetFormatter(&logrus.JSONFormatter{})
  log.SetOutput(os.Stdout)
  log = logrus.NewEntry(log)

  env := os.Getenv("ENV")
  if env == "" {
    env = "local"
  }

  hostname, err := os.Hostname()
  if err != nil {
    log.WithError(err).Error("Problem getting hostname")
    hostname = "unknown"
  }

  log = log.WithFields(logrus.Fields{
    "env":  env,
    "host": hostname,
  })

  log.WithField("foo", "bar").Info("Starting....")
  // {"env":"local","host":"desktop-gkdctup","level":"info","msg":"starting server","foo":"bar","time":"2019-02-05t20:03:29-08:00"}

  ...
}
```

### Node.js w/ Winston

``` javascript 
os = require("os");
winston = require('winston');

env = process.env.ENV;
if (!env) {
  env = "local";
}

hostname = os.hostname();

log = winston.createLogger({
  "level": "info",
  "defaultMeta": {
    "env": env,
    "host": hostname
  },
  "format": winston.format.json(),
  "transports": [
     new winston.transports.Console()
  ]
});

log.info("Starting...", {"foo": "bar"});
// {"message":"Starting...","level":"info","env":"local","host":"DESKTOP-GKDCTUP", "foo": "bar", "timestamp":"2019-02-10T23:17:02.615Z"}

...
```


### Python w/ Logmatic

``` python
import os
import platform
import logging
import logmatic

env = os.getenv("ENV", "local")
host = platform.node()

log = logging.getLogger(__name__)
log.setLevel(logging.INFO)

handler = logging.StreamHandler()
formatter = logmatic.JsonFormatter(extra={
  "host": host,
  "env": env
})
handler.setFormatter(formatter)
log.addHandler(handler)

log.info("Starting...", extra={"foo": "bar"})
# {"asctime": "2019-02-10T15:54:14Z-0800", "name": "__main__", "processName": "MainProcess", "filename": "log.py", "funcName": "<module>", "levelname": "INFO", "lineno": 20, "module": "log", "threadName": "MainThread", "message": "Starting...", "foo": "bar", "timestamp": "2019-02-10T15:54:14Z-0800", "host": "DESKTOP-GKDCTUP", "env": "local"}

...
```

## Wrap-up

The quality of the data in the logs has a direct impact on issue investigation and development speed. Putting a little thought into how you're logging, the metadata you're including, and where they are stored will payoff manyfold. As we saw in the last section, they are not difficult to set up or use. When combined with centralized logging your team will be able to monitor, debug, and explore behavior across your entire system from a single place; Reducing downtime, making customers happier, and improving developers/operations quality of life.

