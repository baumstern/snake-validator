package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type appError struct {
	err          error  // wrapped error
	responseText string // response text to user
	status       int    // HTTP status code
}

func (a *appError) Error() string {
	return fmt.Sprintf("%d (%s): %v", a.status, http.StatusText(a.status), a.err)
}

type AppHandler func(http.ResponseWriter, *http.Request) *appError

func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		log.Println(err.Error())
		http.Error(w, err.responseText, err.status)
	}
}

func NewGame(writer http.ResponseWriter, request *http.Request) *appError {
	if request.Method != http.MethodGet {
		return &appError{errors.New("invalid method: " + request.Method), "/new only accepts GET method", http.StatusMethodNotAllowed}
	}

	width, err := strconv.Atoi(request.URL.Query().Get("w"))
	if err != nil {
		return &appError{err, "w is not number", http.StatusBadRequest}
	}
	height, err := strconv.Atoi(request.URL.Query().Get("h"))
	if err != nil {
		return &appError{err, "h is not number", http.StatusBadRequest}
	}

	if width <= 0 || height <= 0 || width*height == 1 {
		return &appError{err, "map size should be greater or equal than 2", http.StatusBadRequest}
	}

	m, err := json.Marshal(newState(width, height))
	if err != nil {
		return &appError{err, "json marshal failed", http.StatusInternalServerError}
	}

	io.WriteString(writer, string(m))
	return nil
}

func Validate(writer http.ResponseWriter, request *http.Request) *appError {
	if request.Method != http.MethodPost {
		return &appError{errors.New("invalid method: " + request.Method), "/validate only accepts POST method, got: " + request.Method, http.StatusMethodNotAllowed}
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return &appError{errors.New("failed to read request body"), "failed to read request body", http.StatusInternalServerError}
	}
	if len(body) == 0 {
		return &appError{errors.New("empty body"), "body is missing", http.StatusBadRequest}
	}

	context := payloadValidate{}
	json.Unmarshal(body, &context)
	ticks := context.Ticks

	for i := 1; i < len(ticks); i++ {
		if (ticks[i].VelX == -1 && ticks[i-1].VelX == 1) ||
			(ticks[i].VelX == 1 && ticks[i-1].VelX == -1) ||
			(ticks[i].VelY == -1 && ticks[i-1].VelY == 1) ||
			(ticks[i].VelY == 1 && ticks[i-1].VelY == -1) {
			return &appError{errors.New("can't move backwards"), "can't move backwards", http.StatusTeapot}
		}
	}

	posX := context.Snake.X
	posY := context.Snake.Y
	for _, t := range ticks {
		if t.VelX == 1 && t.VelY == 1 ||
			t.VelX == -1 && t.VelY == -1 {
			return &appError{errors.New("can't move diagonal"), "can't move diagonal", http.StatusTeapot}
		}

		posX = posX + t.VelX
		posY = posY + t.VelY
		if posX < 0 || posX >= context.Width ||
			posY < 0 || posY >= context.Height {
			return &appError{errors.New("out of bound"), "out of bound", http.StatusTeapot}
		}

		if posX == context.Fruit.X && posY == context.Fruit.Y {
			context.Score++
			context.Fruit = newFruit(context.Width, context.Height, context.Fruit.X, context.Fruit.Y)
			context.Snake = updateSnake(posX, posY, t.VelX, t.VelY)

			m, _ := json.Marshal(context)
			io.WriteString(writer, string(m))
			return nil
		}
	}

	return &appError{errors.New("not reached to fruit"), "not reached to fruit", http.StatusNotFound}
}
