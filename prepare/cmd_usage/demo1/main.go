package main

import "fmt"
import "os/exec"

func main() {
	var (
		cmd *exec.Cmd
		err error
	)

	cmd = exec.Command("/bin/bash","-c","ls")

	err = cmd.Run()
	fmt.Println(err)
}
