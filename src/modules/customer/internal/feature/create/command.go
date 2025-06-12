package create

type CreateCustomerCommand struct {
	CreateCustomerRequest // embeded type มาเพราะหน้าตาเหมือนกัน
}

type CreateCustomerCommandResult struct {
	CreateCustomerResponse // embeded type มาเพราะหน้าตาเหมือนกัน
}

// ฟังก์ชันช่วยสร้าง CreateCustomerCommandResult
func NewCreateCustomerCommandResult(id int64) *CreateCustomerCommandResult {
	return &CreateCustomerCommandResult{
		CreateCustomerResponse{
			ID: id,
		},
	}
}
