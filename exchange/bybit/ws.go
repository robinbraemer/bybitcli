package bybit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/recws-org/recws"
	"github.com/robinbraemer/bybitcli/util"
	"github.com/tidwall/gjson"
	"time"
)

var WebsocketEndpoint = struct {
	Mainnet WebsocketEndpoints
	Testnet WebsocketEndpoints
}{
	Mainnet: WebsocketEndpoints{
		Public:  "wss://stream.bybit.com/realtime_public",
		Private: "wss://stream.bybit.com/realtime_private",
	},
	Testnet: WebsocketEndpoints{
		Public:  "wss://stream-testnet.bybit.com/realtime_public",
		Private: "wss://stream-testnet.bybit.com/realtime_private",
	},
}

type WebsocketEndpoints struct {
	Public  string
	Private string
}

//MainnetWSPublicEndpoint  = "wss://stream.bybit.com/realtime_public"
//MainnetWSPrivateEndpoint = "wss://stream.bybit.com/realtime_private"
//
//TestnetWSPublicEndpoint  = "wss://stream-testnet.bybit.com/realtime_public"
//TestnetWSPrivateEndpoint = "wss://stream-testnet.bybit.com/realtime_private"

type Websocket struct {
	Options
	public, private wsConn
	resChan         chan *WebsocketResponse

	handlerFn func(*WebsocketResponse)
}

type wsConn struct {
	endpoint string
	log      logr.Logger
	recws.RecConn
}

type WebsocketRequest struct {
	Op   string        `json:"op"`
	Args []interface{} `json:"args"`
}

func NewWebsocket(opts Options, handlerFn func(*WebsocketResponse)) *Websocket {
	if opts.Log == nil {
		opts.Log = util.NopLog
	}
	ep := WebsocketEndpoint.Testnet
	if opts.Mainnet {
		ep = WebsocketEndpoint.Mainnet
	}
	const keepAlive = time.Second * 30
	s := &Websocket{
		handlerFn: handlerFn,
		Options:   opts,
		resChan:   make(chan *WebsocketResponse),
		public: wsConn{
			endpoint: ep.Public,
			log:      opts.Log.WithName("public"),
			RecConn:  recws.RecConn{KeepAliveTimeout: keepAlive},
		},
		private: wsConn{
			endpoint: ep.Private,
			log:      opts.Log.WithName("private"),
			RecConn:  recws.RecConn{KeepAliveTimeout: keepAlive},
		},
	}
	go s.handleLoop()
	return s
}

type WebsocketResponse struct {
	Parse gjson.Result

	isTopic  *bool
	isStatus *bool
}

type (
	// WebsocketStatusResponse occurs on error or subscription confirmation
	WebsocketStatusResponse struct {
		Success bool             `json:"success"`
		RetMsg  string           `json:"ret_msg"`
		ConnID  string           `json:"conn_id"`
		Request WebsocketRequest `json:"request"`
	}
	WebsocketTopicResponse struct {
		Topic string          `json:"topic"`
		Type  string          `json:"type"`
		Data  json.RawMessage `json:"data"`
	}
)

func (r *WebsocketResponse) IsStatusResponse() bool {
	if r.isStatus == nil {
		r.isStatus = new(bool)
		*r.isStatus = r.Parse.Get("success").Exists()
	}
	return *r.isStatus
}

func (r *WebsocketResponse) IsTopicResponse() bool {
	if r.isTopic == nil {
		r.isTopic = new(bool)
		*r.isTopic = r.Parse.Get("topic").Exists()
	}
	return *r.isTopic
}

func (r *WebsocketResponse) TopicResponse() (*WebsocketTopicResponse, error) {
	if !r.IsTopicResponse() {
		return nil, errors.New("not a topic response")
	}
	res := new(WebsocketTopicResponse)
	return res, json.Unmarshal([]byte(r.Parse.Raw), res)
}

func (r *WebsocketResponse) StatusResponse() (*WebsocketStatusResponse, error) {
	if !r.IsStatusResponse() {
		return nil, errors.New("not a subscription response")
	}
	res := new(WebsocketStatusResponse)
	return res, json.Unmarshal([]byte(r.Parse.Raw), res)
}

