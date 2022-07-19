Roadsnap
========

This simple CLI tool allows to cache **JIRA epics** by a project name and then generate an **overview markdown** file that will contain the epic links split by the categories:

- **Done** *(complete by due date & all tasks done)*
- **Ongoing** *(due date in the future)*
- **Overdue** *(due date in the past & some tasks are still open)*
- **Outstanding** *(start date in the future)*

### Usage

Prerequisite: [Golang](https://go.dev/dl/)

```sh
$ cp rsnap-config.toml.dist rsnap-config.toml

$ go build

# Help
$ roadsnap --help

# Cache
$ roadsnap cache

# Generate markdown
$ roadsnap list
```
