package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

type CroneJob struct {
	expr     *cronexpr.Expression
	nextTime time.Time
}

func main() {
	var (
		expr          *cronexpr.Expression //an Expression
		croneJob      *CroneJob
		now           time.Time
		scheduleTable map[string]*CroneJob
	)
	//a map, use to storage CroneJob
	scheduleTable = make(map[string]*CroneJob, 1000)
	now = time.Now()
	//parse the expr without response error
	expr = cronexpr.MustParse("*/3 * * * * * *")
	//CroneJob1
	croneJob = &CroneJob{
		expr:     expr,
		nextTime: expr.Next(now),
	}
	scheduleTable["job1"] = croneJob
	//CroneJob2
	expr = cronexpr.MustParse("*/2 * * * * * *")
	croneJob = &CroneJob{
		expr:     expr,
		nextTime: expr.Next(now),
	}
	scheduleTable["job2"] = croneJob

	go func() {
		var (
			jobName  string
			now      time.Time
			croneJob *CroneJob
		)
		// still work until this go func finished
		for{
			now = time.Now()
			for jobName, croneJob = range scheduleTable {
				if croneJob.nextTime.Before(now) || croneJob.nextTime.Equal(now) {
					// new goroutine
					go func(jobName string) {
						fmt.Println(jobName , "exec")
					}(jobName)
					croneJob.nextTime = croneJob.expr.Next(now)
					fmt.Println(jobName, "next exec time: ", croneJob.nextTime)
				}

			}
			// sleep 100ns
			select {
			case <-time.NewTimer(100 * time.Millisecond).C:
			}
		}
	}()
	time.Sleep(6 * time.Second)
}
