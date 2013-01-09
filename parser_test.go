package golog

import "testing"

func TestBasic(t *testing.T) {

    // reading a single simple term
    helloStr := `hello.`
    hello, err := ReadStringOne(helloStr, Read)
    maybePanic(err)
    if hello.String() != "hello" {
        t.Errorf("Reading `%s` gave `%s`", helloStr, hello.String())
    }

    // reading a couple simple terms
    oneTwoStr := `one. two.`
    oneTwo, err := ReadStringAll(oneTwoStr, Read)
    maybePanic(err)
    if oneTwo[0].String() != "one" {
        t.Errorf("Expected `one` in %#v", oneTwo)
    }
    if oneTwo[1].String() != "two" {
        t.Errorf("Expected `two` in %#v", oneTwo)
    }
}
