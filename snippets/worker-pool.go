
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
	panic("`work(chann interface{})` Not Implemented!")
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
	handler func(result interface{})
	
	_request chan WorkerInterface
	_response chan interface{}
	_alive int
	_state int
	_signal chan (chan bool)
	_mutex *sync.Mutex
}

func (p *Pool) init() {
	// Check attributes
	if p.size < 1 {
		panic(fmt.Sprintf("Invalid `Pool.size` = %d\n", p.size))
	}
	if p.handler == nil {
		panic("You must set `Pool.handler` !\n")
	}
	
	if p.gracefully != true {
		p.gracefully = false
	}
	
	p._alive = 0
	p._state = PoolStarted
	p._signal = make(chan (chan bool))
	p._mutex = &sync.Mutex{}
	p._request = make(chan WorkerInterface)
	p._response = make(chan interface{})
}

func (p *Pool) start() {
	if (p.size < 1 || p.handler == nil ||
		p._alive != 0 || p._state < PoolStarted ||
		p._signal == nil || p._mutex == nil || 
		p._request == nil || p._response == nil) {
		panic("You must call `Pool.init()` first!\n")
	}
	
	resp := make(chan interface{})
	resume := make(chan bool)
	aliveMutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Pool ERROR: %q\n", r)
		}
	}()

	// Response handler
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("`p.response`: closed: %q\n", r)
			}
		}()

		for{
			result, ok := <- p._response
			if ok {
				p.handler(result)
				log.Printf("#### pool.response: %q\n", result)
			} else {
				log.Printf("#### pool.response Closed!\n")
				break
			}

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
				p._response <- result // Forword result
				wg.Done()
				aliveMutex.Lock()
				touched := (p._alive == p.size)
				p._alive -= 1
				aliveMutex.Unlock()
				
				if touched {
					log.Printf("...... Resuming pool")
					// `resume` may closed by Signal handler!
					resume <- true
				}

				log.Printf("Worker done: result=%q, alive=%d\n", result, p._alive)
			} else {
				log.Printf("Close counter: alive=%d\n", p._alive)
				break
			}
		}
	}()

	// Signal handler
	go func () {
		for {
			_stopped, ok := <- p._signal
			if ok {
				p._state = PoolStopping
				log.Printf("...... Stopping pool\n")
				if !p.gracefully {
					close(resp)
					close(resume)
				}
				close(p._signal)
				close(p._request)
				_stopped <- true
			} else {
				log.Printf("p._signal closed!\n")
				break
			}
		}
	}()

	// Request handler
	for {
		if p._state == PoolStopping || p._state == PoolStopped {
			break
		}

		worker, ok := <-p._request // Block here.
		if ok {
			wg.Add(1)
			// log.Printf("Start request: id=%d\n", req.id)
			go worker.work(worker, resp)
			
			aliveMutex.Lock()
			p._alive += 1
			touched := (p._alive == p.size)
			aliveMutex.Unlock()
			
			if touched {
				log.Printf(" >> Paused pool!\n")
				// `resume` may closed by Signal handler!
				<-resume
				log.Printf(" >> Resumed pool!\n")
			}
		} else {
			log.Printf("p._request closed!\n")
			break
		}
	}

	// Clean stuff
	if p.gracefully {
		log.Printf("...... Waiting for all worker finished!\n")
		wg.Wait()
		close(resp)
		close(resume)
	} else {
		log.Printf(">> Force quit!\n")
	}

	close(p._response)
	p._state = PoolStopped
	log.Printf("Oh, my lord. I'm back!\n")
}

func (p *Pool) apply(worker interface{}) {
	p._request <- worker.(WorkerInterface)
}

func (p *Pool) stop() bool {
	
	var status bool
	log.Printf("...... Trying close pool\n")
	
	p._mutex.Lock()
	switch p._state {
	case PoolStopped:
		log.Printf("Pool already stopped.\n")
		status = false
	case PoolStopping:
		log.Printf("Pool in stopping process.\n")
		status = false
	default:
		log.Printf("Send stop signal\n")
		stopped := make(chan bool)
		p._signal <- stopped
		<-stopped
		status = true
	}
	p._mutex.Unlock()
	
	return status
}

func (p *Pool) terminate() bool {
	p.gracefully = false
	return p.stop()
}

func (p *Pool) alive() int {
	return p._alive
}



/* ============================================================================
 *  Test section
 * ==========================================================================*/
type SomeWorker struct {
	WorkerBase
	tag string
	sleep time.Duration
}

func (sw *SomeWorker) work(self WorkerInterface, resp chan interface{}) {
	// Call super method
	// sw.WorkerBase.work(self, resp)
	
	log.Printf("<%s> is Working...\n", sw.tag)
	time.Sleep(sw.sleep * time.Second)
	log.Printf("<%s> is Done...\n", sw.tag)
	self.finish(resp, true)
}

func test_1() {
	pool_size := 4
	pool := Pool{
		size : pool_size,
		gracefully : true,
	}
	pool.handler = func(result interface{}) {
		log.Printf("^Got result: [%t]\n", result)
	}
	pool.init()
	go pool.start()
	
	// Send request
	for i := 1; i <= 8; i++ {
		worker := &SomeWorker{
			tag : fmt.Sprintf("THE {%d} WORKER", i),
			sleep: time.Duration(i),
		}
		pool.apply(worker)
	}
	
	log.Printf("pool.alive: %d\n", pool.alive())
	log.Printf("---------\n")
	pool.stop()
	pool.stop()
	time.Sleep(1*time.Second)
	pool.stop()
	time.Sleep(1*time.Second)
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
