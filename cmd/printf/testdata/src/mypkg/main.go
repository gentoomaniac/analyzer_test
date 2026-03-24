package mypkg

import "fmt"

func uselessUseOfPrintF() {
	log(fmt.Sprintf("%s%s%s", "Hello,", "World!", func() string { return "foobar" }()))
}

func log(s string) {
	fmt.Println(s)
}
