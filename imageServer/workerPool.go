package imageServer

type WorkerPool struct {
}

type resizeRequest struct {
	image          image
	params         imageParams
	out            chan resizeRequestResult
	batchOperation BatchOperation
}

type resizeRequestResult struct {
	image          image
	params         imageParams
	batchOperation BatchOperation
}

func (wp *WorkerPool) Run(numOfWorkers int, workQueue chan resizeRequest) {
	for i := 0; i < numOfWorkers; i++ {
		w := newWorker(workQueue)
		w.Start()
	}
	return
}

type Worker struct {
	workQueue chan resizeRequest
	quit      chan bool
}

func newWorker(workQueue chan resizeRequest) Worker {
	return Worker{
		workQueue: workQueue,
		quit:      make(chan bool)}
}

func (w Worker) Start() {
	go func() {
		for {
			select {
			case job := <-w.workQueue:
				//TODO handle any errors and return them to out channel
				img, _ := resize(job.image.Contents, job.params)
				job.out <- resizeRequestResult{
					img,
					job.params,
					job.batchOperation,
				}
			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}
