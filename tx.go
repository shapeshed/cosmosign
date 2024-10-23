package cosmosign

import (
	"context"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

const (
	queryInterval = 1 * time.Second
)

// signTransaction signs msg bytes with the provided signer data
func (c *Cosmosign) signTransaction(
	txBuilder client.TxBuilder,
	signerData authsigning.SignerData,
	sequence uint64,
) error {
	sigV2 := signing.SignatureV2{
		PubKey: signerData.PubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: sequence,
	}
	err := txBuilder.SetSignatures(sigV2)
	if err != nil {
		return err
	}

	bytesToSign, err := authsigning.GetSignBytesAdapter(
		c.ctx,
		c.encodingConfig.TxConfig.SignModeHandler(),
		signing.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder.GetTx())
	if err != nil {
		return err
	}

	sigBytes, _, err := c.keyring.Sign(c.keyringUID, bytesToSign, signing.SignMode_SIGN_MODE_DIRECT)
	if err != nil {
		return err
	}

	sigData := signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: sigBytes,
	}
	sig := signing.SignatureV2{
		PubKey:   signerData.PubKey,
		Data:     &sigData,
		Sequence: sequence,
	}

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return err
	}

	return err
}

// SimulateTransaction simulates the transaction
func (c *Cosmosign) simulateTransaction(
	ctx context.Context,
	txBytes []byte,
) (*txtypes.SimulateResponse, error) {
	simRes, err := c.txSvcClient.Simulate(ctx, &txtypes.SimulateRequest{
		TxBytes: txBytes,
	})
	if err != nil {
		return nil, err
	}
	return simRes, nil
}

// BroadcastTransaction broadcasts the transaction to the network
func (c *Cosmosign) broadcastTransaction(
	ctx context.Context,
	txBytes []byte,
) (*txtypes.BroadcastTxResponse, error) {
	res, err := c.txSvcClient.BroadcastTx(ctx, &txtypes.BroadcastTxRequest{
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
		TxBytes: txBytes,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// SendMessages signs and broadcasts a transaction with one or more messages
func (c *Cosmosign) SendMessages(
	msgs ...sdktypes.Msg,
) (*txtypes.BroadcastTxResponse, error) {
	// Initialize the txBuilder
	txBuilder := c.encodingConfig.TxConfig.NewTxBuilder()

	// Set memo if set
	if c.memo != "" {
		txBuilder.SetMemo(c.memo)
	}

	// Set feepayer if set
	if c.feePayer != nil {
		txBuilder.SetFeePayer(c.feePayer)
	}

	// Set feegranter if set
	if c.feeGranter != nil {
		txBuilder.SetFeeGranter(c.feeGranter)
	}

	// Set the messages into the transaction builder
	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}

	// Get the signer
	signer, err := c.keyring.Key(c.keyringUID)
	if err != nil {
		return nil, err
	}

	// Get the signers public key
	pubKey, err := signer.GetPubKey()
	if err != nil {
		return nil, err
	}

	// Get the signer addresss as sdk.AccAddress
	signerAddr, err := signer.GetAddress()
	if err != nil {
		return nil, err
	}

	// Fetch the account number and sequence for the signer
	accountNumber, sequence, err := c.getAccountNumberAndSequence(signerAddr)
	if err != nil {
		return nil, err
	}

	// Populate signer data
	signerData := authsigning.SignerData{
		ChainID:       c.chainID,
		AccountNumber: accountNumber,
		Sequence:      sequence,
		PubKey:        pubKey,
		Address:       signerAddr.String(),
	}

	// Sign the transaction ahead of running the tx simulation
	err = c.signTransaction(txBuilder, signerData, sequence)
	if err != nil {
		return nil, err
	}

	// Get the encoded tx bytes
	simtxBytes, err := c.encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	// Run the simulation
	simulationRes, err := c.simulateTransaction(c.ctx, simtxBytes)
	if err != nil {
		return nil, err
	}

	// Calculate gas and fees using the simulation results
	adjustedGas, fees := calcGasAndFees(simulationRes, c.gasPrices, *c.gasMultiplier)
	if err != nil {
		return nil, err
	}

	// Set the gas limit and fee in the transaction builder
	txBuilder.SetGasLimit(adjustedGas)
	txBuilder.SetFeeAmount(fees)

	// Sign again after updating Gas and Fee
	err = c.signTransaction(txBuilder, signerData, sequence)
	if err != nil {
		return nil, err
	}

	// Encode the transaction to bytes
	txBytes, err := c.encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	// Broadcast the transaction
	res, err := c.broadcastTransaction(c.ctx, txBytes)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SendMessagesWaitTx broadcasts a message and waits for it to be confirmed, returning the result.
func (c *Cosmosign) SendMessagesWaitTx(
	msgs ...sdktypes.Msg,
) (*txtypes.GetTxResponse, error) {
	// Call SendMessages to broadcast the transaction
	broadcastRes, err := c.SendMessages(msgs...)
	if err != nil {
		return nil, err
	}

	// Get the transaction hash from the broadcast result
	txHash := broadcastRes.TxResponse.TxHash

	timeout := 180 * time.Second
	// Wait for the transaction to be confirmed
	confirmedRes, err := c.waitForTx(c.ctx, txHash, timeout)
	if err != nil {
		return nil, err
	}
	// Return the full transaction result (confirmed in a block)
	return confirmedRes, nil
}

// waitForTx polls for a confirmed transaction, returning it when found
func (c *Cosmosign) waitForTx(ctx context.Context, hash string, timeout time.Duration) (*txtypes.GetTxResponse, error) {
	txSvcClient := txtypes.NewServiceClient(c.grpcConn)

	// Create a new context with a timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	tick := time.NewTicker(queryInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctxWithTimeout.Done():
			// If the context deadline or timeout is reached, return an error
			return nil, ctxWithTimeout.Err()
		case <-tick.C:
			// Query the transaction by hash
			txRes, err := txSvcClient.GetTx(ctx, &txtypes.GetTxRequest{Hash: hash})
			if err == nil {
				// Transaction found
				return txRes, nil
			}

			// If transaction not found, retry unless a different error occurs
			if !strings.Contains(err.Error(), "not found") {
				return nil, err
			}
		}
	}
}
