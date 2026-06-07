package utils

import (
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// FormatCurrencyByRegion formats a float64 amount into a localized currency string
// based on the provided 2-letter ISO country code (e.g., "ID", "US", "SG", "JP").
func FormatCurrencyByRegion(amount float64, countryCode string) string {
	// Parse the region from the input country code and fall back to ID if invalid
	reg, err := language.ParseRegion(countryCode)
	if err != nil {
		reg = language.MustParseRegion("ID")
	}

	// Retrieve the official currency for the region and fall back to IDR if not found
	cur, ok := currency.FromRegion(reg)
	if !ok {
		cur = currency.IDR
	}

	// Initialize a printer tailored to the target regional language rules
	p := message.NewPrinter(language.Indonesian)

	// Format the final output using a narrow symbol to prefer "$" over "USD" or "Rp" over "IDR"
	return p.Sprintf("%v%.0f", currency.NarrowSymbol(cur), amount)
}
