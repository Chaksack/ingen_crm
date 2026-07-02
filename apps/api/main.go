package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"ingen.one/api/internal/auth"
	"ingen.one/api/internal/config"
	"ingen.one/api/internal/db"
	"ingen.one/api/internal/entitlement"
	"ingen.one/api/internal/modules/collab"
	"ingen.one/api/internal/modules/identity"
	"ingen.one/api/internal/modules/sales"
	"ingen.one/api/internal/modules/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer pool.Close()

	issuer := auth.NewTokenIssuer(cfg.JWTSecret)
	hub := collab.NewHub()

	identityH := identity.NewHandler(pool, issuer)
	salesH := sales.NewHandler(pool)
	serviceH := service.NewHandler(pool)
	collabH := collab.NewHandler(pool, hub)

	app := fiber.New(fiber.Config{ErrorHandler: apiErrorHandler})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		status := http.StatusOK
		message := "ok"
		if err := pool.Ping(c.Context()); err != nil {
			status = http.StatusServiceUnavailable
			message = "degraded: db unreachable"
		}
		return c.Status(status).JSON(fiber.Map{
			"status":  message,
			"time":    time.Now().UTC().Format(time.RFC3339),
			"version": "phase1-mvp",
		})
	})

	api := app.Group("/api/v1")

	api.Post("/auth/register", identityH.Register)
	api.Post("/auth/login", identityH.Login)

	authed := api.Group("", auth.Middleware(issuer))
	authed.Get("/me", identityH.Me)
	authed.Get("/me/manifest", identityH.Manifest)

	salesGroup := authed.Group("/sales", entitlement.Require(pool, "sales"))
	salesGroup.Get("/accounts", salesH.ListAccounts)
	salesGroup.Post("/accounts", salesH.CreateAccount)
	salesGroup.Get("/accounts/:id", salesH.GetAccount)
	salesGroup.Put("/accounts/:id", salesH.UpdateAccount)
	salesGroup.Delete("/accounts/:id", salesH.DeleteAccount)
	salesGroup.Get("/contacts", salesH.ListContacts)
	salesGroup.Post("/contacts", salesH.CreateContact)
	salesGroup.Get("/contacts/:id", salesH.GetContact)
	salesGroup.Put("/contacts/:id", salesH.UpdateContact)
	salesGroup.Delete("/contacts/:id", salesH.DeleteContact)
	salesGroup.Get("/leads", salesH.ListLeads)
	salesGroup.Post("/leads", salesH.CreateLead)
	salesGroup.Get("/leads/:id", salesH.GetLead)
	salesGroup.Put("/leads/:id", salesH.UpdateLead)
	salesGroup.Delete("/leads/:id", salesH.DeleteLead)

	serviceGroup := authed.Group("/service", entitlement.Require(pool, "service"))
	serviceGroup.Get("/queues", serviceH.ListQueues)
	serviceGroup.Post("/queues", serviceH.CreateQueue)
	serviceGroup.Delete("/queues/:id", serviceH.DeleteQueue)
	serviceGroup.Get("/cases", serviceH.ListCases)
	serviceGroup.Post("/cases", serviceH.CreateCase)
	serviceGroup.Get("/cases/:id", serviceH.GetCase)
	serviceGroup.Put("/cases/:id", serviceH.UpdateCase)
	serviceGroup.Delete("/cases/:id", serviceH.DeleteCase)

	collabGroup := authed.Group("/collab", entitlement.Require(pool, "collab"))
	collabGroup.Get("/users", collabH.ListUsers)
	collabGroup.Get("/messages/:userID", collabH.MessageHistory)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Use("/ws", auth.Middleware(issuer))
	app.Get("/ws", websocket.New(collabH.Serve))

	log.Printf("API listening on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("listen error: %v", err)
	}
}

func apiErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"
	if fe, ok := err.(*fiber.Error); ok {
		code = fe.Code
		message = fe.Message
	}
	return c.Status(code).JSON(fiber.Map{"error": message})
}
