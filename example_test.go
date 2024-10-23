package cosmosign_test

import (
	"log"

	"github.com/shapeshed/cosmosign"
)

func ExampleNewClient() {
	// Initialize cosmosign client (this requires network connectivity)
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

	// Log the initialized client
	log.Printf("Cosmosign client initialized: %+v", cs)

	// Note: This example will not run because it requires external network connectivity
	// Output is omitted intentionally to avoid failures in testing environments.
}

func ExampleCosmosign_ApplyOptions() {
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

	err = cs.ApplyOptions(
		cosmosign.WithGasPrices("0.01ustake"),
		cosmosign.WithRPCURL("http://localhost:26659"),
	)
	if err != nil {
		log.Fatalf("Failed to apply options: %v", err)
	}

	log.Printf("Cosmosign client after applying options: %+v", cs)

	// This example will not run as it requires network connectivity
}
