<<<<<<< Updated upstream
# Service Logs, ECS, and Cloudwatch
=======
# Logging, it's mission critical

When building any kind of service in any language, one of the first things that gets setup during start-up is logging. In this post we will talk about the essentials, centralized logging, and how this blogs implements logging.

## Essentials

Logs are a plain text journals that software keeps about itself. Software developers are often watching [STDOUT & STDERR](https://en.wikipedia.org/wiki/Standard_streams) as they test and debug their code locally. Deployed applications are often write their logs to files on the server or to a centralized logging service (more on this later).

The contents of these journals varies depending on the purpose of the journal. It's not uncommon for services to keep multiple logs. Start-up details (port numbers, DB connection information, warnings), errors, stack traces, scheduled tasks, etc... go in to service logs. Access/request logs can go in the service log, but it's also common to write them to a seperate access log. Databases have their own service logs as well as query logs, which may include all queries (not performant) or just slow quieries. Many applications also keep a seperate log for just errors.

What is written in to the log is often configurable, which will vary depeneding on the environment (dev, testing, staging, production, etc...). A person working in thier devlopment environment wants to see detailed information relevent to their current task. However, someone investigating why a production service isn't working only wants to look for errors. Logging libraries normally provide a way of tagging a log entry with a severity level. When configuring the services/libraries a miniumum level can be defined and only log entries of that level or higher will be written to the logs. 

The logs written by this blog look like:

```
{"env":"local","host":"DESKTOP-GKDCTUP","level":"info","msg":"Loading file ./content/posts/2019_happy_new_year.md","time":"2019-02-05T19:58:17-08:00"}
{"env":"local","host":"DESKTOP-GKDCTUP","level":"fatal","msg":"template: no template \"amp.tmpl\" associated with template \"\"","time":"2019-02-05T19:58:17-08:00"}
```





## Centralized Logging


## Implementation


## Wrap-up

>>>>>>> Stashed changes


