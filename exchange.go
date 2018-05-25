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

type Ticker struct {
	Bid  float64
	Ask  float64
	Last float64
}

type Exchange interface {
	PlaceOrder(buy bool, market string, quantity float64, rate float64) (string, error)
	GetOrders(market string) ([]string, error)
	CancelOrder(UID string) error
	GetTicker(market string) (Ticker, error)
	GetBalance(asset string) (float64, error)
}
