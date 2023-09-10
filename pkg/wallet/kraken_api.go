package wallet

import (
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
)

// KrakenResponse is the response from Kraken API
type KrakenResponse struct {
	Error  []string `json:"error"`
	Result struct {
		SOLEUR struct {
			A []string `json:"a"`
			B []string `json:"b"`
			C []string `json:"c"`
			V []string `json:"v"`
			P []string `json:"p"`
			T []int    `json:"t"`
			L []string `json:"l"`
			H []string `json:"h"`
			O string   `json:"o"`
		} `json:"SOLEUR"`
	} `json:"result"`
}

// fetchSOLEURRate fetches the current SOLEUR rate from Kraken API
func fetchSOLEURRate() (decimal.Decimal, error) {
	resp, err := http.Get("https://api.kraken.com/0/public/Ticker?pair=SOLEUR")
	if err != nil {
		return decimal.NewFromFloat(0), err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return decimal.NewFromFloat(0), err
	}

	var krakenResponse KrakenResponse
	err = json.Unmarshal(body, &krakenResponse)
	if err != nil {
		return decimal.NewFromFloat(0), err
	}

	if len(krakenResponse.Result.SOLEUR.P) < 2 {
		return decimal.NewFromFloat(0), errors.New("unexpected data structure from API")
	}

	rateStr := krakenResponse.Result.SOLEUR.P[1]
	rate, err := decimal.NewFromString(rateStr)
	if err != nil {
		return decimal.NewFromFloat(0), err
	}

	return rate, nil
}
