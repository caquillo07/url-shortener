package server

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber"
	"go.uber.org/zap"

	"github.com/caquillo07/sample_url_shortener/pkg/storage"
)

func (s *Server) createURL(c *fiber.Ctx) error {
	type CreateRequest struct {
		URL string `json:"url"`
	}
	type CreateResponse struct {
		URL string `json:"url"`
	}

	req := CreateRequest{}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if req.URL == "" {
		return newAPIError("url is required", http.StatusBadRequest)
	}
	url := req.URL
	if !strings.Contains(url, "http") {
		url = "http://" + url
	}
	shortURL := &storage.ShortURL{
		URL: url,
	}
	if err := s.storage.CreateURL(c.Context(), shortURL); err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(CreateResponse{
		URL: getShortURL(c, shortURL),
	})
}

func (s *Server) handleVisit(c *fiber.Ctx) error {
	shortURL, err := s.storage.GetURL(c.Context(), c.Params("id"))
	if err != nil {
		return newAPIError("url not found", http.StatusNotFound)
	}

	// send off on its own routine to not hold the user back
	go func(c fiber.Ctx) {
		if err := s.storage.RegisterVisit(c.Context(), shortURL.ID, &storage.Visit{
			URLID:     shortURL.ID,
			IP:        c.IP(),
			Referer:   c.Get("Referer"),
			UserAgent: string(c.Fasthttp.UserAgent()),
		}); err != nil {
			// if we get an error recording the visit, we do not want to stop
			// the caller from redirecting, so just log and move on
			zap.L().Error("error registering visit", zap.Error(err))
		}
	}(*c)

	c.Redirect(shortURL.URL, http.StatusTemporaryRedirect)
	return nil
}

func getShortURL(c *fiber.Ctx, su *storage.ShortURL) string {
	return c.Protocol() + "://" + c.Hostname() + "/" + su.ID
}
