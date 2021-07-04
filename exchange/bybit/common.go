package bybit

import (
	"context"
	"github.com/robinbraemer/bybitbot/util"
	"math"
	"time"
)

type Base interface {
	ServerTime() util.Number
	ReturnCode() int
	ReturnMessage() string
}

type BaseResult struct {
	RetCode         int         `json:"ret_code"`
	RetMsg          string      `json:"ret_msg"`
	ExtCode         string      `json:"ext_code"`
	Result          interface{} `json:"result"`
	TimeNow         util.Number `json:"time_now"`
	RateLimitStatus int         `json:"rate_limit_status"`
}

func (b *BaseResult) ServerTime() util.Number { return b.TimeNow }
func (b *BaseResult) ReturnCode() int         { return b.RetCode }
func (b *BaseResult) ReturnMessage() string   { return b.RetMsg }

func ParseTime(unixSec util.Number) (time.Time, error) {
	sec, dec := math.Modf(float64(unixSec))
	return time.Unix(int64(sec), int64(dec*(1e9))), nil
}

func UpdateServerTime(ctx context.Context, serverTime time.Time) {
	c := FromContext(ctx)
	c.ServerTimeOffset = time.Since(serverTime)
	c.LastTimeSync = time.Now()
}
