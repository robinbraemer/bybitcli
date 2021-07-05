package bybit

import (
	"context"
	"github.com/robinbraemer/bybitcli/util"
	"net/http"
	"time"
)

type Options struct {
	APIKey, SecretKey string // Required
	Testnet           bool   // Whether to use testnet
	Log               util.Logger
	Client            *http.Client
}

func New(ctx context.Context, opts Options) context.Context {
	if opts.Log == nil {
		opts.Log = util.NopLog
	}
	if opts.Client == nil {
		opts.Client = http.DefaultClient
	}
	return contextWithClient(ctx, &Context{Options: opts})
}

type Context struct {
	Options
	LastTimeSync     time.Time
	ServerTimeOffset time.Duration
}

func FromContext(ctx context.Context) *Context {
	return ctx.Value(ctxKey).(*Context)
}

func IsTestnet(ctx context.Context) bool {
	return FromContext(ctx).Options.Testnet
}

func contextWithClient(ctx context.Context, client *Context) context.Context {
	return context.WithValue(ctx, ctxKey, client)
}

type contextKey struct{}

var ctxKey = &contextKey{}
