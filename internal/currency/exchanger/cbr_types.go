package exchanger

type Currency struct {
	ID       string `xml:",attr"`
	NumCode  int
	CharCode string
	Nominal  int
	Name     string
	Value    float64
}

type CurrencyList struct {
	Date       string     `xml:",attr"`
	Currencies []Currency `xml:"Valute"`
}
