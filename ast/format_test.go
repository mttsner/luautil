package ast

import (
	"strconv"
	"testing"
)

func TestValid(t *testing.T) {
	test := " !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	expected := " !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	result := formatString(test)

	if expected != result {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestEscape(t *testing.T) {
	test := "\\ \" \a \b \t \n \v \f \r"
	expected := "\\\\ \\\" \\a \\b \\t \\n \\v \\f \\r"
	result := formatString(test)

	if expected != result {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestExtended(t *testing.T) {
	var test []byte
	var expected string
	for i := 0; i < 7; i++ {
		test = append(test, byte(i))
		expected += "\\" + strconv.Itoa(i)
	}
	for i := 14; i < 32; i++ {
		test = append(test, byte(i))
		expected += "\\" + strconv.Itoa(i)
	}
	for i := 127; i < 256; i++ {
		test = append(test, byte(i))
		expected += "\\" + strconv.Itoa(i)
	}

	result := formatString(string(test))

	if expected != result {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}