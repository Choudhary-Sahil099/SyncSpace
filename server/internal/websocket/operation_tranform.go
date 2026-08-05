package websocket

func transformInsertInsert(
	incoming Operation,
	history Operation,
) Operation {

	if history.Position < incoming.Position ||
		(history.Position == incoming.Position &&
			history.UserID < incoming.UserID) {

		incoming.Position += len(history.Text)
	}

	return incoming
}

func transformInsertDelete(
	incoming Operation,
	history Operation,
) Operation {

	if history.Position < incoming.Position {

		shift := history.Length

		if history.Position+history.Length > incoming.Position {
			shift = incoming.Position - history.Position
		}

		incoming.Position -= shift
	}

	return incoming
}

func transformDeleteInsert(
	incoming Operation,
	history Operation,
) Operation {

	if history.Position <= incoming.Position {

		incoming.Position += len(history.Text)
	}

	return incoming
}

func transformDeleteDelete(
	incoming Operation,
	history Operation,
) Operation {

	historyStart := history.Position
	historyEnd := history.Position + history.Length

	incomingStart := incoming.Position
	incomingEnd := incoming.Position + incoming.Length

	// No overlap: history delete is completely before incoming.
	if historyEnd <= incomingStart {

		incoming.Position -= history.Length
		return incoming
	}

	// No overlap: incoming delete is completely before history.
	if incomingEnd <= historyStart {
		return incoming
	}

	// if overlap exist
	overlapStart := max(historyStart, incomingStart)
	overlapEnd := min(historyEnd, incomingEnd)

	overlap := overlapEnd - overlapStart

	incoming.Length -= overlap

	if historyStart < incomingStart {
		incoming.Position = historyStart
	}

	if incoming.Length < 0 {
		incoming.Length = 0
	}

	return incoming
}

func Transform(
	incoming Operation,
	history Operation,
) Operation {

	switch {

	case incoming.Type == InsertOperation &&
		history.Type == InsertOperation:

		return transformInsertInsert(
			incoming,
			history,
		)

	case incoming.Type == InsertOperation &&
		history.Type == DeleteOperation:

		return transformInsertDelete(
			incoming,
			history,
		)

	case incoming.Type == DeleteOperation &&
		history.Type == InsertOperation:

		return transformDeleteInsert(
			incoming,
			history,
		)

	case incoming.Type == DeleteOperation &&
		history.Type == DeleteOperation:

		return transformDeleteDelete(
			incoming,
			history,
		)

	default:
		return incoming
	}
}