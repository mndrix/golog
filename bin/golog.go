package main

// standard library imports
import . "fmt"

// project library imports
import . "github.com/mndrix/golog"

func main () {
    db0 := NewDatabase()
    db1 := db0.Asserta(NewTerm("alpha"))
    db2 := db1.Asserta(NewTerm("beta"))
    db3 := db2.Asserta(NewTerm("gamma", NewTerm("greek to me")))
    Printf("db1:\n%s\n", db1.String())
    Printf("db2:\n%s\n", db2.String())
    Printf("db3:\n%s\n", db3.String())
}
