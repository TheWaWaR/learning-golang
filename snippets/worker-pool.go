
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
			log.Printf("@!!! <%s> done closed!", self.tag)
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
	PoolStarted = -1
	PoolResumed = -2
	PoolPaused = -3
	PoolStopping = -4
	PoolStopped = -5
)

type Pool struct {
	size int
	requests chan PoolRequest
	gracefully bool
	_signal chan int
	_alive int
	_state int
}


func (self *Pool) start() {
	done := make(chan bool)
	resume := make(chan bool)
	m := sync.Mutex{}

	self._signal = make(chan int)
	self._alive = 0

	// Counter
	go func () {
		for {
			successed, ok := <- done
			if ok {
				m.Lock()
				self._alive -= 1
				if self._alive == (self.size - 1) {
					log.Printf("...... Resuming pool")
					resume <- true
				}
				m.Unlock()
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
			sig, ok := <- self._signal
			if ok {
				switch sig {
				case PoolStopped:
					log.Printf("...... Stopping pool\n")
					self._state = PoolStopping
					close(self._signal)
					close(done)
					close(resume)
					close(self.requests)
				}
			} else {
				log.Printf("self._signal closed!\n")
				break
			}
		}
	}()

	// Request handler
	for {
		if self._alive == self.size {
			log.Printf(" >> Paused pool!\n")
			self._state = PoolPaused
			<-resume
			self._state = PoolResumed
			log.Printf(" >> Resumed pool!\n")
		}
		
		req, ok := <-self.requests // Block here.
		if ok {
			log.Printf("Start request: id=%d\n", req.id)
			go req.worker.work(done)
			m.Lock()
			self._alive += 1
			m.Unlock()
		} else {
			log.Printf("self.requests closed!\n")
			break
		}
	}
	self._state = PoolStopped
	log.Printf("Oh, my lord. I'm back!\n")
}

func (self *Pool) stop() bool {
	log.Printf("...... Trying close pool\n")
	switch self._state {
	case PoolStopped:
		log.Printf("Pool already stopped.\n")
	case PoolStopping:
		log.Printf("Pool in stopping process.\n")
	default:
		log.Printf("Send stop signal\n")
		self._signal <- PoolStopped
		return true
	}
	return false
}


func test() {
	pool := Pool{size: 2}
	pool.requests = make(chan PoolRequest)
	go pool.start()

	for i := 1; i <= 20; i++ {
		req := PoolRequest{i, Worker{fmt.Sprintf("THE {%d} WORKER", i)}}
		pool.requests <- req
	}
	log.Printf("---------\n")
	pool.stop()
	
	// time.Sleep(1 * time.Second)
	// pool.pause()
	// time.Sleep(1 * time.Second)
	// pool.resume()
	time.Sleep(3 * time.Second)
	log.Printf("======================\n")
	pool.stop()

	time.Sleep(3 * time.Second)
	log.Printf("========== THE END ========\n")
}


func main() {
	test()
}
