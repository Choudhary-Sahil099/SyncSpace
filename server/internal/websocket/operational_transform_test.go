package websocket

import (
	"testing"
)
func TestTransformInsertInsert(t *testing.T) {

	history := Operation{
		Type:     InsertOperation,
		Position: 2,
		Text:     "A",
		UserID:   "user1",
	}

	incoming := Operation{
		Type:     InsertOperation,
		Position: 5,
		Text:     "B",
		UserID:   "user2",
	}

	result := Transform(incoming, history)

	if result.Position != 6 {
		t.Errorf(
			"expected position 6, got %d",
			result.Position,
		)
	}
}

func TestTransformInsertDelete(t *testing.T) {

	history := Operation{
		Type:     DeleteOperation,
		Position: 2,
		Length:   3,
	}

	incoming := Operation{
		Type:     InsertOperation,
		Position: 8,
		Text:     "A",
	}

	result := Transform(incoming, history)

	if result.Position != 5 {
		t.Errorf(
			"expected position 5, got %d",
			result.Position,
		)
	}
}

func TestTransformDeleteInsert(t *testing.T) {

	history := Operation{
		Type:     InsertOperation,
		Position: 2,
		Text:     "ABC",
	}

	incoming := Operation{
		Type:     DeleteOperation,
		Position: 6,
		Length:   2,
	}

	result := Transform(incoming, history)

	if result.Position != 9 {
		t.Errorf(
			"expected position 9, got %d",
			result.Position,
		)
	}
}

func TestTransformDeleteDelete(t *testing.T) {

	history := Operation{
		Type:     DeleteOperation,
		Position: 2,
		Length:   3,
	}

	incoming := Operation{
		Type:     DeleteOperation,
		Position: 7,
		Length:   2,
	}

	result := Transform(incoming, history)

	if result.Position != 4 {
		t.Errorf(
			"expected position 4, got %d",
			result.Position,
		)
	}
}