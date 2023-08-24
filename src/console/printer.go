package console

import (
	"fmt"
	"github.com/gosuri/uilive"
	"time"
)

type Printer struct {
	printch    chan string
	termWriter *uilive.Writer
}

func (pr *Printer) printStatus(msg string) {
	fmt.Fprintf(pr.termWriter, msg)
	time.Sleep(time.Millisecond * 100)
}

func (pr *Printer) Flush(msg string) {
	pr.printch <- msg
}

func (pr *Printer) Start() {
	for {
		select {
		case msg := <-pr.printch:
			pr.printStatus(msg)
		}
	}
}

func NewPrinter() *Printer {

	printer := Printer{}
	writer := uilive.New()
	// start listening for updates and render
	writer.Start()

	printer.printch = make(chan string, 32)
	printer.termWriter = writer
	return &printer
}
