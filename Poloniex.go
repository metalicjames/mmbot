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
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

const Poloniex_API = "https://poloniex.com"

type Poloniex struct {
	key    string
	secret []byte
	client *http.Client
}

func PoloniexConnect(apiKey string, secret []byte) *Poloniex {
	return &Poloniex{apiKey, secret, &http.Client{}}
}

type GetPoloniexTickerResp struct {
	HighestBid string
	LowestAsk  string
	Last       string
}

func (plo *Poloniex) Name() string {
	return "poloniex"
}

func (polo *Poloniex) GetTicker(market string) (Ticker, error) {
	m_, err := polo.sendGetRecv(Poloniex_API + "/public?command=returnTicker")
	m := m_.(map[string]interface{})

	if err != nil {
		return Ticker{}, err
	}

	ticker := GetPoloniexTickerResp{}
	mapstructure.Decode(m[market], &ticker)

	var ret Ticker
	ret.Ask, err = strconv.ParseFloat(ticker.LowestAsk, 64)
	if err != nil {
		return Ticker{}, err
	}

	ret.Bid, err = strconv.ParseFloat(ticker.HighestBid, 64)
	if err != nil {
		return Ticker{}, err
	}

	ret.Last, err = strconv.ParseFloat(ticker.Last, 64)
	if err != nil {
		return Ticker{}, err
	}
	return ret, nil
}

func (polo *Poloniex) CancelOrder(orderNumber string) error {
	apiURL := Poloniex_API + "/tradingApi"

	data := url.Values{}
	data.Add("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	data.Add("orderNumber", orderNumber)
	data.Add("command", "cancelOrder")

	m_, err := polo.sendPostRecv(apiURL, data.Encode())
	if err != nil {
		return err
	}

	var m (map[string]string)
	mapstructure.Decode(m_, &m)

	if m["error"] != "" {
		return errors.New(m["error"])
	}

	return nil
}

func (polo *Poloniex) GetBalance(asset string) (float64, error) {
	apiURL := Poloniex_API + "/tradingApi"

	data := url.Values{}
	data.Add("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	data.Add("command", "returnBalances")

	m_, err := polo.sendPostRecv(apiURL, data.Encode())
	if err != nil {
		return 0, err
	}

	var m (map[string]string)
	mapstructure.Decode(m_, &m)

	if m["error"] != "" {
		return 0, errors.New(m["error"])
	}

	for cur, val := range m {
		if cur == asset {
			fp, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return 0, err
			}
			return fp, nil
		}
	}

	return 0, nil
}

type GetPoloniexOrdersResp struct {
	OrderNumber string
}

func (polo *Poloniex) GetOrders(currencyPair string) ([]string, error) {
	apiURL := Poloniex_API + "/tradingApi"

	data := url.Values{}
	data.Add("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	data.Add("currencyPair", currencyPair)
	data.Add("command", "returnOpenOrders")

	m_, err := polo.sendPostRecv(apiURL, data.Encode())
	if err != nil {
		return nil, err
	}

	var m []GetPoloniexOrdersResp
	mapstructure.Decode(m_, &m)

	var ret []string
	for _, v := range m {
		ret = append(ret, v.OrderNumber)
	}

	return ret, nil
}

func (polo *Poloniex) PlaceOrder(buy bool, currencyPair string, amount float64, rate float64) (string, error) {
	apiURL := Poloniex_API + "/tradingApi"

	data := url.Values{}
	data.Add("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	data.Add("currencyPair", currencyPair)
	data.Add("amount", fmt.Sprintf("%f", amount))
	data.Add("rate", fmt.Sprintf("%f", rate))
	if buy {
		data.Add("command", "buy")
		data.Add("postOnly", "1")
	} else {
		data.Add("command", "sell")
	}

	m_, err := polo.sendPostRecv(apiURL, data.Encode())
	if err != nil {
		return "", err
	}
	var m (map[string]string)
	mapstructure.Decode(m_, &m)
	log.Println("placed order", currencyPair)
	if m["error"] != "" {
		return "", errors.New(m["error"])
	}

	var res GetPoloniexOrdersResp
	mapstructure.Decode(m, &res)

	return res.OrderNumber, nil
}

func (polo *Poloniex) sendGetRecv(url string) (interface{}, error) {
	req, _ := http.NewRequest("GET", url, nil)
	return polo.processRequest(req)
}

func (polo *Poloniex) sendPostRecv(url string, data string) (interface{}, error) {
	req, _ := http.NewRequest("POST", url, strings.NewReader(data))
	req.Header.Add("Key", polo.key)
	req.Header.Add("Sign", hmacSign([]byte(data), polo.secret))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return polo.processRequest(req)
}

func (polo *Poloniex) processRequest(req *http.Request) (interface{}, error) {
	req.Close = true
	req.Header.Add("Accept", "application/json")

	resp, err := polo.client.Do(req)
	if err != nil {
		log.Printf("Req err: %v", err)
		return nil, err
	}

	defer resp.Body.Close()

	var r interface{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&r)
	if err != nil {
		log.Printf("Resp err: %v", err)
		return nil, err
	}

	return r, nil
}
