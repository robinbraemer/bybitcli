package linear

//
//import (
//	"crypto/hmac"
//	"crypto/sha256"
//	"encoding/hex"
//	"encoding/json"
//	"fmt"
//	"github.com/chuckpreslar/emission"
//	"github.com/gorilla/websocket"
//	bot "github.com/robinbraemer/bybitcli"
//	linear2 "github.com/robinbraemer/bybitcli/exchange/bybit/linear"
//	"github.com/robinbraemer/bybitcli/util"
//	"github.com/tidwall/gjson"
//	"strconv"
//	"strings"
//	"sync"
//	"time"
//)
//
//type ws struct {
//	log             util.Logger
//	addr            string
//	emitter         *emission.Emitter
//	localOrderBooks map[string]*LocalOrderBook // by symbol
//	conn            *recws.RecConn
//
//	subscribeCmds       []*WSCmd
//	queuedSubscribeCmds []*WSCmd // sent when ws is started
//}
//
//// StartWebSocket a new websocket connection, blocks and stops until context is canceled.
//func StartWebSocket(ctx context.Context) error {
//	ws := FromContext(ctx).ws
//	ws.connect()
//
//	for _, cmd := range ws.queuedSubscribeCmds {
//		if err := ws.sendCmd(cmd); err != nil {
//			return fmt.Errorf("erorr sending queued subscribtion: %w", err)
//		}
//	}
//	ws.queuedSubscribeCmds = nil
//
//	go func() {
//		t := time.NewTicker(time.Second * 15)
//		defer t.Stop()
//		for {
//			select {
//			case <-t.C:
//				if err := ws.ping(); err != nil {
//					ws.log.Error(err, "Error sending ping")
//				}
//			case <-ctx.Done():
//				ws.close()
//				return
//			}
//		}
//	}()
//
//	for {
//		messageType, data, err := ws.conn.ReadMessage()
//		if err != nil {
//			if errors.Is(err, recws.ErrNotConnected) {
//				return nil // done
//			}
//			ws.log.Error(err, "Error reading ws message")
//			time.Sleep(time.Millisecond * 100)
//			continue
//		}
//		if err = ws.processMessage(messageType, data); err != nil {
//			ws.log.Error(err, "Error processing ws message")
//		}
//	}
//}
//
//func (ws *ws) connect() { ws.conn.Dial(ws.addr, nil) }
//func (ws *ws) close()   { ws.conn.Close() }
//
//func Subscribe(ctx context.Context, arg string) error {
//	ws := FromContext(ctx).ws
//	cmd := &WSCmd{
//		Op:   "subscribe",
//		Args: []interface{}{arg},
//	}
//	if ws.isConnected() {
//		if err := ws.sendCmd(cmd); err != nil {
//			return err
//		}
//	} else {
//		ws.queuedSubscribeCmds = append(ws.queuedSubscribeCmds, cmd)
//	}
//	ws.subscribeCmds = append(ws.subscribeCmds, cmd)
//	return nil
//}
//
//// ws
//func (c *bybit.client) subscribeHandler() error {
//	if c.apiKey != "" && c.secretKey != "" {
//		if err := c.authWS(); err != nil {
//			return fmt.Errorf("error authenticating: %w", err)
//		}
//	}
//	for _, cmd := range c.subscribeCmds {
//		if err := c.sendCmd(cmd); err != nil {
//			return fmt.Errorf("error sending command: %w", err)
//		}
//	}
//	return nil
//}
//
//type WSCmd struct {
//	Op   string        `json:"op"`
//	Args []interface{} `json:"args"`
//}
//
//func (c *bybit.client) authWS() error {
//	expires := time.Now().Unix()*1000 + 10000
//	req := fmt.Sprintf("GET/realtime%d", expires)
//	sig := hmac.New(sha256.New, []byte(c.secretKey))
//	sig.Write([]byte(req))
//	signature := hex.EncodeToString(sig.Sum(nil))
//	return c.sendCmd(&WSCmd{
//		Op: "auth",
//		Args: []interface{}{
//			c.apiKey,
//			//fmt.Sprintf("%v", expires),
//			expires,
//			signature,
//		},
//	})
//}
//
//func (ws *ws) sendCmd(cmd *WSCmd) error {
//	data, err := json.Marshal(cmd)
//	if err != nil {
//		return err
//	}
//	ws.log.Info("Sending command", "op", cmd.Op, "argsNum", len(cmd.Args))
//	return ws.send(string(data))
//}
//
//func (ws *ws) send(msg string) (err error) {
//	defer func() {
//		if r := recover(); r != nil {
//			err = fmt.Errorf("recovered panic while writing ws message: %v", r)
//		}
//	}()
//	return ws.conn.WriteMessage(websocket.TextMessage, []byte(msg))
//}
//
//func (ws *ws) ping() (err error) {
//	if !ws.isConnected() {
//		return
//	}
//	defer func() {
//		if r := recover(); r != nil {
//			err = fmt.Errorf("panic while sending ping: %v", r)
//		}
//	}()
//	return ws.conn.WriteMessage(websocket.TextMessage, []byte(`{"op":"ping"}`))
//}
//
//func (ws *ws) isConnected() bool { return ws.conn.IsConnected() }
//
//const (
//	WSOrderBook25L1 = "orderBookL2_25" // order_book_25L1.BTCUSD
//	WSKLine         = "kline"          // kline.BTCUSD.1m
//	WSTrade         = "trade"          // trade/trade.BTCUSD
//	WSInsurance     = "insurance"
//	WSInstrument    = "instrument"
//
//	WSPosition  = "position"
//	WSExecution = "execution"
//	WSOrder     = "order"
//
//	WSDisconnected = "disconnected"
//
//	topicOrderBook25l1prefix = string(WSOrderBook25L1 + ".")
//)
//
//type OrderBookL2 struct {
//	Price  bot.Number `json:"price"`
//	Symbol string     `json:"symbol"`
//	ID     int        `json:"id"`
//	Side   string     `json:"side"`
//	Size   int        `json:"size"`
//}
//
//func (b OrderBookL2) Key() string { return strconv.Itoa(b.ID) }
//
//type OrderBookL2Delta struct {
//	Delete []*OrderBookL2
//	Update []*OrderBookL2
//	Insert []*OrderBookL2
//}
//
//func (ws *ws) processMessage(messageType int, data []byte) error {
//	ret := gjson.ParseBytes(data)
//
//	if ret.Get("ret_msg").String() == "pong" {
//		return ws.handlePong()
//	}
//
//	if topicValue := ret.Get("topic"); topicValue.Exists() {
//		topic := topicValue.String()
//		if strings.HasPrefix(topic, topicOrderBook25l1prefix) {
//			symbol := topic[len(topicOrderBook25l1prefix):]
//			type_ := ret.Get("type").String()
//			raw := ret.Get("data").Raw
//
//			switch type_ {
//			case "snapshot":
//				var data []*OrderBookL2
//				err := json.Unmarshal([]byte(raw), &data)
//				if err != nil {
//					return fmt.Errorf("error unmarshal snapshot: %w", err)
//				}
//				ws.processOrderBookSnapshot(symbol, data...)
//			case "delta":
//				var delta OrderBookL2Delta
//				err := json.Unmarshal([]byte(raw), &delta)
//				if err != nil {
//					return fmt.Errorf("error unmarshal delta: %w", err)
//				}
//				ws.processOrderBookDelta(symbol, &delta)
//			}
//			//} else if strings.HasPrefix(topic, WSTrade) {
//			//	symbol := strings.TrimLeft(topic, WSTrade+".")
//			//	raw := ret.Get("data").Raw
//			//	var data []*Trade
//			//	err := json.Unmarshal([]byte(raw), &data)
//			//	if err != nil {
//			//		log.Printf("%v", err)
//			//		return
//			//	}
//			//	ws.processTrade(symbol, data...)
//			//} else if strings.HasPrefix(topic, WSKLine) {
//			//	// kline.BTCUSD.1m
//			//	topicArray := strings.Split(topic, ".")
//			//	if len(topicArray) != 3 {
//			//		return
//			//	}
//			//	symbol := topicArray[1]
//			//	raw := ret.Get("data").Raw
//			//	var data KLine
//			//	err := json.Unmarshal([]byte(raw), &data)
//			//	if err != nil {
//			//		log.Printf("%v", err)
//			//		return
//			//	}
//			//	b.processKLine(symbol, data)
//			//} else if strings.HasPrefix(topic, WSInsurance) {
//			//	// insurance.BTC
//			//	topicArray := strings.Split(topic, ".")
//			//	if len(topicArray) != 2 {
//			//		return
//			//	}
//			//	currency := topicArray[1]
//			//	raw := ret.Get("data").Raw
//			//	var data []*Insurance
//			//	err := json.Unmarshal([]byte(raw), &data)
//			//	if err != nil {
//			//		log.Printf("%v", err)
//			//		return
//			//	}
//			//	b.processInsurance(currency, data...)
//			//} else if strings.HasPrefix(topic, WSInstrument) {
//			//	topicArray := strings.Split(topic, ".")
//			//	if len(topicArray) != 2 {
//			//		return
//			//	}
//			//	symbol := topicArray[1]
//			//	raw := ret.Get("data").Raw
//			//	var data []*Instrument
//			//	err := json.Unmarshal([]byte(raw), &data)
//			//	if err != nil {
//			//		log.Printf("%v", err)
//			//		return
//			//	}
//			//	b.processInstrument(symbol, data...)
//		} else if topic == string(WSPosition) {
//			raw := ret.Get("data").Raw
//			var p []*linear2.Position
//			err := json.Unmarshal([]byte(raw), &p)
//			if err != nil {
//				return fmt.Errorf("error unmarshal %T: %w", p, err)
//			}
//			ws.emitter.Emit(WSPosition, p)
//			//} else if topic == WSExecution {
//			//	raw := ret.Get("data").Raw
//			//	var data []*Execution
//			//	err := json.Unmarshal([]byte(raw), &data)
//			//	if err != nil {
//			//		log.Printf("%v", err)
//			//		return
//			//	}
//			//	b.processExecution(data...)
//			//} else if topic == WSOrder {
//			//	raw := ret.Get("data").Raw
//			//	var data []*Order
//			//	err := json.Unmarshal([]byte(raw), &data)
//			//	if err != nil {
//			//		log.Printf("%v", err)
//			//		return
//			//	}
//			//	b.processOrder(data...)
//			//}
//		}
//	}
//	return nil
//}
//
//func (ws *ws) handlePong() (err error) {
//	defer func() {
//		if r := recover(); r != nil {
//			err = fmt.Errorf("panic while handling pong: %v", r)
//		}
//	}()
//	pongHandler := ws.conn.PongHandler()
//	if pongHandler != nil {
//		return pongHandler("pong")
//	}
//	return nil
//}
//
//func (ws *ws) processOrderBookSnapshot(symbol string, book ...*OrderBookL2) {
//	var value *LocalOrderBook
//	var ok bool
//
//	value, ok = ws.localOrderBooks[symbol]
//	if !ok {
//		value = newLocalOrderBook()
//		ws.localOrderBooks[symbol] = value
//	}
//	value.saveSnapshot(book)
//
//	ws.emitter.Emit(WSOrderBook25L1, symbol, value.OrderBook())
//}
//
//// Emitter returns the websocket event emitter for registering listeners.
//func Emitter(ctx context.Context) *emission.Emitter {
//	return FromContext(ctx).ws.emitter
//}
//
//func (ws *ws) processOrderBookDelta(symbol string, delta *OrderBookL2Delta) {
//	value, ok := ws.localOrderBooks[symbol]
//	if !ok {
//		return
//	}
//	value.Update(delta)
//	ws.emitter.Emit(WSOrderBook25L1, symbol, value.OrderBook())
//}
//
//type LocalOrderBook struct {
//	book map[string]*OrderBookL2
//	mu   sync.RWMutex
//}
//
//func (b *LocalOrderBook) saveSnapshot(newOrderBook []*OrderBookL2) {
//	b.mu.Lock()
//	defer b.mu.Unlock()
//	b.book = make(map[string]*OrderBookL2)
//	for _, v := range newOrderBook {
//		b.book[v.Key()] = v
//	}
//}
//
//type OrderBook struct {
//	Bids      []Item    `json:"bids"`
//	Asks      []Item    `json:"asks"`
//	Timestamp time.Time `json:"timestamp"`
//}
//
//// Item stores the amount and price values
//type Item struct {
//	Amount float64 `json:"amount"`
//	Price  float64 `json:"price"`
//}
//
//func (b *LocalOrderBook) OrderBook() (ob OrderBook) {
//	for _, v := range b.book {
//		switch v.Side {
//		case "Buy":
//			ob.Bids = append(ob.Bids, Item{
//				Price:  float64(v.Price),
//				Amount: float64(v.Size),
//			})
//		case "Sell":
//			ob.Asks = append(ob.Asks, Item{
//				Price:  float64(v.Price),
//				Amount: float64(v.Size),
//			})
//		}
//	}
//	sort.Slice(ob.Bids, func(i, j int) bool {
//		return ob.Bids[i].Price > ob.Bids[j].Price
//	})
//	sort.Slice(ob.Asks, func(i, j int) bool {
//		return ob.Asks[i].Price < ob.Asks[j].Price
//	})
//	ob.Timestamp = time.Now()
//	return
//}
//
//func (b *LocalOrderBook) Update(delta *OrderBookL2Delta) {
//	b.mu.Lock()
//	defer b.mu.Unlock()
//
//	for _, elem := range delta.Delete {
//		delete(b.book, elem.Key())
//	}
//
//	for _, elem := range delta.Update {
//		if v, ok := b.book[elem.Key()]; ok {
//			// price is same while id is same
//			// v.Price = elem.Price
//			v.Size = elem.Size
//			v.Side = elem.Side
//		}
//	}
//
//	for _, elem := range delta.Insert {
//		b.book[elem.Key()] = elem
//	}
//}
//
//func newLocalOrderBook() *LocalOrderBook {
//	return &LocalOrderBook{book: map[string]*OrderBookL2{}}
//}
