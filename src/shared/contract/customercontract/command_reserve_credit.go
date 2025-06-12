package customercontract

type ReserveCreditCommand struct {
	CustomerID   int64 `json:"customer_id"`
	CreditAmount int   `json:"credit_amount"`
}
