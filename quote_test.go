package luautil 

import (
	"strconv"
	"testing"
)

func TestAll(t *testing.T) {
	var test []byte
	var expected string
	for i := 0; i < 256; i++ {
		test = append(test, byte(i))
		expected += "\\" + strconv.Itoa(i)
	}

	result := Quote(string(test))

	if expected != result {
		t.Errorf("Expected got %s", result)
	}
}

func TestEscapes(t *testing.T) {
	test := "\\ \" \a \b \t \n \v \f \r"
	expected := `"\\ \" \a \b \t \n \v \f \r"`
	result := Quote(test)

	if expected != result {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestValid(t *testing.T) {
	test := " !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	expected := "\" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~\""
	result := Quote(test)

	if expected != result {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestExtended(t *testing.T) {
	var test []byte

	expected := `"`
	for i := 0; i < 7; i++ {
		test = append(test, byte(i))
		expected += "\\00" + strconv.Itoa(i)
	}
	for i := 14; i < 32; i++ {
		test = append(test, byte(i))
		expected += "\\0" + strconv.Itoa(i)
	}
	for i := 127; i < 256; i++ {
		test = append(test, byte(i))
		expected += "\\" + strconv.Itoa(i)
	}
	expected += `"`

	result := Quote(string(test))

	if expected != result {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}