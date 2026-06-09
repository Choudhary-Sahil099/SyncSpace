package websocket

import "fmt"

func ApplyOperation(
	document string,
	op *Operation,
) (string, error) {

	if op == nil {
		return document, nil
	}

	switch op.Type {

	case "insert":

		if op.Position < 0 ||
			op.Position > len(document) {

			return document,
				fmt.Errorf("invalid position")
		}

		return document[:op.Position] +
				op.Text +
				document[op.Position:],
			nil

	case "delete":

		if op.Position < 0 ||
			op.Position+op.Length > len(document) {

			return document,
				fmt.Errorf("invalid delete")
		}

		return document[:op.Position] +
				document[op.Position+op.Length:],
			nil

	default:

		return document,
			fmt.Errorf("unknown operation")
	}
}
