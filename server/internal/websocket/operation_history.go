package websocket

type OperationHistory struct {
	Operations []Operation
}

func NewOperationHistory() *OperationHistory {
	return &OperationHistory{
		Operations: []Operation{},
	}
}

func (h *OperationHistory) Add(op Operation) {
	h.Operations = append(h.Operations, op)
}

func (h *OperationHistory) GetOperationsSince(version int64) []Operation {

	operations := []Operation{}

	for _, op := range h.Operations {

		if op.Version > version {
			operations = append(operations, op)
		}
	}

	return operations
}