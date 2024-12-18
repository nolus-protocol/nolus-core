package types

import (
	"strings"

	"github.com/Nolus-Protocol/nolus-core/app/params"
)

// TODO always use the base-price query for all protocols, example : {"base_price": { "currency": "OSMO"}}  ( base price of a protocol is the lpn ticker)
//  := amount * (float64(QuoteAmountAsInt) / float64(AmountAsInt)) should be sufficient for all cases
// lets say fee is paid in OSMO and we use the osmosis protocl with base price usdc-noble -> 2 queries for the price of OSMO ( {"base_price": { "currency": "OSMO"}}) and then for the price of our l1 base asset NLS ( {"base_price": { "currency": "NLS"}})
// {"data":{"amount":{"amount":"20000000000000000000000000","ticker":"NLS"},"amount_quote":{"amount":"266989135256384142681063","ticker":"USDC_NOBLE"}}}
// {"data":{"amount":{"amount":"204636307898908853530883789062500000","ticker":"OSMO"},"amount_quote":{"amount":"140267298255822741431020268016499489","ticker":"USDC_NOBLE"}}}
// 2 calculations; 1 for osmo to protocol's base price and 1 for unls to protocol's base price
// := osmo * (usdc_noble/osmo) ?
// := unls * (usdc_noble/unls) ?

// lets say fee is paid in USDC_NOBLE
// {"data":{"amount":{"amount":"1","ticker":"USDC_NOBLE"},"amount_quote":{"amount":"1","ticker":"USDC_NOBLE"}}}
// {"data":{"amount":{"amount":"20000000000000000000000000","ticker":"NLS"},"amount_quote":{"amount":"266989135256384142681063","ticker":"USDC_NOBLE"}}}
// := usdc * (usdc_noble/usdc_noble) ?
// := unls * (usdc_noble/unls) ?

// lets say we use a short protocol st_atom with base price st_atom and fee is paid in OSMO
// {"data":{"amount":{"amount":"17462298274040222167968750000000000000","ticker":"OSMO"},"amount_quote":{"amount":"898204223623089861470521043125424927","ticker":"ST_ATOM"}}}
// {"data":{"amount":{"amount":"198523347012726641969138086096791084855","ticker":"NLS"},"amount_quote":{"amount":"199016001066004468121070185771025592","ticker":"ST_ATOM"}}}
// := osmo * (st_atom/osmo) ?
// := unls * (st_atom/unls) ?

// lets say we use a short protocol st_atom with base price st_atom and fee is paid in ST_ATOM
// {"data":{"amount":{"amount":"1","ticker":"ST_ATOM"},"amount_quote":{"amount":"1","ticker":"ST_ATOM"}}}
// := st_atom * (st_atom/st_atom) ?
// := unls * (st_atom/unls) ?

// PROBLEM: how to ensure tickers for queries are correct if we don't have them in the tax params? We have to keep an up-to date map inside the l1 which could also be updated through a gov prop for easier migration
// but we will still have the dependency of keeping the map up-to-date in the l1

var baseAssetTicker = strings.ToUpper(params.HumanCoinUnit)

type PriceFeed struct {
	Amount string `json:"amount"`
	Ticker string `json:"ticker"`
}
