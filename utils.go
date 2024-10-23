package cosmosign

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/math"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func calcGasAndFees(
	simRes *txtypes.SimulateResponse, // Simulation response
	gasPrices sdktypes.DecCoins, // Gas price from the feemarket module
	multiplier float64, // Multiplier to adjust gas estimate
) (adjustedGas uint64, fees sdktypes.Coins) {
	// Step 1: Calculate adjusted gas with multiplier and buffer
	adjustedGas = uint64(float64(simRes.GasInfo.GasUsed) * multiplier)

	// Step 2: Convert adjusted gas to sdk.Dec (LegacyDec) for precision
	glDec := math.LegacyNewDecFromBigInt(new(big.Int).SetUint64(adjustedGas))

	// Step 3: Create a slice of sdk.Coins to store the calculated fees
	fees = make(sdktypes.Coins, len(gasPrices))

	// Step 4: Loop through each gas price in gasPrices and calculate fees
	for i, gp := range gasPrices {
		// Multiply the gas price amount by the adjusted gas
		feeAmount := gp.Amount.Mul(glDec)

		// Step 5: Round the fee to integer and store in the fees array
		fees[i] = sdktypes.NewCoin(gp.Denom, feeAmount.Ceil().RoundInt())
	}
	return adjustedGas, fees
}

// getAccountNumberAndSequence fetches the account information and returns the account number and sequence

func (c *Cosmosign) getAccountNumberAndSequence(address sdktypes.AccAddress) (uint64, uint64, error) {
	accountRes, err := c.accountQueryClient.Account(c.ctx, &authtypes.QueryAccountRequest{Address: address.String()})
	if err != nil {
		return 0, 0, err
	}

	// Unmarshal the account into BaseAccount
	ba := authtypes.BaseAccount{}
	err = ba.Unmarshal(accountRes.Account.Value)
	if err != nil {
		return 0, 0, err
	}

	// Return AccountNumber and Sequence
	return ba.AccountNumber, ba.Sequence, nil
}

// SetupGRPCConnection establishes a GRPC connection, optionally using system's TLS certificates
func SetupGRPCConnection(address string, useTLS bool) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if useTLS {
		// Load the system's certificate pool
		systemCertPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to load system certificates: %w", err)
		}

		// Create TLS credentials using the system's certificate pool
		creds := credentials.NewTLS(&tls.Config{
			RootCAs:    systemCertPool,
			MinVersion: tls.VersionTLS12,
		})

		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		// Use insecure credentials if TLS is not required
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Dial the GRPC server with the appropriate credentials
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", address, err)
	}

	return conn, nil
}
