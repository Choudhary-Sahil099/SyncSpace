package websocket

type OTEngine struct {
	History *OperationHistory
}

func NewOTEngine() *OTEngine {
	return &OTEngine{
		History: NewOperationHistory(),
	}
}

// Transform incoming operation against every operation
// that happened after its BaseVersion.
func (o *OTEngine) TransformAgainstHistory(
	incoming Operation,
) Operation {

	history := o.History.GetOperationsSince(
		incoming.BaseVersion,
	)

	for _, op := range history {
		incoming = Transform(
			incoming,
			op,
		)
	}

	return incoming
}