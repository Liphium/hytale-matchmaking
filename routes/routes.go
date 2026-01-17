package routes

import (
	control_routes "github.com/Liphium/hytale-matchmaking/routes/control"
	servers_routes "github.com/Liphium/hytale-matchmaking/routes/servers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) {
	router.Route("/control", control_routes.SetupRoutes)
	router.Route("/servers", servers_routes.SetupRoutes)
}
