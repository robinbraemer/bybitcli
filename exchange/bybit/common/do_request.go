package common

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/robinbraemer/bybitbot/exchange/bybit"
	"github.com/robinbraemer/bybitbot/util"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

func PublicRequest(
	ctx context.Context, method, apiURL string,
	params util.M, result bybit.Base,
) error {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var p []string
	for _, k := range keys {
		p = append(p, fmt.Sprintf("%v=%v", k, params[k]))
	}

	param := strings.Join(p, "&")
	c := bybit.FromContext(ctx)
	fullURL := endpoint(ctx) + apiURL
	if param != "" {
		fullURL += "?" + param
	}
	return doRequest(ctx, method, fullURL, nil, c.Log.WithName("PublicRequest"), result)
}

func SignedRequest(
	ctx context.Context,
	method, apiURL string,
	params util.M,
	result bybit.Base,
) (err error) {
	c := bybit.FromContext(ctx)
	if err = syncTime(ctx); err != nil {
		return err
	}

	if params == nil {
		params = util.M{}
	}

	params["api_key"] = c.APIKey
	params["timestamp"] = timestamp(c.ServerTimeOffset)

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var p []string
	for _, k := range keys {
		p = append(p, fmt.Sprintf("%s=%v", k, params[k]))
	}

	param := strings.Join(p, "&")
	signature := sign(param, c.SecretKey)

	fullURL := endpoint(ctx) + apiURL
	var reqBody []byte
	if method == http.MethodPost {
		// Post requests send the parameters in the body
		params["sign"] = signature
		reqBody, err = json.Marshal(params) // std Go json sorts parameters alphabetically
		if err != nil {
			return err
		}
	} else {
		// Non-post requests send the parameters in the url query
		param += "&sign=" + signature
		fullURL += "?" + param
	}

	return doRequest(ctx, method, fullURL, reqBody, c.Log.WithName("SignedRequest"), result)
}

func timestamp(serverTimeOffset time.Duration) int64 {
	return time.Now().Add(serverTimeOffset).UnixNano() / 1e6
}

func sign(param, secretKey string) string {
	sig := hmac.New(sha256.New, []byte(secretKey))
	_, _ = sig.Write([]byte(param))
	return hex.EncodeToString(sig.Sum(nil))
}

func doRequest(
	ctx context.Context,
	method string,
	url string,
	reqBody []byte,
	log util.Logger,
	result bybit.Base,
) error {
	c := bybit.FromContext(ctx).Client
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	log = log.WithValues("url", url).WithValues("method", method)
	log.Info("Sending request")

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("error during request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status %s (%d)", res.Status, res.StatusCode)
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	if err = json.Unmarshal(resBody, result); err != nil {
		return fmt.Errorf("error unmarshal result: %w", err)
	}

	log.WithValues("status", res.Status).
		WithValues("code", res.StatusCode).
		WithValues("retCode", result.ReturnCode()).
		WithValues("retMsg", result.ReturnMessage()).
		Info("Received response")

	// update server time offset
	serverTime, err := bybit.ParseTime(result.ServerTime())
	if err != nil {
		log.Error(err, "Error parsing server time")
	} else {
		bybit.UpdateServerTime(ctx, serverTime)
	}

	return nil
}

func syncTime(ctx context.Context) error {
	c := bybit.FromContext(ctx)
	if time.Since(c.LastTimeSync) > time.Hour {
		return SyncTime(ctx)
	}
	return nil
}

func endpoint(ctx context.Context) string {
	if bybit.IsTestnet(ctx) {
		return TestnetEndpoint
	}
	return MainnetEndpoint
}
