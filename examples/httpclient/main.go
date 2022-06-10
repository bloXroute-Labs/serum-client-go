package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/bloXroute-Labs/serum-api/bxserum/provider"
	api "github.com/bloXroute-Labs/serum-api/proto"
	log "github.com/sirupsen/logrus"
)

func main() {
	callOrderbookHTTP()
	callOpenOrdersHTTP()
	callTradesHTTP()
	callTickersHTTP()
	callUnsettledHTTP()

	ownerAddr, ok := os.LookupEnv("PUBLIC_KEY")
	if !ok {
		log.Infof("PUBLIC_KEY environment variable not set")
		log.Infof("Skipping Place and Cancel Order examples")
		return
	}

	ooAddr, _ := os.LookupEnv("OPEN_ORDERS")

	clientOrderID := callPlaceOrderHTTP(ownerAddr, ooAddr)
	callCancelByClientOrderIDHTTP(ownerAddr, ooAddr, clientOrderID)
}

func callOrderbookHTTP() {
	h, _ := provider.NewHTTPClient()

	// Unary response
	orderbook, err := h.GetOrderbook("ETH-USDT", 0)
	if err != nil {
		log.Errorf("error with GetOrderbook request for ETH-USDT: %v", err)
	} else {
		fmt.Println(orderbook)
	}

	fmt.Println()

	orderbook, err = h.GetOrderbook("SOLUSDT", 2)
	if err != nil {
		log.Errorf("error with GetOrderbook request for SOLUSDT: %v", err)
	} else {
		fmt.Println(orderbook)
	}

	fmt.Println()

	orderbook, err = h.GetOrderbook("SOL:USDC", 3)
	if err != nil {
		log.Errorf("error with GetOrderbook request for SOL:USDC: %v", err)
	} else {
		fmt.Println(orderbook)
	}
}

func callOpenOrdersHTTP() {
	client := &http.Client{Timeout: time.Second * 60}
	opts, err := provider.DefaultRPCOpts(provider.MainnetSerumAPIHTTP)
	h := provider.NewHTTPClientWithOpts(client, opts)

	orders, err := h.GetOpenOrders("SOLUSDT", "HxFLKUAmAMLz1jtT3hbvCMELwH5H9tpM2QugP8sKyfhc")
	if err != nil {
		log.Errorf("error with GetOrders request for SOLUSDT: %v", err)
	} else {
		fmt.Println(orders)
	}

	fmt.Println()
}

func callUnsettledHTTP() {
	client := &http.Client{Timeout: time.Second * 60}
	opts, err := provider.DefaultRPCOpts(provider.MainnetSerumAPIHTTP)
	h := provider.NewHTTPClientWithOpts(client, opts)

	response, err := h.GetUnsettled("SOLUSDT", "HxFLKUAmAMLz1jtT3hbvCMELwH5H9tpM2QugP8sKyfhc")
	if err != nil {
		log.Errorf("error with GetOrders request for SOLUSDT: %v", err)
	} else {
		fmt.Println(response)
	}

	fmt.Println()
}

func callTradesHTTP() {
	h, _ := provider.NewHTTPClient()

	trades, err := h.GetTrades("SOLUSDT", 5)
	if err != nil {
		log.Errorf("error with GetTrades request for SOLUSDT: %v", err)
	} else {
		fmt.Println(trades)
	}

	fmt.Println()
}

func callTickersHTTP() {
	h, _ := provider.NewHTTPClient()

	tickers, err := h.GetTickers("SOLUSDT")
	if err != nil {
		log.Errorf("error with GetTickers request for SOLUSDT: %v", err)
	} else {
		fmt.Println(tickers)
	}

	fmt.Println()
}

const (
	// SOL/USDC market
	marketAddr = "9wFFyRfZBsuAha4YcuxcXLKwMxJR43S7fPfQLusDBzvT"

	orderSide   = api.Side_S_ASK
	orderType   = api.OrderType_OT_LIMIT
	orderPrice  = float64(170200)
	orderAmount = float64(0.1)
)

func callPlaceOrderHTTP(ownerAddr, ooAddr string) uint64 {
	client := &http.Client{Timeout: time.Second * 30}
	rpcOpts, err := provider.DefaultRPCOpts(provider.MainnetSerumAPIHTTP)
	h := provider.NewHTTPClientWithOpts(client, rpcOpts)

	// generate a random clientOrderId for this order
	rand.Seed(time.Now().UnixNano())
	clientOrderID := rand.Uint64()

	opts := provider.PostOrderOpts{
		ClientOrderID:     clientOrderID,
		OpenOrdersAddress: ooAddr,
	}

	sig, err := h.SubmitOrder(ownerAddr, ownerAddr, marketAddr,
		orderSide, []api.OrderType{orderType}, orderAmount,
		orderPrice, opts)
	if err != nil {
		log.Fatalf("failed to submit order (%v)", err)
	}

	fmt.Printf("placed order %v with clientOrderID %v\n", sig, clientOrderID)

	return clientOrderID
}

func callCancelByClientOrderIDHTTP(ownerAddr, ooAddr string, clientOrderID uint64) {
	time.Sleep(60 * time.Second)
	client := &http.Client{Timeout: time.Second * 30}
	opts, err := provider.DefaultRPCOpts(provider.MainnetSerumAPIHTTP)
	h := provider.NewHTTPClientWithOpts(client, opts)

	_, err = h.SubmitCancelByClientOrderID(clientOrderID, ownerAddr,
		marketAddr, ooAddr, true)
	if err != nil {
		log.Fatalf("failed to cancel order by client ID (%v)", err)
	}

	fmt.Printf("canceled order for clientOrderID %v\n", clientOrderID)
}
