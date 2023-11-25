package consumer

import (
	log "github.com/sirupsen/logrus"
	"time"
)

// Bootstrapping --
func Bootstrapping(taskDef string, initFunc Callback) {
	w := NewWorker(nil)
	initFunc(w)
	w.PrepareTask(taskDef)
	if Bootstrapper.F != nil && len(Bootstrapper.F) > 0 {
		for _, ff := range Bootstrapper.F {
			f := ff
			f.Init(w)
			defer f.Close(w)

			hs := f.GetHandlers()
			w.PrepareHandlers(hs)

			if p := w.GetProducer(); p != nil {
				f.SetProducer(p)
			}
		}
	} else {
		log.Fatal("Cannot initialize an empty worker.")
	}
	time.Sleep(2 * time.Second)
	w.Walk()
}
