package app

import "fmt"

func (app *Application) Background(fn func()) {
	app.Wg.Add(1)

	go func() {
		defer app.Wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.Log.Error(fmt.Sprintf("%v", err))
			}
		}()

		fn()
	}()
}
