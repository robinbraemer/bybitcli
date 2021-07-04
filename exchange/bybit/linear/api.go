package linear

import (
	"context"
	"encoding/json"
	"github.com/robinbraemer/bybitbot/exchange/bybit"
	"github.com/robinbraemer/bybitbot/exchange/bybit/common"
	. "github.com/robinbraemer/bybitbot/util"
	"net/http"
	"time"
)

type (
	ActiveOrdersOptions struct {
		Symbol      string          `json:"symbol"` // required
		OrderStatus OrderStatus     `json:"order_status"`
		Direction   SearchDirection `json:"direction"`
		Limit       int             `json:"limit"`
		Cursor      string          `json:"cursor"`
	}
	ActiveOrderList struct {
		bybit.BaseResult
		Result ActiveOrderListResult `json:"result"`
	}
	ActiveOrderListResult struct {
		Data   []*ActiveOrder `json:"data"`
		Cursor string         `json:"cursor"`
	}
	ActiveOrder struct {
		OrderID        string      `json:"order_id"`
		UserID         int         `json:"user_id"`
		Symbol         string      `json:"symbol"`
		Side           Side        `json:"side"`
		OrderType      string      `json:"order_type"`
		Price          Number      `json:"price"`
		Qty            Number      `json:"qty"`
		TimeInForce    TimeInForce `json:"time_in_force"`
		OrderStatus    OrderStatus `json:"order_status"`
		LastExecPrice  Number      `json:"last_exec_price"`
		CumExecQty     Number      `json:"cum_exec_qty"`
		CumExecValue   Number      `json:"cum_exec_value"`
		CumExecFee     Number      `json:"cum_exec_fee"`
		OrderLinkID    string      `json:"order_link_id"`
		ReduceOnly     bool        `json:"reduce_only"`
		CloseOnTrigger bool        `json:"close_on_trigger"`
		CreatedTime    time.Time   `json:"created_time"`
		UpdatedTime    time.Time   `json:"updated_time"`
	}
)

type Side string

const (
	Buy  Side = "Buy"
	Sell Side = "Sell"
)

func (s Side) Opposite() Side {
	if s == Buy {
		return Sell
	}
	return Buy
}

type TimeInForce string

const (
	GoodTillCancel    TimeInForce = "GoodTillCancel"
	ImmediateOrCancel TimeInForce = "ImmediateOrCancel"
	FillOrKill        TimeInForce = "FillOrKill"
	PostOnly          TimeInForce = "PostOnly"
)

type OrderStatus string

const (
	CreatedOrderStatus         OrderStatus = "Created" // order has been accepted by the system but not yet put through the matching engine
	RejectedOrderStatus        OrderStatus = "Rejected"
	PartiallyFilledOrderStatus OrderStatus = "PartiallyFilled"
	FilledOrderStatus          OrderStatus = "Filled"
	CancelledOrderStatus       OrderStatus = "Cancelled"
	PendingCancelOrderStatus   OrderStatus = "PendingCancel" // matching engine has received the cancellation request but it may not be canceled successfully
	NewOrderStatus             OrderStatus = "New"           // order has been placed successfully
)

type SearchDirection string

const (
	Next SearchDirection = "next"
	Prev SearchDirection = "prev"
)

// https://bybit-exchange.github.io/docs/linear/#t-getactive
func ActiveOrders(ctx context.Context, opts ActiveOrdersOptions) (*ActiveOrderList, error) {
	params := M{"symbol": opts.Symbol}
	if opts.OrderStatus != "" {
		params["order_status"] = opts.OrderStatus
	}
	if opts.Direction != "" {
		params["direction"] = opts.Direction
	}
	if opts.Limit != 0 {
		params["limit"] = opts.Limit
	}
	if opts.Cursor != "" {
		params["cursor"] = opts.Cursor
	}
	r := new(ActiveOrderList)
	return r, common.SignedRequest(ctx, http.MethodGet, "/private/linear/order/list", params, r)
}

