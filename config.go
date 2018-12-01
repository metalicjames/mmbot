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

type Config struct {
	Exchange string
	Apikey   string
	Secret   string
	Markets  []Market
}

type Market struct {
	Market   string
	High     float64
	Low      float64
	Start    float64
	Interval float64
	Quantity float64
}

func Load() ([]*Book, error) {
	var conf Config
	err := LoadStruct("./config.json", &conf)
	if err != nil {
		return nil, err
	}

	var exchange Exchange
	switch conf.Exchange {
	case "poloniex":
		exchange = PoloniexConnect(conf.Apikey, []byte(conf.Secret))
	case "vertpig":
		exchange = VertpigConnect(conf.Apikey, []byte(conf.Secret))
	case "bittrex":
		exchange = BittrexConnect(conf.Apikey, conf.Secret)
	}

	var ret []*Book
	for _, m := range conf.Markets {
		ret = append(ret, NewBook(m.Market, m.High, m.Low, m.Start, m.Interval, m.Quantity, exchange))
	}

	return ret, nil
}
