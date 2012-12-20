package main

// standard library imports
import . "fmt"

// project library imports
import . "github.com/mndrix/golog"
import "github.com/mndrix/ps"

func main () {
    // Woody = howdy(woody)
    woody := NewTerm("woody")
    howdy := NewTerm("howdy", woody)

    // Foo = foo(hello, world, Woody)
    a := NewTerm("hello")
    b := NewTerm("world")
    foo := NewTerm("baz", a, b, howdy)
    Printf("%s is %s\n", foo.Indicator(), foo)

    // playing with inline definitions
    welcome := NewTerm(
        "welcome",
        NewTerm("to"),
        NewTerm(
            "prolog",
            NewTerm("in"),
            NewTerm("go"),
        ),
    )
    Printf("%s is %s\n", welcome.Indicator(), welcome)

    // playing with persistent maps
    m0 := ps.NewMap()
    m1 := m0.Set("hi", "friend")
    m2 := m1.Set("hello", "world")
    m3 := m2.Delete("hi")
    Printf("m0: %#v\n", m0)
    Printf("m1: %#v\n", m1)
    Printf("m2: %#v\n", m2)
    Printf("m3: %#v\n", m3)

    // playing with persistent lists
    l0 := ps.NewList()
    l1 := l0.Cons("first")
    l2 := l1.Cons("second")
    l3 := l2.Cons("third")
    l4 := l3.Reverse()
    show := func (v ps.Any) { Printf("%s, ", v) }
    Printf("\nl0: ")
    l0.ForEach(show)
    Printf("\nl1: ")
    l1.ForEach(show)
    Printf("\nl2: ")
    l2.ForEach(show)
    Printf("\nl3: ")
    l3.ForEach(show)
    Printf("\nl4: ")
    l4.ForEach(show)
    Println()
}
