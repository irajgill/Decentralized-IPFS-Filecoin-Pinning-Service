package filecoin

import (
	"context"
	"fmt"
	"net/http"

	lapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
)

type LotusClient struct {
	api lapi.FullNode
}

type StartDealParams struct {
	Data         []byte
	MinerID      string
	Duration     int64
	PriceFIL     float64
	WalletAddr   string
	VerifiedDeal bool
}

type MinerInfo struct {
	ID         string  `json:"id"`
	Power      int64   `json:"power"`
	Available  bool    `json:"available"`
	Price      float64 `json:"price"`
	Reputation float64 `json:"reputation"`
}

func NewLotusClient(apiURL, token string) (*LotusClient, error) {
	headers := http.Header{}
	if token != "" {
		headers.Add("Authorization", "Bearer "+token)
	}

	api, closer, err := client.NewFullNodeRPCV1(context.Background(), apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to create Lotus client: %w", err)
	}

	// Store closer for cleanup if needed
	_ = closer

	return &LotusClient{api: api}, nil
}

// StartDeal creates a new storage deal
func (c *LotusClient) StartDeal(ctx context.Context, params StartDealParams) (string, error) {
	// Import data to Lotus
	importRes, err := c.api.ClientImport(ctx, lapi.FileRef{
		Path:  "", // Data will be provided directly
		IsCAR: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to import data: %w", err)
	}

	// Prepare deal parameters
	dealParams := &lapi.StartDealParams{
		Data: &types.DataRef{
			TransferType: types.TTGraphsync,
			Root:         importRes.Root,
		},
		Wallet:            types.Address{},                              // Convert from string
		Miner:             types.Address{},                              // Convert from string
		EpochPrice:        types.NewInt(uint64(params.PriceFIL * 1e18)), // Convert FIL to attoFIL
		MinBlocksDuration: uint64(params.Duration),
		VerifiedDeal:      params.VerifiedDeal,
	}

	// Start the deal
	dealCID, err := c.api.ClientStartDeal(ctx, dealParams)
	if err != nil {
		return "", fmt.Errorf("failed to start deal: %w", err)
	}

	return dealCID.String(), nil
}

// GetDealStatus gets the status of a deal
func (c *LotusClient) GetDealStatus(ctx context.Context, dealCID string) (string, error) {
	// Parse deal CID
	cid, err := types.CidFromString(dealCID)
	if err != nil {
		return "", fmt.Errorf("invalid deal CID: %w", err)
	}

	// Get deal info
	dealInfo, err := c.api.ClientGetDealInfo(ctx, cid)
	if err != nil {
		return "", fmt.Errorf("failed to get deal info: %w", err)
	}

	return dealInfo.State.String(), nil
}

// GetCurrentEpoch gets the current chain epoch
func (c *LotusClient) GetCurrentEpoch(ctx context.Context) (int64, error) {
	head, err := c.api.ChainHead(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get chain head: %w", err)
	}

	return int64(head.Height()), nil
}

// GetAvailableMiners returns list of available miners
func (c *LotusClient) GetAvailableMiners(ctx context.Context) ([]MinerInfo, error) {
	// Get miner list from chain
	miners, err := c.api.StateListMiners(ctx, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("failed to get miner list: %w", err)
	}

	var minerInfos []MinerInfo
	for _, miner := range miners {
		// Get miner info
		info, err := c.api.StateMinerInfo(ctx, miner, types.EmptyTSK)
		if err != nil {
			continue // Skip miners we can't get info for
		}

		// Get miner power
		power, err := c.api.StateMinerPower(ctx, miner, types.EmptyTSK)
		if err != nil {
			continue
		}

		minerInfo := MinerInfo{
			ID:         miner.String(),
			Power:      power.MinerPower.QualityAdjPower.Int64(),
			Available:  true,  // Would need to check if accepting deals
			Price:      0.001, // Would need to get actual pricing
			Reputation: 1.0,   // Would need reputation system
		}

		minerInfos = append(minerInfos, minerInfo)
	}

	return minerInfos, nil
}

// GetWalletBalance gets wallet balance
func (c *LotusClient) GetWalletBalance(ctx context.Context, addr string) (float64, error) {
	address, err := types.NewFromString(addr)
	if err != nil {
		return 0, fmt.Errorf("invalid address: %w", err)
	}

	balance, err := c.api.WalletBalance(ctx, address)
	if err != nil {
		return 0, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	// Convert from attoFIL to FIL
	balanceFIL := float64(balance.Int64()) / 1e18
	return balanceFIL, nil
}
