Roadsnap
========

This simple CLI tool allows to cache **JIRA epics** by a project name and then generate an **overview markdown** file that will contain the epic links split by the categories:

- **Done** *(complete by status & all tasks done)*
- **Ongoing** *(status in progress)*
- **Overdue** *(due date in the past & status not done or not all stories done)*
- **Outstanding** *(status todo)*

### Usage

Prerequisites
* [Docker](https://docker.com)
* Make

```sh
# Build your application
$ make build

# Configure your application
$ make config

# Check the reference
$ make help
```
