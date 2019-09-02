package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type result struct {
	err    error
	output []byte
}

func main() {
	//execute a cmd, let it exec 2 second while sleep 2 second, echo hello
	var (
		cmd        *exec.Cmd
		res        *result
		resultChan chan *result
		ctx        context.Context
		cancelFunc context.CancelFunc
	)

	resultChan = make(chan *result, 1000)

	ctx, cancelFunc = context.WithCancel(context.TODO())
	go func() {
		var (
			err    error
			output []byte
		)

		cmd = exec.CommandContext(ctx, "/bin/bash", "-c", "sleep 2;echo hello")
		output, err = cmd.CombinedOutput()
		resultChan <- &result{
			err:    err,
			output: output,
		}
	}()

	time.Sleep(1 * time.Second)

	cancelFunc()

	res = <-resultChan
	fmt.Println(res.err, string(res.output))
}
