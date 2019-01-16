package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: mybot slack-bot-token\n")
		os.Exit(1)
	}

	// start a websocket-based Real Time API session
	ws, id := slackConnect(os.Args[1])
	fmt.Println("mybot ready, ^C exits")

	for {
		// read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			log.Fatal(err)
		}

		// see if we're mentioned
		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
			// if so try to parse if
			parts := strings.Fields(m.Text)
			if len(parts) == 3 && parts[1] == "stock" {
				// looks good, get the quote and reply with the result
				go func(m Message) {
					m.Text = getQuote(parts[2])
					postMessage(ws, m)
				}(m)
				// NOTE: the Message object is copied, this is intentional
			} else {
				// huh?
				m.Text = fmt.Sprintf("sorry, that does not compute\n")
				postMessage(ws, m)
			}
		}
	}
}

// Get the quote from iextrading
func getQuote(sym string) string {
	url := fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/quote", sym)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return fmt.Sprintf("error: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	var respObj stockPrice
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return fmt.Sprintf("%s (%s) is trading at %f", respObj.CompanyName, respObj.Symbol, respObj.LatestPrice)
}

type stockPrice struct {
	LatestPrice float64 `json:"latestPrice"`
	CompanyName string  `json:"companyName"`
	Symbol      string  `json:"symbol"`
}
