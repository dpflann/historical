package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	out, err := exec.Command("bash", "-i", "-c", "history -r; history").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out[:]))
}