package util

const (
	EUR  = "EUR"
	CAD  = "CAD"
	USD  = "USD"
	GBP  = "GBP"
	GHS  = "GHS"
	NGN  = "NGN"
	KES  = "KES"
	FCFA = "FCFA"
	ZAR  = "ZAR"
	YEN  = "YEN"
	CNY  = "CNY"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case EUR, CAD, USD, GBP, GHS, NGN, KES, FCFA, ZAR, CNY:
		return true
	}
	return false
}
