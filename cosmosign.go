package cosmosign

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cometbft/cometbft/rpc/client/http"
)

const (
	DefaultAddressPrefix    = "cosmos"
	DefaultGasMultiplier    = 1.0
	DefaultGasPrices        = "0.0ustake"
	DefaultGRPCURL          = "localhost:1919"
	DefaultGRPCTLS          = false
	DefaultRPCURL           = "http://localhost:26657"
	DefaultRPCWebsocketPath = "/websocket"
	DefaultTimeout          = 5 * time.Second
)

type Cosmosign struct {
	accountQueryClient authtypes.QueryClient
	addressPrefix      string
	ctx                context.Context
	chainID            string
	encodingConfig     testutil.TestEncodingConfig
	fees               sdktypes.Coins
	feeGranter         sdktypes.AccAddress
	feePayer           sdktypes.AccAddress
	gas                uint64
	gasPrices          sdktypes.DecCoins
	gasMultiplier      *float64
	grpcConn           *grpc.ClientConn
	keyring            keyring.Keyring
	keyringBackend     string
	keyringRootDir     string
	keyringUID         string
	memo               string
	mu                 sync.Mutex
	rpcClient          *http.HTTP
	rpcURL             string
	rpcWebsocketPath   string
	txSvcClient        txtypes.ServiceClient
	grpcURL            string
	grpcTLS            bool
}

// NewClient initializes a new cosmosign instance
func NewClient(opts ...Option) (*Cosmosign, error) {
	ctx := context.Background()
	client := &Cosmosign{
		addressPrefix:    DefaultAddressPrefix,
		ctx:              ctx,
		rpcURL:           DefaultRPCURL,
		rpcWebsocketPath: DefaultRPCWebsocketPath,
		grpcURL:          DefaultGRPCURL,
		grpcTLS:          DefaultGRPCTLS,
	}

	var err error

	for _, opt := range opts {
		opt(client)
	}

	if client.addressPrefix != "" {
		client.mu.Lock()
		defer client.mu.Unlock()
		config := sdktypes.GetConfig()
		config.SetBech32PrefixForAccount(client.addressPrefix, client.addressPrefix+"pub")
	}

	if client.rpcClient == nil {
		rpcURL := client.rpcURL
		websocketPath := client.rpcWebsocketPath
		if client.rpcClient, err = http.New(rpcURL, websocketPath); err != nil {
			return nil, err
		}
	}

	if client.grpcConn == nil {
		if client.grpcConn, err = setupGRPCConnection(client.grpcURL, client.grpcTLS); err != nil {
			return nil, err
		}
	}

	// Set the chainID by querying the node and using the `NodeInfo.Network` value
	statusResp, err := client.rpcClient.Status(ctx)
	if err != nil {
		return nil, err
	}
	client.chainID = statusResp.NodeInfo.Network

	client.accountQueryClient = authtypes.NewQueryClient(client.grpcConn)
	client.txSvcClient = txtypes.NewServiceClient(client.grpcConn)
	client.encodingConfig = testutil.MakeTestEncodingConfig()

	if client.keyring == nil {
		if client.keyring, err = keyring.New(client.addressPrefix, client.keyringBackend, client.keyringRootDir, nil, client.encodingConfig.Codec); err != nil {
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
