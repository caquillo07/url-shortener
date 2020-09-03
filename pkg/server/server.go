package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
	"go.uber.org/zap"

	"github.com/caquillo07/sample_url_shortener/pkg/storage"
)

const (
	port = "3000"
)

type Server struct {
	storage    storage.Storage
	app    *fiber.App
}

// Handler custom handler to allow for error checking
type Handler func(c *fiber.Ctx) error

func NewSever(store storage.Storage) *Server {
	s := &Server{
		storage: store,
		app: fiber.New(&fiber.Settings{
			DisableStartupMessage: true,
		}),
	}

	s.app.Use(middleware.Logger())
	s.app.Use(Recover())
	s.app.Use(middleware.RequestID())
	s.app.Use(cors.New())
	// Custom error handler
	s.app.Settings.ErrorHandler = errorHandler

	s.app.Post("/new", handler(s.createURL))
	s.app.Get("/:id", handler(s.handleVisit))
	return s
}

func (s *Server) Run() error {

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c

		// sig is a ^C, handle it
		// create context with timeout (this could also be config driven)
		const shutdownTimeout = 10 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// start http shutdown
		zap.L().Info("shutting down..")
		if err := s.app.Shutdown(); err != nil {
			zap.L().Error("error when shutting down server", zap.Error(err))
		}

		// verify, in worst case call cancel via defer
		select {
		case <-time.After(shutdownTimeout + (time.Second * 1)):
			zap.L().Info("not all connections done")
		case <-ctx.Done():
			// done
		}
	}()

	zap.L().Info("Listening on http://localhost:" + port)
	return s.app.Listen(port)
}

// handler is a wrapper that allows the the server route functions to return
// an error. This is useful, because otherwise you would have to do the call
// to the Next handler call on each error. Ain't no body got time for that
func handler(h Handler) fiber.Handler {
	return func(ctx *fiber.Ctx) {
		if err := h(ctx); err != nil {
			ctx.Next(err)
			return
		}
	}
}

func errorHandler(ctx *fiber.Ctx, err error) {
	type errResponse struct {
		Error string `json:"error"`
		Code  int    `json:"code"`
	}

	logError := func(err error) {
		if err == nil {
			return
		}
		zap.L().Error("failed to send error response", zap.Error(err))
	}

	// weirdly this error is not typed, so it has to be string matched
	if has := strings.HasPrefix(err.Error(), "bodyparser: cannot parse content-type:"); has {
		logError(ctx.Status(http.StatusBadRequest).JSON(errResponse{
			Error: "Content-Type: application/json header is required",
			Code:  http.StatusBadRequest,
		}))
		return
	}

	e, ok := err.(apiError)
	if !ok {
		zap.L().Info("masking internal error: ", zap.Error(err))
		logError(ctx.Status(http.StatusInternalServerError).JSON(errResponse{
			Error: "internal error",
			Code:  http.StatusInternalServerError,
		}))
		return
	}

	logError(ctx.Status(e.Code).JSON(errResponse{
		Error: e.Message,
		Code:  e.Code,
	}))
}

// Custom recover middleware to get stacktrace printed on error
// Recover will recover from panics and calls the ErrorHandler
func Recover() fiber.Handler {
	return func(ctx *fiber.Ctx) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				fmt.Printf("recovered from panic: %v\n%s", err, debug.Stack())
				ctx.Next(err)
				return
			}
		}()
		ctx.Next()
	}
}
