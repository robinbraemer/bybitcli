package linear

import (
	"github.com/robinbraemer/bybitcli/util"
	"time"
)

type Trade struct {
	Symbol        string      `json:"symbol"`
	TickDirection string      `json:"tick_direction"`
	Price         util.Number `json:"price"`
	Size          util.Number `json:"size"`
	Timestamp     time.Time   `json:"timestamp"`
	TradeTimeMs   string      `json:"trade_time_ms"`
	Side          Side        `json:"side"`
	TradeID       string      `json:"trade_id"`
}
