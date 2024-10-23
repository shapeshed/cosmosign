package cosmosign

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// Option is a function that configures the Cosmosign client.
type Option func(*Cosmosign)

// WithAddressPrefix sets the comsos addressPrefix for the client
func WithAddressPrefix(addressPrefix string) Option {
	return func(c *Cosmosign) {
		c.addressPrefix = addressPrefix
		c.mu.Lock()
		defer c.mu.Unlock()
		config := sdktypes.GetConfig()
		config.SetBech32PrefixForAccount(c.addressPrefix, c.addressPrefix+"pub")
	}
}

// WithGasToken sets the gas token used for fees
func WithFees(fees string) Option {
	return func(c *Cosmosign) {
		parsedFees, err := sdktypes.ParseCoinsNormalized(fees)
		if err != nil {
			panic(err)
		}
		c.fees = parsedFees
	}
}

// WithFeeGranter returns a copy of the Factory with an updated fee granter.
func WithFeeGranter(feeGranter sdktypes.AccAddress) Option {
	return func(c *Cosmosign) {
		c.feeGranter = feeGranter
	}
}

// WithFeeGranter returns a copy of the Factory with an updated fee granter.
func WithFeePayer(feePayer sdktypes.AccAddress) Option {
	return func(c *Cosmosign) {
		c.feePayer = feePayer
	}
}

// WithGasToken sets the gas token used for fees
func WithGas(gas uint64) Option {
	return func(c *Cosmosign) {
		c.gas = gas
	}
}

// WithGasToken sets the gas token used for fees
func WithGasPrices(gasPrices string) Option {
	return func(c *Cosmosign) {
		parsedGasPrices, err := sdktypes.ParseDecCoins(gasPrices)
		if err != nil {
			panic(err)
		}
		c.gasPrices = parsedGasPrices
	}
}

// WithGasToken sets the gas token used for fees
func WithGasMultipler(gasMultiplier float64) Option {
	return func(c *Cosmosign) {
		c.gasMultiplier = new(float64)
		*c.gasMultiplier = gasMultiplier
	}
}

// WithGRPCAddr sets the gRPC address for the client
func WithGRPCURL(addr string) Option {
	return func(c *Cosmosign) {
		c.grpcURL = addr
	}
}

// WithGRPCTLS sets the gRPC address for the client
func WithGRPCTLS(grpcTLS bool) Option {
	return func(c *Cosmosign) {
		c.grpcTLS = grpcTLS
	}
}

// WithRPCAddr sets the RPC address for the client
func WithRPCURL(addr string) Option {
	return func(c *Cosmosign) {
		c.rpcURL = addr
	}
}

// WithRPCWebsocketPath sets the RPC websocket path for the client
func WithRPCWebsocketPath(path string) Option {
	return func(c *Cosmosign) {
		c.rpcWebsocketPath = path
	}
}

// WithKeyringUID sets the keyring uid (account) to use in signing
func WithKeyringUID(keyringUID string) Option {
	return func(c *Cosmosign) {
		c.keyringUID = keyringUID
	}
}

// WithKeyringBackend sets the backend to use for the keyring
func WithKeyringBackend(keyringBackend string) Option {
	return func(c *Cosmosign) {
		c.keyringBackend = keyringBackend
	}
}

// WithKeyringBackend sets the backend to use for the keyring
func WithKeyringRootDir(keyringRootDir string) Option {
	return func(c *Cosmosign) {
		c.keyringRootDir = keyringRootDir
	}
}

// WithMemo sets the memo to use on the transaction
func WithMemo(memo string) Option {
	return func(c *Cosmosign) {
		c.memo = memo
	}
}

// ApplyOptions applies options to the running client
func (c *Cosmosign) ApplyOptions(opts ...Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, opt := range opts {
		opt(c)
	}

	return nil
}
