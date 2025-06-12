package customercontract

type GetCustomerByIDQuery struct {
	ID int64 `json:"id"`
}

type GetCustomerByIDQueryResult struct {
	ID     int64  `json:"id"`
	Email  string `json:"email"`
	Credit int    `json:"credit"`
}
