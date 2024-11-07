package cosmosign

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	DefaultAddressPrefix = "cosmos"
	DefaultGasMultiplier = 1.0
	DefaultGasPrices     = "0.0ustake"
	DefaultTimeout       = 5 * time.Second
)

type Cosmosign struct {
	accountQueryClient authtypes.QueryClient
	addressPrefix      string
	chainID            string
	cmtSvcClient       cmtservice.ServiceClient
	ctx                context.Context
	encodingConfig     testutil.TestEncodingConfig
	feeGranter         sdktypes.AccAddress
	feePayer           sdktypes.AccAddress
	fees               sdktypes.Coins
	gasMultiplier      *float64
	gasPrices          sdktypes.DecCoins
	gas                uint64
	grpcConn           *grpc.ClientConn
	keyringBackend     string
	keyringAppName     string
	keyring            keyring.Keyring
	keyringRootDir     string
	keyringUID         string
	memo               string
	mu                 sync.Mutex
	txSvcClient        txtypes.ServiceClient
}

// NewClient initializes a new cosmosign instance
func NewClient(ctx context.Context, opts ...Option) (*Cosmosign, error) {
	var err error

	client := &Cosmosign{
		ctx: ctx,
	}

	for _, opt := range opts {
		opt(client)
	}

	if client.grpcConn == nil {
		return nil, ErrGRPCClientIsNil
	}

	client.cmtSvcClient = cmtservice.NewServiceClient(client.grpcConn)
	client.accountQueryClient = authtypes.NewQueryClient(client.grpcConn)
	client.txSvcClient = txtypes.NewServiceClient(client.grpcConn)
	client.encodingConfig = testutil.MakeTestEncodingConfig()

	// Set the address prefix via a GRPC query
	err = client.setaddressPrefix()
	if err != nil {
		return nil, err
	}

	// Set the chainID via a GRPC query
	err = client.setChainID()
	if err != nil {
		return nil, err
	}

	if client.keyring == nil {
		if client.keyring, err = keyring.New(client.keyringAppName, client.keyringBackend, client.keyringRootDir, nil, client.encodingConfig.Codec); err != nil {
			return nil, err
		}
	}

	if client.gasPrices == nil {
		parsedGasPrices, err := sdktypes.ParseDecCoins(DefaultGasPrices)
		if err != nil {
			panic(err)
		}
		client.gasPrices = parsedGasPrices
	}

	if client.gasMultiplier == nil {
		gasMultiplier := new(float64)
		*gasMultiplier = DefaultGasMultiplier
		client.gasMultiplier = gasMultiplier
	}

	return client, nil
}

func (c *Cosmosign) setaddressPrefix() error {
	prefix, err := c.accountQueryClient.Bech32Prefix(c.ctx, &authtypes.Bech32PrefixRequest{})
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.addressPrefix = prefix.Bech32Prefix

	return nil
}

func (c *Cosmosign) setChainID() error {
	nodeInfo, err := c.cmtSvcClient.GetNodeInfo(c.ctx, &cmtservice.GetNodeInfoRequest{})
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.chainID = nodeInfo.DefaultNodeInfo.Network

	return nil
}
