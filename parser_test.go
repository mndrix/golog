package golog

import "testing"

func TestBasic(t *testing.T) {
    helloStr := `hello(world).`
    hello, err := ReadStringOne(helloStr, Read)
    maybePanic(err)
    if hello.String() != "hello(world)" {
        t.Errorf("Reading `%s` gave `%s`", helloStr, hello.String())
    }
}
