module github.com/fiatjaf/go-lnurl

go 1.14

require (
	github.com/btcsuite/btcd/btcec/v2 v2.2.0
	github.com/fiatjaf/ln-decodepay v1.1.0
	github.com/tidwall/gjson v1.6.1
)

replace github.com/fiatjaf/ln-decodepay => github.com/breez/ln-decodepay v1.4.1-0.20220718090404-ae4beb90748a
