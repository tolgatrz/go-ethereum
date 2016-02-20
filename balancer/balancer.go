package balancer

import "runtime"

var B = New(runtime.GOMAXPROCS(0))

// Task repsents a single batch of work offered to a worker.
type Task struct {
	fn func() error // work function
	c  chan error   // return channel
}

// NewTask returns a new task and sets the proper fields.
func NewTask(fn func() error, c chan error) Task {
	return Task{
		fn: fn,
		c:  c,
	}
}

// Worker is a worker that will take one it's assigned tasks
// and execute it
type Worker struct {
	id      int       // worker id
	tasks   chan Task // tasks to do (buffered)
	pending int       // count of pending work
	index   int       // index in the heap
}

// work will take the oldest task and execute the function and
// yield the result back in to the return error channel.
func (w *Worker) work(queue <-chan Task) {
	for task := range queue {
		task.c <- task.fn() // get task and execute
	}
}

// Pool is a pool of workers
type Pool []*Worker

// Balancer is responsible for balancing any given tasks
// to the pool of workers. The workers are managed by the
// balancer and will try to make sure that the workers are
// equally balanced in "work to complete".
type Balancer struct {
	pool Pool
	work chan Task
}

// New returns a new load balancer
func New(poolSize int) *Balancer {
	balancer := &Balancer{
		work: make(chan Task, 1000),
		pool: make(Pool, poolSize),
	}

	// fill the pool with the given pool size
	for i := 0; i < poolSize; i++ {
		// create new worker
		worker := &Worker{id: i, tasks: make(chan Task, 100)}
		// spawn worker process
		go func(i int) {
			worker.work(balancer.work)
		}(i)
		balancer.pool[i] = worker
	}

	return balancer
}

// Push pushes the given tasks in to the work channel.
func (b *Balancer) Push(work Task) {
	b.work <- work
}
