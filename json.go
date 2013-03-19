package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// A collection of convenience functions for sending JSON data in response
// to AJAX requests. The `json` function can be used whenever a request
// is successful.
//
// Values with type `jsonFail` and `jsonError` should be panic'd on. The
// JSON handler will render them appropriately. `jsonFail` should be used for
// user failures like bad input while `jsonError` should be used for serious
// faults, like bugs in the business logic or system failures.

func jsonHandler(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		c := newController(w, req)
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case authError:
					c.jsonRender(m{
						"status":  "noauth",
						"message": formatMessage(e.msg),
					})
				case jsonFail:
					c.jsonFail(e)
				case error:
					log.Printf("ERROR: %s", e)
					c.jsonRender(m{
						"status":  "error",
						"message": formatMessage(e.Error()),
					})
				default:
					panic(r)
				}
			}
		}()

		h(c)
	}
}

func (c *controller) json(userData interface{}) {
	c.jsonRender(m{
		"status":  "success",
		"content": userData,
	})
}

func (c *controller) jsonRender(data interface{}) {
	enc := json.NewEncoder(c.w)
	if err := enc.Encode(data); err != nil {
		// Something is seriously wrong.
		log.Fatal(err)
	}
}

type jsonFail string

func jsonf(format string, v ...interface{}) jsonFail {
	return jsonFail(fmt.Sprintf(format, v...))
}

func (jf jsonFail) Error() string {
	return string(jf)
}

func (c *controller) jsonFail(jf jsonFail) {
	c.jsonRender(m{
		"status":  "fail",
		"message": formatMessage(string(jf)),
	})
}
