package golog

import "testing"

func TestBasic(t *testing.T) {

    // reading a single simple term
    helloStr := `hello.`
    hello, err := ReadTermStringOne(helloStr, Read)
    maybePanic(err)
    if hello.String() != "hello" {
        t.Errorf("Reading `%s` gave `%s`", helloStr, hello.String())
    }

    // reading a couple simple terms
    oneTwoStr := `one. two.`
    oneTwo, err := ReadTermStringAll(oneTwoStr, Read)
    maybePanic(err)
    if oneTwo[0].String() != "one" {
        t.Errorf("Expected `one` in %#v", oneTwo)
    }
    if oneTwo[1].String() != "two" {
        t.Errorf("Expected `two` in %#v", oneTwo)
    }

    // a term with one infix operator
    plusStr := `a + b.`
    plus, err := ReadTermStringOne(plusStr, Read)
    maybePanic(err)
    if plus.String() != "+(a, b)" {
        t.Errorf("Expected `+(a, b)` but got %s", plus)
    }

    // a term with one prefix operator
    tildeStr := `\+ j.`
    tilde, err := ReadTermStringOne(tildeStr, Read)
    maybePanic(err)
    if tilde.String() != `\+(j)` {
        t.Errorf("Expected `\\+(j)` but got %s", tilde)
    }
}
