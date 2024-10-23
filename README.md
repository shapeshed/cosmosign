# Cosmosign

[![Go Reference](https://pkg.go.dev/badge/github.com/shapeshed/cosmosign.svg)](https://pkg.go.dev/github.com/shapeshed/cosmosign)
[![golangci-lint](https://github.com/shapeshed/cosmosign/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/shapeshed/cosmosign/actions/workflows/golangci-lint.yml)
[![test](https://github.com/shapeshed/cosmosign/actions/workflows/go.yml/badge.svg)](https://github.com/shapeshed/cosmosign/actions/workflows/go.yml)

**Cosmosign** is a lightweight Go library for signing and broadcasting
transactions to Cosmos blockchains.

## Installation

```sh
go get github.com/shapeshed/cosmosign
```

## Quickstart

```go
package main

import (
	"log"

	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	cosmosign "github.com/shapeshed/cosmosign"
)

func main() {
	cs, err := cosmosign.NewClient(
		cosmosign.WithGRPCURL("http://localhost:19190"),
		cosmosign.WithRPCURL("http://localhost:26657"),
		cosmosign.WithGasPrices("0.0ustake"),
		cosmosign.WithKeyringBackend("pass"),
		cosmosign.WithKeyringRootDir("/home/cosmos/"),
		cosmosign.WithKeyringUID("account1"),
	)
	if err != nil {
		log.Fatalf("Failed to initialise cosmosign: %v", err)
	}

	fromAddress := "cosmos1..."
	toAddress := "cosmos1..."
	amount, err := types.ParseCoinsNormalized("1000ustake")
	if err != nil {
		log.Fatalf("Failed to parse amount: %v", err)
	}

	msg := banktypes.NewMsgSend(fromAddress, toAddress, amount)

	res, err := cs.SendMessages(msg)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	if res.TxResponse.Code == 0 {
		log.Printf("Transaction successful, hash: %s", res.TxResponse.TxHash)
	} else {
		log.Printf("Transaction failed, code: %d, log: %s", res.TxResponse.Code, res.TxResponse.RawLog)
	}
}
```

## Options

You may pass an arbitrary number of options when creating the `cosmosign`
client. Each option has a default value, but you may override them using the
available `With` methods.

| Option             | Description                         | Default Value              | Method to Override             |
| ------------------ | ----------------------------------- | -------------------------- | ------------------------------ |
| `AddressPrefix`    | Bech32 prefix for account addresses | `"cosmos"`                 | `WithAddressPrefix(string)`    |
| `Fees`             | Transaction fees                    | `""`                       | `WithFees(string)`             |
| `FeeGranter`       | Address of the fee granter          | `""`                       | `WithFeeGranter(string)`       |
| `FeePayer`         | Address of the fee payer            | `""`                       | `WithFeePayer(string)`         |
| `Gas`              | Gas limit                           | `""`                       | `WithGas(uint64)`              |
| `GasPrices`        | Gas prices to pay per unit of gas   | `"0.0ustake"`              | `WithGasPrices(string)`        |
| `GasMultiplier`    | Multipler for gas estimation        | `"1.0"`                    | `WithGasMultiplier(float64)`   |
| `GRPCURL`          | gRPC endpoint for Cosmos nodes      | `"localhost:1919"`         | `WithGRPCURL(string)`          |
| `GRPCTLS`          | Enable TLS for gRPC connections     | `false`                    | `WithGRPCTLS(bool)`            |
| `RPCURL`           | RPC endpoint for Cosmos node        | `"http://localhost:26657"` | `WithRPCURL(string)`           |
| `RPCWebsocketPath` | WebSocket path for RPC connection   | `"/websocket"`             | `WithRPCWebsocketPath(string)` |
| `KeyringUID`       | Identifier for keyring account      | `""`                       | `WithKeyringUID(string)`       |
| `KeyringBackend`   | Backend used for keyring            | `""`                       | `WithKeyringBackend(string)`   |
| `KeyringRootDir`   | Root directory path for the keyring | `""`                       | `WithKeyringRootDir(string)`   |
| `Memo`             | Transaction memo                    | `""`                       | `WithMemo(string)`             |

## Updating an existing client

Options can be updated on an instantiated client.

```go
err = cs.ApplyOptions(
    cosmosign.WithKeyringUID("another-signer"),
    cosmosign.WithGasMulplier(2.0),
    cosmosign.WithGasPrices("0.025ustake"),
    cosmosign.WithMemo("doge ftw"),
)
```

## License

This project is licensed under the Apache License, Version 2.0. See the
[LICENSE][1] file for more details.

[1]: https://github.com/shapeshed/cosmosign/blob/main/LICENCE
