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

	app.Use(func(c *fiber.Ctx) error {

		cred := c.Query("credential", "-")
		if cred != credential {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.Next()
	})

	// Handle for managing the get token
	app.Get("/get_token", func(c *fiber.Ctx) error {
		token, ok := getTokenForServer()
		if !ok {
			return c.SendStatus(fiber.StatusNotFound)
		}

		token.Mutex.Lock()
		defer token.Mutex.Unlock()

		return c.JSON(fiber.Map{
			"id":             token.Id,
			"identity_token": token.Token.IdentityToken,
			"session_token":  token.Token.SessionToken,
			"uuid":           token.Token.UUID,
		})
	})

	// Handle renews for the tokens
	app.Get("/renew", func(c *fiber.Ctx) error {
		id := c.Query("id", "-")
		i, err := strconv.Atoi(id)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		refreshToken(i)
		return c.SendStatus(fiber.StatusOK)
	})

	// Token for generating a new token using the API
	app.Get("/add_new", generateToken)

	app.Listen(os.Getenv("LISTEN"))
}
