package pipe

type Line struct {
	pipes []Pipe
}

type Pipe struct {
	in   <-chan interface{}
	out  chan interface{}
	work func(in <-chan interface{}, out chan<- interface{})
}

const BUFFER_SIZE = 1000

func NewInit() *Line {
	return &Line{}
}

func (p *Pipe) run() <-chan interface{} {
	go func() {
		p.work(p.in, p.out)
		close(p.out)
	}()
	return p.out
}

func (pl *Line) Append(work func(in <-chan interface{}, out chan<- interface{})) *Line {
	new := Pipe{work: work}
	var size int
	if len(pl.pipes) != 0 {
		before := pl.pipes[len(pl.pipes)-1]
		new.in = before.out
		size = cap(before.out)
	} else {
		size = BUFFER_SIZE
	}
	new.out = make(chan interface{}, size)
	pl.pipes = append(pl.pipes, new)
	return pl
}

func (pl *Line) Chan() <-chan interface{} {
	for _, p := range pl.pipes {
		p.run()
	}
	return pl.pipes[len(pl.pipes)-1].out
}
