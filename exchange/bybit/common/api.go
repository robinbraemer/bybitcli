package common

import (
	"context"
	"github.com/robinbraemer/bybitcli/exchange/bybit"
	"github.com/robinbraemer/bybitcli/util"
	"net/http"
	"time"
)

const (
	MainnetEndpoint = "https://api.bybit.com"
	TestnetEndpoint = "https://api-testnet.bybit.com"

	//MainnetWSPublicEndpoint  = "wss://stream.bybit.com/realtime_public"
	//MainnetWSPrivateEndpoint = "wss://stream.bybit.com/realtime_private"
	//
	//TestnetWSPublicEndpoint  = "wss://stream-testnet.bybit.com/realtime_public"
	//TestnetWSPrivateEndpoint = "wss://stream-testnet.bybit.com/realtime_private"
)

func SyncTime(ctx context.Context) error {
	_, err := ServerTime(ctx)
	return err
}

// ServerTime gets the server time.
func ServerTime(ctx context.Context) (time.Time, error) {
	ret := new(bybit.BaseResult)
	err := PublicRequest(ctx, http.MethodGet, "/v2/public/time", nil, ret)
	if err != nil {
		return time.Time{}, nil
	}
	t, err := bybit.ParseTime(ret.TimeNow)
	if err != nil {
		return time.Time{}, err
	}
	bybit.UpdateServerTime(ctx, t)
	return t, nil
}

type (
	SymbolList struct {
		bybit.BaseResult
		Result []*Symbol `json:"result"`
	}
	Symbol struct {
		Name           string      `json:"name"`
		Alias          string      `json:"alias"`
		Status         string      `json:"status"`
		BaseCurrency   string      `json:"base_currency"`
		QuoteCurrency  string      `json:"quote_currency"`
		PriceScale     int         `json:"price_scale"`
		TakerFee       util.Number `json:"taker_fee"`
		MakerFee       util.Number `json:"maker_fee"`
		LeverageFilter struct {
			MinLeverage  int         `json:"min_leverage"`
			MaxLeverage  int         `json:"max_leverage"`
			LeverageStep util.Number `json:"leverage_step"`
		} `json:"leverage_filter"`
		PriceFilter struct {
			MinPrice util.Number `json:"min_price"`
			MaxPrice util.Number `json:"max_price"`
			TickSize util.Number `json:"tick_size"`
		} `json:"price_filter"`
		LotSizeFilter struct {
			MaxTradingQty util.Number `json:"max_trading_qty"`
			MinTradingQty util.Number `json:"min_trading_qty"`
			QtyStep       util.Number `json:"qty_step"`
		} `json:"lot_size_filter"`
	}
)

func Symbols(ctx context.Context) (*SymbolList, error) {
	r := new(SymbolList)
	return r, PublicRequest(ctx, http.MethodGet, "/v2/public/symbols", nil, r)
}
