package main

import "fmt"

func (app *application) background(fn func()) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.log.Error(fmt.Sprintf("%v", err))
			}
		}()

		fn()
	}()
}
