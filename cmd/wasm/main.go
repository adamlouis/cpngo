package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/adamlouis/cpngo/cpngo"
)

func fire(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		msg := fmt.Sprintf("go error: expected 1 argument, got %d\n", len(args))
		fmt.Printf(msg)
		return msg
	}

	net := &cpngo.Net{}
	if err := json.Unmarshal([]byte(args[0].String()), net); err != nil {
		msg := fmt.Sprintf("go error: failed to unmarshal net was json: %v", err)
		fmt.Println(msg)
		return msg
	}

	rnr, err := cpngo.NewRunner(net)
	if err != nil {
		msg := fmt.Sprintf("go error: failed to create runner: %v", err)
		fmt.Println(msg)
		return msg
	}

	if err := rnr.FireAny(); err != nil {
		msg := fmt.Sprintf("go error: failed to fire: %v", err)
		fmt.Println(msg)
		return msg
	}

	result := rnr.Net()
	j, err := json.Marshal(result)
	if err != nil {
		msg := fmt.Sprintf("go error: failed to marshal result: %v", err)
		fmt.Println(msg)
		return msg
	}
	return string(j)
}

func main() {
	js.Global().Set("GoFire", js.FuncOf(fire))
	select {} // run forever
}
