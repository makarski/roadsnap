#!/bin/bash

GREEN="\e[32m"
YELLOW="\e[93m"
NOCOLOR="\e[0m"
DIM_COLOR="\e[2m"

source .env 2>/dev/null
if [[ -z "${RS_JIRA_DIR}" ]]
then
    read -p  "$(printf "> ${YELLOW}Enter path to your roadsnaps dir: "${NOCOLOR})" roadsnaps_dir
    echo "export RS_JIRA_DIR=\"${roadsnaps_dir}\"" > .env
    source .env
else
    echo -e $(printf "${GREEN}Your snapshot dir is set to ${YELLOW}${RS_JIRA_DIR}${NOCOLOR}")
fi

config_name=${USER}-rsnap-conf.toml
config_key=./user_configs/${config_name}

if [ -f ${config_key} ]
then
    echo -e $(printf "${GREEN}Configuration file ${YELLOW}${config_key}${GREEN} already exists.
    To rerun the config remove the file${NOCOLOR}")
    exit 0;
fi

echo -e $(printf "${GREEN}Configuring Jira Access")
echo $'\n'

read -p  "$(printf "> ${YELLOW}Enter jira email":${NOCOLOR}) " jira_email
read -p  "$(printf "> ${YELLOW}Enter jira account_id":${NOCOLOR}) " jira_account_id
read -p  "$(printf "> ${YELLOW}Enter jira base_url":${NOCOLOR}) " jira_base_url
read -p  "$(printf "> ${YELLOW}Enter jira token":${NOCOLOR}) " jira_token

echo $'\n'
echo -e $(printf "${GREEN}Enter Jira Project Names ${DIM_COLOR}(Ctrl-D to stop)${NOCOLOR}")
read -r -d $'\04' jira_projects

RS_PROJECT_NAMES=""

for p in ${jira_projects}
do
    RS_PROJECT_NAMES+=\"${p}\",
done

mkdir user_configs

RS_JIRA_EMAIL=\"${jira_email}\" \
RS_JIRA_ACCOUNT_ID=\"${jira_account_id}\" \
RS_JIRA_BASE_URL=\"${jira_base_url}\" \
RS_JIRA_TOKEN=\"${jira_token}\" \
RS_PROJECT_NAMES=${RS_PROJECT_NAMES} \
envsubst < rsnap-config.toml.template > ${config_key}

echo -e "$(printf "> ${GREEN}Successfully generated config file ${config_key}":${NOCOLOR})"
