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
docker run bloxroute/bdn-operations-relay:v0.0.1 -p 9080:9080 --config=config.yml
```

### Test config

You can generate test values for the `dapp-private-key`, `dapp-address`, and `solver-private-key` using the following code:

```go
package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

func generateConfig() {
	dappPrivateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	dAppAddress := crypto.PubkeyToAddress(dappPrivateKey.PublicKey).String()
	dappPrivateKeyHex := hex.EncodeToString(crypto.FromECDSA(dappPrivateKey))

	solverProvateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	solverPrivateKeyHex := hex.EncodeToString(crypto.FromECDSA(solverProvateKey))

	fmt.Println("dapp-private-key", dappPrivateKeyHex)
	fmt.Println("dapp-address", dAppAddress)
	fmt.Println("solver-private-key", solverPrivateKeyHex)
}
```