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
	"fmt"
	"log"
)

type Order struct {
	UID      string
	Buy      bool
	Quantity float64
	Rate     float64
	Filled   bool
	Middle   bool
}

type Book struct {
	Orders   []Order
	Market   string
	High     float64
	Low      float64
	Start    float64
	Interval float64
	Ex       Exchange
	FirstRun bool
}

func NewBook(market string, high float64, low float64, start float64, interval float64, quantity float64, exchange Exchange) *Book {
	b := &Book{nil, market, high, low, start, interval, exchange, true}

	err := LoadStruct("./vpbook"+market, b)
	if err != nil {
		log.Printf("%v", err)

		midFound := false
		for i := high; i >= low; i -= interval * start {
			if !midFound {
				if i <= start {
					b.Orders = append(b.Orders, Order{"", false, quantity / i, i, false, true})
					midFound = true
				} else {
					b.Orders = append(b.Orders, Order{"", false, quantity / i, i, false, false})
				}
			} else {
				b.Orders = append(b.Orders, Order{"", true, quantity / i, i, false, false})
			}
		}
	}

	var currency, asset float64
	for _, order := range b.Orders {
		if !order.Middle {
			if order.Buy {
				currency += order.Quantity * order.Rate
			} else {
				asset += order.Quantity
			}
		}
	}

	b.Ex = exchange

	log.Printf("Market: %s, Currency: %f, Asset: %f, # orders: %d", market, currency, asset, len(b.Orders))

	return b
}

func (b *Book) Tick() error {
	// Update order statuses (filled)

	defer func() {
		err := SaveStruct("./vpbook"+b.Market, b)
		if err != nil {
			log.Printf("%v", err)
		}
	}()

	open, err := b.Ex.GetOrders(b.Market)
	if err != nil {
		return err
	}

	openOrders := map[string]bool{}
	for _, uid := range open {
		openOrders[uid] = true
	}

	filledOne := false
	for i, order := range b.Orders {
		if _, ok := openOrders[order.UID]; !ok && !order.Middle {
			b.Orders[i].Filled = true
			filledOne = true
		}
	}

	if filledOne && !b.FirstRun {
		// Get price
		ticker, err := b.Ex.GetTicker(b.Market)
		if err != nil {
			return err
		}

		price := (ticker.Ask + ticker.Bid) / 2

		// Set new midpoint
		log.Printf("Current price: %f", price)

		orig := 0
		for i, order := range b.Orders {
			if order.Middle {
				orig = i
				b.Orders[i].Middle = false
				break
			}
		}

		log.Printf("Old middle: %v", orig)

		middle := orig
		for i, order := range b.Orders {
			run := false
			if order.Filled && i < orig {
				run = true
			} else if i+1 < len(b.Orders) {
				if !b.Orders[i+1].Filled && order.Filled {
					run = true
				}
			}

			if run {
				middle = i
				break
			}
		}

		b.Orders[middle].Middle = true

		log.Printf("New middle: %d", middle)

		for i, order := range b.Orders {
			if order.Filled {
				if i <= middle {
					b.Orders[i].Buy = false
				} else {
					b.Orders[i].Buy = true
				}
			}
			//log.Printf("%+v", b.Orders[i])
		}
	}

	// Check we have enough balance for the orders we want to place

	var reqAsset, reqCurrency float64
	for _, order := range b.Orders {
		if order.Filled && !order.Middle {
			if order.Buy {
				reqCurrency += order.Quantity * order.Rate
			} else {
				reqAsset += order.Quantity
			}
		}
	}

	asset := b.Market[:3]
	currency := b.Market[3:]
	assetBal, err := b.Ex.GetBalance(asset)
	if err != nil {
		return err
	}

	if assetBal < reqAsset {
		return fmt.Errorf("Not enough asset to place orders. Wanted: %s%f, Have: %s%f", asset, reqAsset, asset, assetBal)
	}

	currencyBal, err := b.Ex.GetBalance(currency)
	if err != nil {
		return err
	}

	if currencyBal < reqCurrency {
		return fmt.Errorf("Not enough currency to place orders. Wanted: %s%f, Have: %s%f", currency, reqCurrency, currency, currencyBal)
	}

	// re-submit filled orders
	for i, order := range b.Orders {
		if order.Filled && !order.Middle {
			uid, err := b.Ex.PlaceOrder(order.Buy, b.Market, order.Quantity, order.Rate)
			if err != nil {
				log.Printf("%+v", err)
				continue
			}

			b.Orders[i].Filled = false
			b.Orders[i].UID = uid

			log.Printf("Placed Order: %+v", b.Orders[i])
		}
	}

	b.FirstRun = false

	return nil
}
