# Deployer

[![DeepSource](https://deepsource.io/gh/Junte/deployer.svg/?label=active+issues&show_trend=true&token=jzluf0dMTSlbqqtmaIh5VMV4)](https://deepsource.io/gh/Junte/deployer/?ref=repository-badge)

## Overview

One of the most important aspects of DevOps is to deploy a new application version from ci/cd tool (
GitLab, GitHub
Actions,...) to target linux servers.
As a solution, you can put a private key to the ci/cd tool, then log to the target server by ssh and
run the deploy
script.
BUT! It is very dangerous: your private key can be stolen by cybercriminals on the ci/cd provider
side.
This tool suggests another approach: the deployer background task is running on the target server
and listening to the
public port.
From the deploy stage of the ci/cd pipeline, you can post an HTTP request in the specified format to
this port and the
tool will execute the deploy command.
No private information should be stored on the ci/cd provider (only deployer address).
Even if somebody will have access to your settings on ci/cd provider he can't run custom script on
the target server
because you control allowed commands

## Installation

- download binary from [releases page](https://github.com/Junte/deployer/releases)
- create configuration file `config.yaml` and place it at folder with executable
- run tool `./deployer`

As a suggestion, use the systemd as a manager for the tool. Systemd service file example:

```
[Unit]
Description=Deployer service.

[Service]
Type=simple
Restart=always
RestartSec=10
WorkingDirectory=/opt/services/deployer
User=deploy
ExecStart=/opt/services/deployer/deployer

[Install]
WantedBy=multi-user.target
```

## Configuration

```yaml
port: ":7777"  # required - listening port 
environment: "dev" # required - environment label
tls: # optional - if provided https server is started
  cert: ./tls/cert.crt
  key: ./tls/cert.key
notification: # optional - notifications config
  slack:
    apiToken: "<...>" # slack api token
    channel: "#deployments" # slack channel to send message
components: # list of components for deploy
  backend: # component name. Value of "component" query parameter
    workdir: /opt/services/app/ # if set - use the workdir
    command: [ "./deploy_backend.sh", "--tag={{ .Args.tag }}" ] # required - deploy command
    key: "<...>" # optional - random key for additional protection. If not provided - don't check. Value of "key" query parameter 
  frontend:
    command: [ "/opt/services/app/deploy_frontend.sh", "--tag={{ .Args.tag }}" ] # required - deploy command
```

### Command format

#### Arguments

The command can be any shell script with/without parameters.
In a request to deployer, some additional query parameters can be provided.
They can be injected in command in format `{{ .Args.<query_parameter> }}`.
For example, the `tag` query parameter can be used in command by adding `{{ .Args.tag }}` to the
desired place.

#### Async

If you don't need wait for deploy results, you can add `async` query param. For example:

```
...&text=bla&async...
or 
...&text=bla&async=true...
```

## Examples

### Curl

```shell script
curl -X POST -d "component=backend&key=secret&tag=42" https://deployer.example.com
```

### GitLab CI

```yaml
deploy:
  image: curlimages/curl:7.74.0
  stage: deploy
  dependencies: [ ]
  script:
    - curl -X POST -d "component=backend&key=${DEPLOYER_KEY}&tag=${CI_PIPELINE_ID}" ${DEPLOYER_HOST}
```

* DEPLOYER_KEY, DEPLOYER_HOST are taken from ci/cd variables setting

### GitHub Actions

```yaml
deploy:
  runs-on: ubuntu-20.04
  if: ${{ github.ref == 'refs/heads/main' }}
  steps:
    - name: Deploy
      run: curl -k -X POST -d "component=backend&key=${{ secrets.DEPLOYER_KEY }}&tag=${{ tag }}" ${{ secrets.DEPLOYER_HOST }}
```

* DEPLOYER_KEY, DEPLOYER_HOST are taken from github secrets 
