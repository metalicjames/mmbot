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
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Printf("mmbot v0.0.3  Copyright (C) 2018  James Lovejoy.")
	log.Printf("This program comes with ABSOLUTELY NO WARRANTY.")
	log.Printf("This is free software, and you are welcome to redistribute it under certain conditions.")
	log.Printf("Read the LICENSE and README for more details.")

	books, err := Load()
	if err != nil {
		log.Printf("%v", err)
		return
	}

	ticker := time.NewTicker(time.Second * 30)
	go func() {
		for range ticker.C {
			for _, b := range books {
				go func(b *Book) {
					if err := b.Tick(); err != nil {
						log.Printf("%v", err)
					}
				}(b)
			}
		}
	}()

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	ticker.Stop()
}
