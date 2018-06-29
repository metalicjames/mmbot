/* mmbot - a constant interval market maker bot
Copyright (C) 2018  James Lovejoy

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>. */

package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

// var API = "https://www.bittrex.com/api/v1.1"
// use this when you want bittrex

type Bittrex struct {
	apikey string
	secret string
	client *http.Client
}

type BtGetBalanceRet struct {
	Available float64
	Currency  string
	Balance   float64
}

type BtGetTickerResp struct {
	Bid  float64
	Ask  float64
	Last float64
}

type BtGetOrdersResp struct {
	OrderUuid string
}

type BtMarket struct {
	MarketCurrencyLong string
	BaseCurrencyLong   string
}

type BtMarketSummaries struct {
	High           float64
	Low            float64
	Volume         float64
	Last           float64
	BaseVolume     float64
	Bid            float64
	Ask            float64
	OpenBuyOrders  float64
	OpenSellOrders float64
	PrevDay        float64
}

type BtQuantityRate struct {
	Quantity float64
	Rate     float64
}

type BtOrderBooks struct {
	Buy  BtQuantityRate
	Sell BtQuantityRate
}

func BittrexConnect(apikey string, secret string) *Bittrex {
	return &Bittrex{apikey, secret, &http.Client{}}
}

// Public APIs /public/blah

// public/getticker
func (bx *Bittrex) GetTicker(market string) (Ticker, error) {
	url := API + "/public/getticker?market=" + market
	m, err := bx.sendRecv(url)
	if err != nil {
		return Ticker{}, err
	}

	if m["success"] != true {
		return Ticker{}, errors.New(m["message"].(string))
	}

	var ticker BtGetTickerResp
	mapstructure.Decode(m["result"], &ticker)

	var ret Ticker
	ret.Ask = ticker.Ask
	ret.Bid = ticker.Bid
	ret.Last = ticker.Last
	return ret, nil
}

// public/BtMarkets
func (bx *Bittrex) BtMarkets(market string) (BtMarket, error) {
	url := API + "/public/getticker?market=" + market
	m, err := bx.sendRecv(url)
	if err != nil {
		return BtMarket{}, err
	}

	if m["success"] != true {
		return BtMarket{}, errors.New(m["message"].(string))
	}
	ret := BtMarket{"0", "0"}
	mapstructure.Decode(m["result"], &ret)

	return ret, nil
}

// public/BtMarketsummaries
func (bx *Bittrex) BtMarketSummaries() (BtMarketSummaries, error) {
	url := API + "/public/BtMarketsummaries"

	m, err := bx.sendRecv(url)
	if err != nil {
		return BtMarketSummaries{}, err
	}

	if m["success"] != true {
		return BtMarketSummaries{}, errors.New(m["message"].(string))
	}

	resp := BtMarketSummaries{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	mapstructure.Decode(m["result"], &resp)
	var ret BtMarketSummaries

	ret.PrevDay = resp.PrevDay
	ret.OpenSellOrders = resp.OpenSellOrders
	ret.OpenBuyOrders = resp.OpenBuyOrders
	ret.Ask = resp.Ask
	ret.Bid = resp.Bid
	ret.BaseVolume = resp.BaseVolume
	ret.Last = resp.Last
	ret.Volume = resp.Volume
	ret.High = resp.High
	ret.Low = resp.Low

	return ret, nil
}

// public/getorderbook
func (bx *Bittrex) GetOrderBook(market string) (BtOrderBooks, error) {
	url := API + "/public/getorderbook?market=" + market + "&type=both"

	m, err := bx.sendRecv(url)
	if err != nil {
		return BtOrderBooks{}, err
	}

	if m["success"] != true {
		return BtOrderBooks{}, errors.New(m["message"].(string))
	}

	resp := BtOrderBooks{}
	mapstructure.Decode(m["result"], &resp)

	return resp, nil
}

// Market APIs /market/blah

// market/cancel
func (bx *Bittrex) CancelOrder(UID string) error {
	url := API + "/market/cancel?apikey=" + bx.apikey + "&nonce=1&uuid=" + UID

	m, err := bx.sendRecv(url)
	if err != nil {
		return err
	}

	if m["success"] != true {
		return errors.New(m["message"].(string))
	}

	return nil
}

// market/getopenorders
func (bx *Bittrex) GetOrders(market string) ([]string, error) {
	url := API + "/market/getopenorders?apikey=" + bx.apikey + "&market=" + market

	m, err := bx.sendRecv(url)
	if err != nil {
		return nil, err
	}

	if m["success"] != true {
		return nil, errors.New(m["message"].(string))
	}

	var res []BtGetOrdersResp
	mapstructure.Decode(m["result"], &res)

	var ret []string
	for _, v := range res {
		ret = append(ret, v.OrderUuid)
	}

	return ret, nil
}

// market/buylimit and market/selllimit
func (bx *Bittrex) PlaceOrder(buy bool, market string,
	quantity float64, rate float64) (string, error) {
	url := API + "/market/"
	if buy {
		url += "buylimit"
	} else {
		url += "selllimit"
	}

	url += "?apikey=" + bx.apikey + "&market=" + market + "&quantity=" + strconv.FormatFloat(quantity, 'f', -1, 64) + "&rate=" + strconv.FormatFloat(rate, 'f', -1, 64)

	m, err := bx.sendRecv(url)
	if err != nil {
		return "", err
	}

	if m["success"] != true {
		return "", errors.New(m["message"].(string))
	}

	return m["result"].(map[string]interface{})["uuid"].(string), nil
}

// Account APIs

// account/getbalance
func (bx *Bittrex) GetBalance(asset string) (float64, error) {
	url := API + "/account/getbalances?apikey=" + bx.apikey

	m, err := bx.sendRecv(url)
	if err != nil {
		return 0, err
	}

	if m["success"] != true {
		return 0, errors.New(m["message"].(string))
	}

	var res []BtGetBalanceRet
	mapstructure.Decode(m["result"], &res)

	for _, cur := range res {
		if cur.Currency == asset {
			return cur.Available, nil // parse to float
		}
	}

	return 0, nil
}

func (bx *Bittrex) sendRecv(url string) (map[string]interface{}, error) {
	log.Printf("Req: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("apisign", hmacSign(url, bx.secret))

	resp, err := bx.client.Do(req)
	if err != nil {
		log.Printf("Req err: %v", err)
		return nil, err
	}

	defer resp.Body.Close()

	var m map[string]interface{}

	var r interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&r)
	if err != nil {
		log.Printf("Resp err: %v", err)
		return nil, err
	}

	m = r.(map[string]interface{})

	return m, nil
}
