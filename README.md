Roadsnap
========

This simple CLI tool allows to cache **JIRA epics** by a project name and then generate an **overview markdown** file that will contain the epic links split by the categories:

- **Done** *(complete by status & all tasks done)*
- **Ongoing** *(status in progress)*
- **Overdue** *(due date in the past & status not done or not all stories done)*
- **Outstanding** *(status todo)*

### Usage

Prerequisite: [Golang](https://go.dev/dl/)

```sh
$ cp rsnap-config.toml.dist rsnap-config.toml

$ go build

# Help
$ roadsnap --help

# Generate column stacked chart
$ roadsnap -dir=path/to/your/snapshots/dir chart

# Cache
$ roadsnap -dir=path/to/your/snapshots/dir cache

# Cache in interactive mode
$ roadsnap -dir=path/to/your/snapshots/dir -i cache

# Generate markdown
$ roadsnap -dir=path/to/your/snapshots/dir list

# Generate in interactive mode
$ roadsnap -dir=path/to/your/snapshots/dir -i list
```
