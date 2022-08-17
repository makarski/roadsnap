.PHONY: build run help app-help config env cache-all cache-one report-all report-one chart-all

config_file=${USER}-rsnap-conf.toml
config_dir=${CURDIR}/user_configs

GREEN="\033[32m"
YELLOW="\033[93m"
NOCOLOR="\033[0m"
DIM_COLOR="\033[2m"

build:
	@docker build -t roadsnap .

config:
	@./interactive_configs.sh

cache-all: config env
	$(call run_app, "cache")

cache-one: config env
	$(call run_app, "-i", "cache")

report: config env
	$(call run_app, "report")

chart-all: config env
	$(call run_app, "chart")

help: 
	@printf '${USAGE}'

app-help:
	@docker run -it roadsnap -help

env:
	$(eval include .env)

define run_app
	@docker run \
			-it \
			-v ${config_dir}:/roadsnap/user_configs \
			-v ${RS_JIRA_DIR}:/roadsnap/snapshots \
			roadsnap -config=/roadsnap/user_configs/${config_file} -dir=/roadsnap/snapshots ${1} ${2}
endef

define USAGE
 '${GREEN}'MAKE Commands'${NOCOLOR}'\n\
 =============\n\
\n\
* '${YELLOW}'build'${NOCOLOR}'      : build your docker image\n\
* '${YELLOW}'config'${NOCOLOR}'     : configure your application\n\
* '${YELLOW}'cache-all'${NOCOLOR}'  : caches JIRA epics for all configured projects\n\
* '${YELLOW}'cache-one'${NOCOLOR}'  : interactive mode - user is asked what project to cache\n\
* '${YELLOW}'report'${NOCOLOR}'     : (re)generates markdown snapshot report for all available cached projects (by month)\n\
* '${YELLOW}'chart-all'${NOCOLOR}'  : generates stacked column charts for all projects, all dates - allows to analyze trends\n\

endef