type (
	PositionsOptions struct {
		Symbol string `json:"symbol"` // required
	}
	PositionList struct {
		bybit.BaseResult
		Result []*Position `json:"result"`
	}
	Position struct {
		UserID              int      `json:"user_id"`
		Symbol              string   `json:"symbol"`
		Side                Side     `json:"side"`
		Size                Number   `json:"size"`
		PositionValue       Number   `json:"position_value"`
		EntryPrice          Number   `json:"entry_price"`
		LiqPrice            Number   `json:"liq_price"`
		BustPrice           Number   `json:"bust_price"`
		Leverage            Number   `json:"leverage"`
		IsIsolated          bool     `json:"is_isolated"`
		AutoAddMargin       Number   `json:"auto_add_margin"`
		PositionMargin      Number   `json:"position_margin"`
		OccClosingFee       Number   `json:"occ_closing_fee"`
		RealisedPnl         Number   `json:"realised_pnl"`
		CumRealisedPnl      Number   `json:"cum_realised_pnl"`
		FreeQty             Number   `json:"free_qty"`
		TpSlMode            TpSlMode `json:"tp_sl_mode"`
		UnrealisedPnl       Number   `json:"unrealised_pnl"`
		DeleverageIndicator int      `json:"deleverage_indicator"`
		TrailingStop        Number   `json:"trailing_stop"`
		StopLoss            Number   `json:"stop_loss"`
		TakeProfit          Number   `json:"take_profit"`
		RiskID              int      `json:"risk_id"`
	}
)

type TpSlMode string

const (
	Full    TpSlMode = "Full"
	Partial TpSlMode = "Partial"
)

// https://bybit-exchange.github.io/docs/linear/#t-myposition
func Positions(ctx context.Context, opts PositionsOptions) (*PositionList, error) {
	r := new(PositionList)
	return r, common.SignedRequest(ctx, http.MethodGet, "/private/linear/position/list", M{"symbol": opts.Symbol}, r)
}

type (
	CreateOrderOptions struct {
		Side           Side        `json:"side,omitempty"`          // required
		Symbol         string      `json:"symbol,omitempty"`        // required
		OrderType      OrderType   `json:"order_type,omitempty"`    // required
		Qty            Number      `json:"qty,omitempty"`           // required
		Price          Number      `json:"price,omitempty"`         // required if doing limit price order
		TimeInForce    TimeInForce `json:"time_in_force,omitempty"` // required
		TakeProfit     Number      `json:"take_profit,omitempty"`
		StopLoss       Number      `json:"stop_loss,omitempty"`
		TpTriggerBy    TriggerBy   `json:"tp_trigger_by,omitempty"`
		SlTriggerBy    TriggerBy   `json:"sl_trigger_by,omitempty"`
		ReduceOnly     bool        `json:"reduce_only"`      // required
		CloseOnTrigger bool        `json:"close_on_trigger"` // required
		OrderLinkID    string      `json:"order_link_id,omitempty"`
	}
	CreateOrderResponse struct {
		bybit.BaseResult
		Result CreateOrderResult `json:"result"`
	}
	CreateOrderResult struct {
		OrderID        string      `json:"order_id"`
		UserID         int         `json:"user_id"`
		Symbol         string      `json:"symbol"`
		Side           Side        `json:"side"`
		OrderType      OrderType   `json:"order_type"`
		Price          Number      `json:"price"`
		Qty            Number      `json:"qty"`
		TimeInForce    TimeInForce `json:"time_in_force"`
		OrderStatus    OrderStatus `json:"order_status"`
		LastExecPrice  Number      `json:"last_exec_price"`
		CumExecQty     Number      `json:"cum_exec_qty"`
		CumExecValue   Number      `json:"cum_exec_value"`
		CumExecFee     Number      `json:"cum_exec_fee"`
		SlTriggerBy    TriggerBy   `json:"sl_trigger_by"`
		TpTriggerBy    TriggerBy   `json:"tp_trigger_by"`
		StopLoss       Number      `json:"stop_loss"`
		TakeProfit     Number      `json:"take_profit"`
		ReduceOnly     bool        `json:"reduce_only"`
		CloseOnTrigger bool        `json:"close_on_trigger"`
		OrderLinkID    string      `json:"order_link_id"`
		CreatedTime    time.Time   `json:"created_time"`
		UpdatedTime    time.Time   `json:"updated_time"`
	}
)

// https://bybit-exchange.github.io/docs/linear/#t-placeactive
func CreateOrder(ctx context.Context, opts CreateOrderOptions) (*CreateOrderResponse, error) {
	params, err := toMap(opts)
	if err != nil {
		return nil, err
	}
	r := new(CreateOrderResponse)
	return r, common.SignedRequest(ctx, http.MethodPost, "/private/linear/order/create", params, r)
}

type TriggerBy string

const (
	LastPrice  TriggerBy = "LastPrice"
	IndexPrice TriggerBy = "IndexPrice"
	MarkPrice  TriggerBy = "MarkPrice"
	Unknown    TriggerBy = "UNKNOWN" // returned by CreateOrderResult
)

type OrderType string

const (
	Limit  OrderType = "Limit"
	Market OrderType = "Market"
)

type SortOrder string

const (
	Desc SortOrder = "desc"
	Asc  SortOrder = "asc"
)

func toMap(v interface{}) (M, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	m := M{}
	return m, json.Unmarshal(b, &m)
}
