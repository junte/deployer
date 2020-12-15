# Deployer

## Overview
One of the most important aspects of DevOps is to deploy a new application version to target servers.
As a solution, you can put a private key to the ci/cd tool (GitLab, GitHub Actions,...) and log to the target server by ssh and run the deploy script. 
BUT! It is very dangerous: your private key can be stolen by cybercriminals on the ci/cd provider side. 
This tool suggests another approach: the deployer background task is running on the target server and listening to the public port. 
From the deploy stage of the ci/cd pipeline, you can post an HTTP request in the specified format to this port and the tool will execute the deploy command. 
No private information should be stored on the ci/cd provider (only deployer address).
Even if somebody will have access to your settings on ci/cd provider he can't run custom script on the target server - you control allowed commands

## Installation

## Configuration
```yaml
port: ":7777"  # required - listening port 
tls:  # optional - if provided requests handled by https 
  cert: ./tls/cert.crt
  key: ./tls/cert.key
components: # list of components for deploy
  backend:
    command: ["/opt/services/app/deploy_backend.sh", "--tag=${arg_tag}"] # required - deploy command
    key: "<...>" # optional - random key for small protection. Should be passed in request
  frontend:
    command: "/opt/services/app/deploy_frontend.sh", "--tag=${arg_tag}"] # required - deploy command
    key: "<...>" # optional - random key for small protection. Should be passed in request
```

## Examples
### GitLab CI
For example, we for us gitlab "pipeline id" is our application version: 
```yaml
deploy:
  image: curlimages/curl:7.74.0
  stage: deploy
  dependencies: []
  script:
    - curl -X POST -d "component=${DEPLOYER_COMPONENT}&key=${DEPLOYER_KEY}&tag=${CI_PIPELINE_ID}" ${DEPLOYER_HOST}
```