package main

import (
	"fmt"
	"os/exec"
)

func main() {
	var (
		cmd    *exec.Cmd
		output []byte
		err    error
	)
	//Generate Cmd(pipe)
	cmd = exec.Command("/bin/bash", "-c", "sleep 5;ls -l")

	if output, err = cmd.CombinedOutput(); err != nil {
		fmt.Println(err)
		return
	}

	//Print output of the fork
	fmt.Println(string(output))
}
