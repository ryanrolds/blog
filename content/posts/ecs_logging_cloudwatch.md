# Logging, it's mission critical

When building any kind of service in any language, one of the first things that gets setup during start-up is logging. In this post we will talk about the essentials, centralized logging, and how this blogs implements logging.

## Essentials

Logs are a plain text journals that software keeps about itself. Software developers are often watching [STDOUT & STDERR](https://en.wikipedia.org/wiki/Standard_streams) as they test and debug their code locally. Deployed applications are often write their logs to files on the server or to a centralized logging service (more on this later).

The contents of these journals varies depending on the purpose of the journal. It's not uncommon for services to keep multiple logs. Start-up details (port numbers, DB connection information, warnings), errors, stack traces, scheduled tasks, etc... go in to service logs. Access/request logs can go in the service log, but it's also common to write them to a seperate access log. Databases have their own service logs as well as query logs, which may include all queries (not performant) or just slow quieries. Many applications also keep a seperate log for just errors.

What is written in to the log is often configurable, which will vary depeneding on the environment (dev, testing, staging, production, etc...). A person working in thier devlopment environment wants to see detailed information relevent to their current task. However, someone investigating why a production service isn't working only wants to look for errors. Logging libraries normally provide a way of tagging a log entry with a severity level. When configuring the services/libraries a miniumum level can be defined and only log entries of that level or higher will be written to the logs. 

The logs written by this blog look like:

```
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"Loading file ./content/posts/2019_happy_new_year.md","time":"2019-02-05T20:03:29-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"Loading file ./content/posts/basic_rss_feed.md","time":"2019-02-05T20:03:29-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"warning","msg":"Image not found for post","time":"2019-02-05T20:03:29-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"Loading file ./content/posts/efficient_http_caching.md","time":"2019-02-05T20:03:29-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"warning","msg":"Image not found for post","time":"2019-02-05T20:03:29-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"Loading file ./content/posts/this_blog_part_1.md","time":"2019-02-05T20:03:29-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"warning","msg":"Image not found for post","time":"2019-02-05T20:03:29-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"Loading file ./content/404.md","time":"2019-02-05T20:03:29-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"Loading file ./content/500.md","time":"2019-02-05T20:03:29-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"Starting server","time":"2019-02-05T20:03:29-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"::1 - - [05/Feb/2019:20:03:32 -0800] \"GET / HTTP/1.1\" 200 5682","time":"2019-02-05T20:03:32-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"::1 - - [05/Feb/2019:20:03:32 -0800] \"GET /static/style.css?m=c9c82ed84e35f71f9533b81494a6f2a6 HTTP/1.1\" 200 3931","time":"2019-02-05T20:03:32-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"::1 - - [05/Feb/2019:20:03:32 -0800] \"GET /static/logo.png?m=30fac5d7c5602071f356c220903432f4 HTTP/1.1\" 200 12525","time":"2019-02-05T20:03:32-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"::1 - - [05/Feb/2019:20:03:32 -0800] \"GET /static/ryanolds.jpg?m=213e216ee736fcf35e6216e901f2947f HTTP/1.1\" 200 158955","time":"2019-02-05T20:03:32-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"::1 - - [05/Feb/2019:20:03:32 -0800] \"GET /static/prog_intro_to_math.jpg?m=13a6992b40e5a484ea8ba7380ccbdb97 HTTP/1.1\" 200 83886","time":"2019-02-05T20:03:32-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"::1 - - [05/Feb/2019:20:03:32 -0800] \"GET /static/thinking_fast_slow.jpg?m=e87375870d0d98768e1e05b43de5b5ae HTTP/1.1\" 200 35803","time":"2019-02-05T20:03:32-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"::1 - - [05/Feb/2019:20:03:33 -0800] \"GET /favicon.ico HTTP/1.1\" 200 4334","time":"2019-02-05T20:03:33-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"::1 - - [05/Feb/2019:20:03:38 -0800] \"GET /posts/this_blog_part_1 HTTP/1.1\" 200 29478","time":"2019-02-05T20:03:38-08:00"}
```

Some services use a more human readable format, which may be structured like the JSON structured logs above. Each log entry contains the environment, hostname, level, message, and time. Each the key-value pairs in the log entry are useful when debugging, they provide context about the message. The last 8 lines are example of HTTP access logs from someone (me) loading the homepage and one blog post. The log message contains a structured log entry in Apache's access log format. The Apache format contains important details about HTTP requests. A lot of services don't mix service logs and access log. I'm not a fan of lumping them together, but little bit of work to split them up is low priority.

Above we can see two levels, info and warning. Most logging libraries also also `fatal`, `error`, and `debug`, but there are less common ones like `verbose`, `notice`, `critical`, and more. When writting something to the logs think about the use case, when would you *want* to read this? If it's only during development then make it `debug` level. If it's useful information but shouldn't raise concerns give it a level of `info`. If you're writing out the cause of a crash make it a `fatal` or `error` log. `Warn` is useful for things that should get attention, but are not causing significant issues.

Looking at the logs requires SSHing into the server and either looking at output from SystemD or the contents of files. If the service developers are nice the logs will be located in `/var/log`. A major pain point is that if you're trying to track down a problem across multiple hosts/servers it requires stablishing multiple SSH connections and rooting around each of the system. This can be very painful in large system. This is solved by centralized logging platforms.

## Centralized Logging

Centralized logging provides a place that all logs (system, service, access, etc...) can sent to for aggregration, indexing, monitoring, and storage. The logging system will provide a UI and APIs to search and visualize the logs. Many systems also support integrations with monitoring/alerting systems. If a service writes a `fatal, or `error`, level event to the logging system 


## Implementation


## Wrap-up

