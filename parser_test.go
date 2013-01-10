package golog

import "testing"

func TestBasic(t *testing.T) {

    // read single terms
    single := make(map[string]string)
    single[`hello. world.`] = `hello`       // read only the first term
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

    // read single terms (with user-defined operators)
    user := make(map[string]string)
    user[`a x b.`] = `x(a, b)`
    user[`a x b x c.`] = `x(x(a, b), c)`
    user[`two weeks.`] = `weeks(two)`
    for test, wanted := range user {
        r, err := NewTermReader(test)
        maybePanic(err)
        r.Op(400, yfx, "x")
        r.Op(200, yf, "weeks")

        got, err := r.Next()
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
