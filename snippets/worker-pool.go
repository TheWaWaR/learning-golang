
package main


import (
	"log"
	"sync"
	"time"
	"fmt"
)


type WorkerInterface interface {
	work(chan bool)
	finish(chan bool)
}

type Worker struct {
	tag string
}

func (self *Worker) work(done chan bool) {
	log.Printf("<%s> is Working...\n", self.tag)
	time.Sleep(2 * time.Second)
	log.Printf("<%s> is Done...\n", self.tag)
	self.finish(done)
}

func (self *Worker) finish(done chan bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("@!!! <%s> channel done closed!", self.tag)
		}
	}()
	done <- true
	log.Printf("@~ <%s> done sent!", self.tag)
}

type PoolRequest struct {
	id int
	worker Worker
}

// Pool state
const (
	PoolStarted	= 0
	PoolStopping	= 1
	PoolStopped	= 2
)

type Pool struct {
	size int
	requests chan PoolRequest
	gracefully bool
	
	_signal chan (chan bool)
	_alive int
	_state int
	_stopMutex sync.Mutex
}

func (self *Pool) start() {
	done := make(chan bool)
	resume := make(chan bool)
	aliveMutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	
	if self.gracefully != true {
		self.gracefully = false
	}

	self._signal = make(chan (chan bool))
	self._alive = 0
	self._state = PoolStarted
	self._stopMutex = sync.Mutex{}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Pool ERROR: %q\n", r)
		}
	}()
	
	// Counter
	go func () {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("`done`: Resume closed?: %q\n", r)
			}
		}()

		for {
			successed, ok := <- done
			if ok {
				wg.Done()
				
				aliveMutex.Lock()
				touched := (self._alive == self.size)
				self._alive -= 1
				aliveMutex.Unlock()
				
				if touched {
					log.Printf("...... Resuming pool")
					// `resume` may closed by Signal handler!
					resume <- true
				}

				log.Printf("Worker done: successed=%t, alive=%d\n", successed, self._alive)
			} else {
				log.Printf("Close counter: alive=%d\n", self._alive)
				break
			}
		}
	}()

	// Signal handler
	go func () {
		for {
			_stopped, ok := <- self._signal
			if ok {
				self._state = PoolStopping
				log.Printf("...... Stopping pool\n")
				if !self.gracefully {
					close(done)
					close(resume)
				}
				close(self._signal)
				close(self.requests)
				_stopped <- true
			} else {
				log.Printf("self._signal closed!\n")
				break
			}
		}
	}()

	// Request handler
	for {
		if self._state == PoolStopping || self._state == PoolStopped {
			break
		}

		req, ok := <-self.requests // Block here.
		if ok {
			wg.Add(1)
			log.Printf("Start request: id=%d\n", req.id)
			go req.worker.work(done)
			
			aliveMutex.Lock()
			self._alive += 1
			touched := (self._alive == self.size)
			aliveMutex.Unlock()
			
			if touched {
				log.Printf(" >> Paused pool!\n")
				// `resume` may closed by Signal handler!
				<-resume
				log.Printf(" >> Resumed pool!\n")
			}
		} else {
			log.Printf("self.requests closed!\n")
			break
		}
	}

	if self.gracefully {
		log.Printf("...... Waiting for all worker finished!\n")
		wg.Wait()
		close(done)
		close(resume)
	} else {
		log.Printf(">> Force quit!\n")
	}
	
	self._state = PoolStopped
	log.Printf("Oh, my lord. I'm back!\n")
}


func (self *Pool) stop() bool {
	
	var status bool
	log.Printf("...... Trying close pool\n")
	
	self._stopMutex.Lock()
	switch self._state {
	case PoolStopped:
		log.Printf("Pool already stopped.\n")
		status = false
	case PoolStopping:
		log.Printf("Pool in stopping process.\n")
		status = false
	default:
		log.Printf("Send stop signal\n")
		stopped := make(chan bool)
		self._signal <- stopped
		<-stopped
		status = true
	}
	self._stopMutex.Unlock()
	
	return status
}


func test_1() {
	pool := Pool{size: 2, gracefully: true}
	pool.requests = make(chan PoolRequest)
	go pool.start()

	for i := 1; i <= 10; i++ {
		req := PoolRequest{i, Worker{fmt.Sprintf("THE {%d} WORKER", i)}}
		pool.requests <- req
	}
	log.Printf("---------\n")
	pool.stop()
	pool.stop()
	pool.stop()
	pool.stop()
	
	time.Sleep(3 * time.Second)
	log.Printf("======================\n")
	pool.stop()

	time.Sleep(6 * time.Second)
	log.Printf("========== THE END ========\n")
}

func test_2() {
}


func main() {
	test_1()
}
