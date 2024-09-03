# bdn-operations-relay
BDN implementation of Atlas Operations Relay

## Overview

BDN operations relay is a service that binds Atlas and BDN. 

<img src="static/diagram.svg" width="1024">

## Running the service

To run the service, you need to set values into the `config.yaml` file. 

### Docker

To run the service using docker, you can use the following command:

```bash
docker build -t bloxroute/bdn-operations-relay:v0.0.1 .
docker run bloxroute/bdn-operations-relay:v0.0.1 --config=config.yml
```
