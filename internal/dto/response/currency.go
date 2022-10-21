package response

type SetCurrency bool

type ListCurrencies struct {
	Current string
	List    []string
}
