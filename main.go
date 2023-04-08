package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	countExec = 3
)

type Server struct {
	onlineExec int
	sync.Mutex
	queue, quit chan int
}

func (s *Server) Up() {
	fmt.Println("Запуск сервера, количество обработчиков:", countExec)
	s.Lock()
	s.queue = make(chan int)
	s.quit = make(chan int)
	for i := 0; i < countExec; i++ {
		go s.execute(i)
	}
	s.Unlock()
}

func (s *Server) Shutdown() {
	fmt.Println("Остановка сервера...")
	for i := 0; i < countExec; i++ {
		s.quit <- 0
	}
	for {
		s.Lock()
		if s.onlineExec == 0 {
			return
		}
		s.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *Server) Scheduler(n int) {
	s.queue <- n
}

func (s *Server) execute(k int) {
	fmt.Println("Обработчик", k, "запускается")
	s.Lock()
	s.onlineExec++
	s.Unlock()
	for {
		select {
		case <-s.quit:
			fmt.Println("Обработчик", k, "останавливается")
			s.Lock()
			s.onlineExec--
			s.Unlock()
			return
		default:
			i := <-s.queue
			d := rand.Intn(3) + 1
			time.Sleep(time.Duration(d) * time.Second)
			i2 := i * i
			fmt.Println("Обработчик", k, "квадрат числа", i, "равен", i2)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

var s Server

func init() {
	s = Server{}
	rand.Seed(time.Now().UnixNano())
}

func handle(c chan os.Signal) {
	<-c
	fmt.Println("exit")
	s.Shutdown()
	os.Exit(0)
}

func main() {
	fmt.Println("Graceful shutdown\n-----------------")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go handle(c)

	s.Up()

	for i := 0; i < 30; i++ {
		s.Scheduler(i)
	}

	time.Sleep(time.Minute)

}