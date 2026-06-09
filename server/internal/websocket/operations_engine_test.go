package websocket

import "testing"

func TestInsertOperation(t *testing.T) {

	doc := "Hello World"

	result, err := ApplyOperation(
		doc,
		&Operation{
			Type: "insert",
			Position: 5,
			Text: ",",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	expected := "Hello, World"

	if result != expected {
		t.Fatalf(
			"expected %s got %s",
			expected,
			result,
		)
	}
}

func TestDeleteOperation(t *testing.T) {

	doc := "HelloXYZWorld"

	result, err := ApplyOperation(
		doc,
		&Operation{
			Type:     "delete",
			Position: 5,
			Length:   3,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	expected := "HelloWorld"

	if result != expected {
		t.Fatalf(
			"expected %s got %s",
			expected,
			result,
		)
	}
}