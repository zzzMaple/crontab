package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

func main() {
	var (
		expr     *cronexpr.Expression
		err      error
		now      time.Time
		nextTime time.Time
	)
	if expr, err = cronexpr.Parse("*/3 * * * * * *"); err != nil {
		fmt.Println(err)
		return
	}

	now = time.Now()
	fmt.Println("Now: ", now)
	nextTime = expr.Next(now)

	//waiting until this timer broken
	time.AfterFunc(nextTime.Sub(now), func() {
		fmt.Println("already exec")
		fmt.Println("Next exec time:", nextTime)
	})


	//avoid main process finished
	time.Sleep(7 * time.Second)
}
