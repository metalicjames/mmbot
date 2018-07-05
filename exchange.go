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

// Ticker returns the current price ticker for market of the asset
// priced in the currency. The Bid is the highest buy price and the ask
// is the lowest sell price. The Last is the price of the last trade executed.
type Ticker struct {
	Bid  float64
	Ask  float64
	Last float64
}

// Exchange is an interface that implements a generic
// exchange orderbook.
type Exchange interface {

	// PlaceOrder places a new order in the market. It returns the UID of the
	// newly placed order or an error. PlaceOrder should not allow the placement
	// of an order that would cause a trade (limit or post-only). If a given order
	// would cause a trade PlaceOrder should return a POST_ONLY_FAILED error. UIDs
	// should be unique to every order.
	PlaceOrder(buy bool, market string, quantity float64, rate float64) (string, error)

	// GetOrders should return a slice of the UIDs of the orders currently in the
	// market or an error.
	GetOrders(market string) ([]string, error)

	// CalcelOrder should cancel the given order or return an error
	CancelOrder(UID string) error

	// GetTicker gets the ticker (as defined above) for the given market or
	// returns an error.
	GetTicker(market string) (Ticker, error)

	// GetBalance returns the available balance (funds that are not reserved for
	// existing orders) for the given asset or returns an error.
	GetBalance(asset string) (float64, error)
}
