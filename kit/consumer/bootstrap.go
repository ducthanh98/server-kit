package consumer

import (
	"github.com/ducthanh98/server-kit/kit/logger"
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
		logger.Log.Fatal("Cannot initialize an empty worker.")
	}
	time.Sleep(2 * time.Second)
	w.Walk()
}
