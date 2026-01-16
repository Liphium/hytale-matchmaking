package main

import (
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	loadTokens()
	credential := os.Getenv("CREDENTIAL")

	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	app.Route("/api", func(router fiber.Router) {

		router.Use(func(c *fiber.Ctx) error {

			cred := c.Query("credential", "-")
			if cred != credential {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			return c.Next()
		})

		// Handle for managing the get token
		router.Get("/get_token", func(c *fiber.Ctx) error {
			token, ok := getTokenForServer()
			if !ok {
				return c.SendStatus(fiber.StatusNotFound)
			}

			token.Mutex.Lock()
			defer token.Mutex.Unlock()

			return c.JSON(fiber.Map{
				"id":            token.Id,
				"access_token":  token.Token.AccessToken,
				"refresh_token": token.Token.RefreshToken,
				"uuid":          token.Token.UUID,
			})
		})

		// Handle setting token
		router.Get("/set_token", func(c *fiber.Ctx) error {
			id := c.Query("id", "-")
			token := c.Query("access_token", "-")
			i, err := strconv.Atoi(id)
			if token == "-" || id == "-" || err != nil {
				return c.SendStatus(fiber.StatusBadRequest)
			}

			replaceToken(i, token)
			return c.SendStatus(fiber.StatusOK)
		})

		// Handle renews for the tokens
		router.Get("/renew", func(c *fiber.Ctx) error {
			id := c.Query("id", "-")
			i, err := strconv.Atoi(id)
			if err != nil {
				return c.SendStatus(fiber.StatusBadRequest)
			}

			refreshToken(i)
			return c.SendStatus(fiber.StatusOK)
		})

		// Token for generating a new token using the API
		router.Get("/add_new", generateToken)
	})

	// Hello world handler for health checks
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Hytale Auth Manager!")
	})

	app.Listen(os.Getenv("LISTEN"))
}
