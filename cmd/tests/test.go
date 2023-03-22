package main

// type Confuguration struct {
// 	WorkersNumber int
// }

// var testConf = Confuguration{WorkersNumber: 90}

// var f workerpool.TypicalJobFunction[string] = func(data string) workerpool.JobResult {

// 	log.Println("executing function:" + data)
// 	return workerpool.JobResult{Result: "executed well"}
// }

// func main() {
// 	ctx := context.Background()
// 	wp := workerpool.NewWorkerPool[string](testConf.WorkersNumber)
// 	wp.Start(ctx)

// 	for i := 0; i < 80; i++ {
// 		n := fmt.Sprintf("job #%v", i)
// 		j := workerpool.Job[string]{Name: n, Func: f}
// 		wp.SendJob(ctx, j)
// 	}
// 	wp.WaitGroup.Wait()

// 	log.Println("application is finished")
// }
