package cosmosign

import (
	"google.golang.org/grpc"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// Option is a function that configures the Cosmosign client.
type Option func(*Cosmosign)

// WithAddressPrefix sets the addressPrefix for the client.
func WithAddressPrefix(addressPrefix string) Option {
	return func(c *Cosmosign) {
		c.addressPrefix = addressPrefix
		c.mu.Lock()
		defer c.mu.Unlock()
		config := sdktypes.GetConfig()
		config.SetBech32PrefixForAccount(c.addressPrefix, c.addressPrefix+"pub")
	}
}

// WithFees sets the fees for the client.
func WithFees(fees string) Option {
	return func(c *Cosmosign) {
		parsedFees, err := sdktypes.ParseCoinsNormalized(fees)
		if err != nil {
			panic(err)
		}
		c.fees = parsedFees
	}
}

// WithFeeGranter sets the fee granter for the client.
func WithFeeGranter(feeGranter sdktypes.AccAddress) Option {
	return func(c *Cosmosign) {
		c.feeGranter = feeGranter
	}
}

// WithFeePayer sets the fee payer for the client.
func WithFeePayer(feePayer sdktypes.AccAddress) Option {
	return func(c *Cosmosign) {
		c.feePayer = feePayer
	}
}

// WithGas sets the gas limit for the client.
func WithGas(gas uint64) Option {
	return func(c *Cosmosign) {
		c.gas = gas
	}
}

// WithGasPrices sets the gas prices used for the client.
func WithGasPrices(gasPrices string) Option {
	return func(c *Cosmosign) {
		parsedGasPrices, err := sdktypes.ParseDecCoins(gasPrices)
		if err != nil {
			panic(err)
		}
		c.gasPrices = parsedGasPrices
	}
}

// WithGasPrices sets the multipler for gas simulation amount.
func WithGasMultipler(gasMultiplier float64) Option {
	return func(c *Cosmosign) {
		c.gasMultiplier = new(float64)
		*c.gasMultiplier = gasMultiplier
	}
}

// WithGasPrices sets the multipler for gas simulation amount.
func WithGRPCConn(grpcConn *grpc.ClientConn) Option {
	return func(c *Cosmosign) {
		c.grpcConn = grpcConn
	}
}

// WithKeyringUID sets the keyring uid (account) to use in signing.
func WithKeyringAppName(keyringAppName string) Option {
	return func(c *Cosmosign) {
		c.keyringAppName = keyringAppName
	}
}

// WithKeyringUID sets the keyring uid (account) to use in signing.
func WithKeyringUID(keyringUID string) Option {
	return func(c *Cosmosign) {
		c.keyringUID = keyringUID
	}
}

// WithKeyringBackend sets the backend to use for the keyring.
func WithKeyringBackend(keyringBackend string) Option {
	return func(c *Cosmosign) {
		c.keyringBackend = keyringBackend
	}
}

// WithKeyringBackend sets the backend to use for the keyring.
func WithKeyringRootDir(keyringRootDir string) Option {
	return func(c *Cosmosign) {
		c.keyringRootDir = keyringRootDir
	}
}

// WithMemo sets the memo to use on the transaction.
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