func (s *Websocket) Close() {
	s.Log.Info("Closing websockets")
	s.private.Close()
	s.public.Close()
	close(s.resChan)
}

// Subscribe adds topic to subscription list.
func (s *Websocket) Subscribe(public bool, topics ...string) error {
	if topics == nil {
		return nil
	}
	c, err := s.connect(public)
	if err != nil {
		return err
	}
	return c.write(&WebsocketRequest{
		Op:   "subscribe",
		Args: toInterfaces(topics),
	})
}

// Unsubscribe removes topic from subscription list.
func (s *Websocket) Unsubscribe(public bool, topics ...string) error {
	if topics == nil {
		return nil
	}
	c, err := s.connect(public)
	if err != nil {
		return err
	}
	return c.write(&WebsocketRequest{
		Op:   "unsubscribe",
		Args: toInterfaces(topics),
	})
}

func toInterfaces(s []string) (i []interface{}) {
	for _, v := range s {
		i = append(i, v)
	}
	return i
}

// connect starts the connection and authenticates.
func (s *Websocket) connect(public bool) (*wsConn, error) {
	c := s.connection(public)
	if c.IsConnected() {
		return c, nil // already connected
	}
	c.log.Info("Connecting", "endpoint", c.endpoint)
	c.Dial(c.endpoint, nil)
	if err := c.GetDialError(); err != nil {
		return c, err
	}
	if !public {
		return c, c.authenticate(s.APIKey, s.SecretKey)
	}
	go c.readLoop(s.resChan) // start read loop after authentication!
	go c.pingLoop()
	return c, nil
}

func (s *Websocket) connection(public bool) *wsConn {
	if public {
		return &s.public
	}
	return &s.private
}

func (c *wsConn) authenticate(key, secret string) error {
	expires := time.Now().UnixNano()*1000 + 1000
	req := fmt.Sprintf("GET/realtime%d", expires)
	sig := hmac.New(sha256.New, []byte(secret))
	_, _ = sig.Write([]byte(req))
	signature := hex.EncodeToString(sig.Sum(nil))
	c.log.Info("Authenticating")
	err := c.write(&WebsocketRequest{
		Op: "auth",
		Args: []interface{}{
			key,
			expires,
			signature,
		},
	})
	if err != nil {
		return err
	}
	// Next response must be authentication success
	res, err := c.read()
	if err != nil {
		return err
	}
	if !res.IsStatusResponse() {
		return fmt.Errorf("expected status response after authenticating, got: %s", res.Parse.Raw)
	}
	s, err := res.StatusResponse()
	if err != nil {
		return err
	}
	if !s.Success {
		return fmt.Errorf("failed authentication: %s", s.RetMsg)
	}
	return nil
}

func (c *wsConn) write(req *WebsocketRequest) error {
	c.log.Info("Sending request", "op", req.Op, "args", fmt.Sprintf("%v", req.Args))
	return c.WriteJSON(req)
}

// read until connection closed
func (c *wsConn) readLoop(resChan chan<- *WebsocketResponse) {
	for {
		res, err := c.read()
		if err != nil {
			if errors.Is(err, recws.ErrNotConnected) {
				return
			}
			c.log.Error(err, "Error reading next message")
			return
		}
		resChan <- res
	}
}

func (c *wsConn) read() (*WebsocketResponse, error) {
	_, msg, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}
	return &WebsocketResponse{Parse: gjson.ParseBytes(msg)}, nil
}

func (c *wsConn) pingLoop() {
	t := time.NewTimer(time.Second * 30)
	defer t.Stop()
	for {
		<-t.C
		if !c.IsConnected() {
			return
		}
		if err := c.write(&WebsocketRequest{Op: "ping"}); err != nil {
			c.log.Error(err, "Error sending ping")
			return
		}
	}
}

func (s *Websocket) handleLoop() {
	for res := range s.resChan {
		s.handlerFn(res)
	}
}
