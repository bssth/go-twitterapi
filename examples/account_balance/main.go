// Print remaining twitterapi.io credits. Useful as a CI guard.
//
//	go run ./examples/account_balance
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bssth/go-twitterapi"
)

func main() {
	c, err := twitterapi.New(twitterapi.Options{})
	if err != nil {
		log.Fatal(err)
	}
	info, err := c.Account.Info(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("recharge_credits=%.4f bonus=%.4f total=%.4f\n",
		info.RechargeCredits, info.TotalBonusCredits, info.RechargeCredits+info.TotalBonusCredits)
}
