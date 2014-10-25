package worker

type Message struct {
	ch chan int
	Id, Callback string
	Data map[string][]string
}

