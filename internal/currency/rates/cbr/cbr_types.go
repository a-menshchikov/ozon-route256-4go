package cbr

type currencyList struct {
	Date       string     `xml:",attr"`
	Currencies []currency `xml:"Valute"`
}

type currency struct {
	ID       string `xml:",attr"`
	NumCode  int
	CharCode string
	Nominal  int
	Name     string
	Value    float64
}
