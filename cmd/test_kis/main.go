package main

import (
	"fmt"
	"log"

	"github.com/Paaaark/hanquant/internal/data"
)

func main() {
    client := data.NewKISClient()

    // Step 1: Get access token
    // token, err := client.GetKISAccessToken()
    // if err != nil {
    //     log.Fatalf("Failed to get access token: %v", err)
    // }

    // fmt.Println("✅ Access Token acquired.", token)

    // Step 2: Fetch daily price for Samsung Electronics (005930)
    body, err := client.GetDailyPrice("005930", "20240530", "20240630", "D")
    if err != nil {
        log.Fatalf("Failed to get daily price: %v", err)
    }

    fmt.Println("✅ Daily price response:")
    fmt.Println(body)

    stockListings, err := data.Load("stock_listings.csv")
    searched := stockListings.SearchStocks("삼성전자")
    fmt.Println(searched)

    top_stocks, err := client.GetTopFluctuationStocks()
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(top_stocks)
}
