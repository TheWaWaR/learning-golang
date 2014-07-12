
package main


import (
	"log"
	"sync"
	"time"
	"fmt"
)


type WorkerInterface interface {
	// ::See: http://tech.t9i.in/2014/01/inheritance-semantics-in-go/
	work(self WorkerInterface, resp chan interface{})
	finish(resp chan interface{}, result interface{})
}


type WorkerBase struct {}

func (wb *WorkerBase) work(self WorkerInterface, resp chan interface{}) {
	// log.Printf("call `WorkerBase.work()`\n")
	panic("`work(chann interface{})` Not implement!")
}

func (wb *WorkerBase) finish(resp chan interface{}, result interface{}) {
	/* For the case that `resp` channel already closed */
	
	defer func() {
		if r := recover(); r != nil {
			// >> Ignored
			// log.Printf("@!!! <%s> channel done closed!", self.tag)
		}
	}()
	resp <- result
	// log.Printf("@~ <%s> done sent!", self.tag)
}


// Pool state
const (
	PoolStarted	= 0
	PoolStopping	= 1
	PoolStopped	= 2
)

type Pool struct {
	size int
	gracefully bool
	request chan WorkerInterface
	response chan interface{}
	
	_signal chan (chan bool)
	_alive int
	_state int
	_stopMutex sync.Mutex
}

func (self *Pool) start() {
	resp := make(chan interface{})
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
				log.Printf("`resp`: Resume closed?: %q\n", r)
			}
		}()

		for {
			result, ok := <- resp
			if ok {
				self.response <- result // Forword result
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

				log.Printf("Worker done: result=%q, alive=%d\n", result, self._alive)
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
					close(resp)
					close(resume)
				}
				close(self._signal)
				close(self.request)
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

		worker, ok := <-self.request // Block here.
		if ok {
			wg.Add(1)
			// log.Printf("Start request: id=%d\n", req.id)
			go worker.work(worker, resp)
			
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
			log.Printf("self.request closed!\n")
			break
		}
	}

	if self.gracefully {
		log.Printf("...... Waiting for all worker finished!\n")
		wg.Wait()
		close(resp)
		close(resume)
	} else {
		log.Printf(">> Force quit!\n")
	}

	close(self.response)
	self._state = PoolStopped
	log.Printf("Oh, my lord. I'm back!\n")
}

func (self *Pool) apply(worker interface{}) {
	self.request <- worker.(WorkerInterface)
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
func (self *Pool) alive() int {
	return self._alive
}



/* ============================================================================
 *  Test section
 * ==========================================================================*/
type SomeWorker struct {
	WorkerBase
	tag string
}

func (sw *SomeWorker) work(self WorkerInterface, resp chan interface{}) {
	// Call super method
	// sw.WorkerBase.work(self, resp)
	
	log.Printf("<%s> is Working...\n", sw.tag)
	time.Sleep(2 * time.Second)
	log.Printf("<%s> is Done...\n", sw.tag)
	self.finish(resp, true)
}

func test_1() {
	pool_size := 4
	pool := Pool{size: pool_size, gracefully: true}
	pool.request = make(chan WorkerInterface)
	pool.response = make(chan interface{})
	
	go pool.start()
	go func() {
		for {
			result, ok := <- pool.response
			if ok {
				log.Printf("#### pool.response: %q\n", result)
			} else {
				log.Printf("#### pool.response Closed!\n")
				break
			}
		}
	} ()
	for i := 1; i <= 10; i++ {
		worker := &SomeWorker{tag:fmt.Sprintf("THE {%d} WORKER", i)}
		pool.apply(worker)
	}
	log.Printf("pool.alive: %d\n", pool.alive())
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
