# Deployer

## Overview
One of the most important aspects of devops is deploy new version to servers. As as a solution you can put private key to ci/cd tool (GitLab, GitHub Actions,...) and login by ssh to the target server and run deploy script. BUT! It is very dangerous: your private key can stolen by cyber criminals. This tool suggests anothet approach: some tool can be runned on target servers and listen public port. From deploy stage of ci/cd pipeline you can post http request in the specified format to this port and the tool will execute deploy script. No any private information should be stored on ci/cd provider. Event if somebody will have access to your settings he can't run any command on the target server - you limit allowed commands

## Installation

## Configuration

## Examples
### GitLab CI
