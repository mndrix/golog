package read

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
    single[`x(a).`] = `x(a)`
    single[`x(a,b,c).`] = `x(a, b, c)`
    single[`x((a,b)).`] = `x(','(a, b))`
    single[`x(A).`] = `x(A)`
    single[`amen :- true.`] = `:-(amen, true)`
    single[`bee(X) :- X=b.`] = `:-(bee(X), =(X, b))`
    single[`zero(X) :- 0 =:= X.`] = `:-(zero(X), =:=(0, X))`
    single[`succ(0,1) :- true.`] = `:-(succ(0, 1), true)`
    single[`pi(3.14159).`] = `pi(3.14159)`
    single[`etc(_,_).`] = `etc(_, _)`
    single[`[].`] = `[]`                    // based on examples in §6.3.5.1
    single[`[a].`] = `'.'(a, [])`
    single[`[a,b].`] = `'.'(a, '.'(b, []))`
    single[`[a|b].`] = `'.'(a, b)`
    single[`"".`] = `[]`
    single[`"hi".`] = `'.'(104, '.'(105, []))`
    single[`"✓".`] = `'.'(10003, [])`       // 0x2713 Unicode
    single[`''.`] = `''`    // empty atom
    single[`'hi'.`] = `hi`
    for test, wanted := range single {
        got, err := Term(test)
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
    oneTwo, err := TermAll(oneTwoStr)
    maybePanic(err)
    if oneTwo[0].String() != "one" {
        t.Errorf("Expected `one` in %#v", oneTwo)
    }
    if oneTwo[1].String() != "two" {
        t.Errorf("Expected `two` in %#v", oneTwo)
    }
}


func TestEolComment(t *testing.T) {
    terms := TermAll_(`
        one.  % shouldn't hide following term
        two.
    `)
    if len(terms) != 2 {
        t.Errorf("Wrong number of terms: %d vs 2", len(terms))
    }
    if terms[0].String() != "one" {
        t.Errorf("Expected `one` in %#v", terms)
    }
    if terms[1].String() != "two" {
        t.Errorf("Expected `two` in %#v", terms)
    }
}
