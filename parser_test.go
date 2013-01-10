package golog

import "testing"

func TestBasic(t *testing.T) {
    single := make(map[string]string)
    single[`hello.`] = `hello`
    single[`a + b.`] = `+(a, b)`
    single[`first, second.`] = `','(first, second)`
    single[`\+ j.`] = `\+(j)`
    single[`a + b*c.`] = `+(a, *(b, c))`    // test precedence
    single[`a + b + c.`] = `+(+(a, b), c)`  // test left associativity
    single[`a^b^c.`] = `^(a, ^(b, c))`      // test right associativity
    for test, wanted := range single {
        got, err := ReadTerm(test)
        maybePanic(err)
        if got.String() != wanted {
            t.Errorf("Reading `%s` gave `%s` instead of `%s`", test, got, wanted)
        }
    }

    // reading a couple simple terms
    oneTwoStr := `one. two.`
    oneTwo, err := ReadTermAll(oneTwoStr)
    maybePanic(err)
    if oneTwo[0].String() != "one" {
        t.Errorf("Expected `one` in %#v", oneTwo)
    }
    if oneTwo[1].String() != "two" {
        t.Errorf("Expected `two` in %#v", oneTwo)
    }
}
