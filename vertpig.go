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
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

const API = "https://www.vertpig.com/api/v1.1"

type Vertpig struct {
	apikey string
	secret string
	client *http.Client
}

func VertpigConnect(apikey string, secret string) *Vertpig {
	return &Vertpig{apikey, secret, &http.Client{}}
}

type GetTickerResp struct {
	Bid  string
	Ask  string
	Last string
}

func (vp *Vertpig) GetTicker(market string) (Ticker, error) {
	m, err := vp.sendRecv(API + "/public/getticker?market=" + market)
	if err != nil {
		return Ticker{}, err
	}

	if m["success"] != true {
		return Ticker{}, errors.New(m["message"].(string))
	}

	var ticker GetTickerResp
	mapstructure.Decode(m["result"], &ticker)

	var ret Ticker

	ret.Ask, err = strconv.ParseFloat(ticker.Ask, 64)
	if err != nil {
		return Ticker{}, err
	}

	ret.Bid, err = strconv.ParseFloat(ticker.Bid, 64)
	if err != nil {
		return Ticker{}, err
	}

	ret.Last, err = strconv.ParseFloat(ticker.Last, 64)
	if err != nil {
		return Ticker{}, err
	}

	return ret, nil
}

func (vp *Vertpig) CancelOrder(UID string) error {
	url := API + "/market/cancel?apikey=" + vp.apikey + "&nonce=1&uuid=" + UID

	m, err := vp.sendRecv(url)
	if err != nil {
		return err
	}

	if m["success"] != true {
		return errors.New(m["message"].(string))
	}

	return nil
}

func (vp *Vertpig) CancelAll() error {
	url := API + "/market/cancelall?apikey=" + vp.apikey + "&nonce=1"

	_, err := vp.sendRecv(url)
	if err != nil {
		return err
	}

	return nil
}

type GetBalanceRet struct {
	Available string
	Reserved  string
	Currency  string
	Balance   string
}

func (vp *Vertpig) GetBalance(asset string) (float64, error) {
	url := API + "/account/getbalances?apikey=" + vp.apikey + "&nonce=1"

	m, err := vp.sendRecv(url)
	if err != nil {
		return 0, err
	}

	if m["success"] != true {
		return 0, errors.New(m["message"].(string))
	}

	var res []GetBalanceRet
	mapstructure.Decode(m["result"], &res)

	for _, cur := range res {
		if cur.Currency == asset {
			fp, err := strconv.ParseFloat(cur.Available, 64)
			if err != nil {
				return 0, err
			}
			return fp, nil
		}
	}

	return 0, nil
}

type GetOrdersResp struct {
	OrderUUID string
}

func (vp *Vertpig) GetOrders(market string) ([]string, error) {
	url := API + "/market/getopenorders?apikey=" + vp.apikey + "&nonce=1&market=" + market

	m, err := vp.sendRecv(url)
	if err != nil {
		return nil, err
	}

	if m["success"] != true {
		return nil, errors.New(m["message"].(string))
	}

	var res []GetOrdersResp
	mapstructure.Decode(m["result"], &res)

	var ret []string
	for _, v := range res {
		ret = append(ret, v.OrderUUID)
	}

	return ret, nil
}

func (vp *Vertpig) PlaceOrder(buy bool, market string, quantity float64, rate float64) (string, error) {
	url := API + "/market/"
	if buy {
		url += "buylimit"
	} else {
		url += "selllimit"
	}

	url += "?apikey=" + vp.apikey + "&nonce=1&postonly=1&market=" + market + "&quantity=" + strconv.FormatFloat(quantity, 'f', -1, 64) + "&rate=" + strconv.FormatFloat(rate, 'f', -1, 64)

	m, err := vp.sendRecv(url)
	if err != nil {
		return "", err
	}

	if m["success"] != true {
		return "", errors.New(m["message"].(string))
	}

	return m["result"].(map[string]interface{})["uuid"].(string), nil
}

func hmacSign(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha512.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func (vp *Vertpig) sendRecv(url string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("apisign", hmacSign(url, vp.secret))

	resp, err := vp.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var m map[string]interface{}

	var r interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&r)
	if err != nil {
		return nil, err
	}

	m = r.(map[string]interface{})

	return m, nil
}
