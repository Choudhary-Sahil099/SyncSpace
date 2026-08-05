package websocket

// basic operation history struct
type OperationHistory struct {
	Operations []Operation
}

// func for the newHistory
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

func Transform(incoming Operation, history Operation) Operation {
	if incoming.Type == "insert" && history.Type == "insert" {
		if history.Position < incoming.Position ||
			(history.Position == incoming.Position &&
				history.UserID < incoming.UserID) {

			incoming.Position += len(history.Text)
		}
	}
	return incoming
}
