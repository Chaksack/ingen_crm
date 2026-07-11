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

	"ingencore/api/internal/audit"
	"ingencore/api/internal/auth"
	"ingencore/api/internal/config"
	"ingencore/api/internal/db"
	"ingencore/api/internal/entitlement"
	"ingencore/api/internal/modules/automation"
	"ingencore/api/internal/modules/collab"
	"ingencore/api/internal/modules/identity"
	"ingencore/api/internal/modules/sales"
	"ingencore/api/internal/modules/service"
	"ingencore/api/internal/push"
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

	pushSender := push.NewSender(cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.PushContact)
	pushH := push.NewHandler(pushSender)
	auditH := audit.NewHandler(pool)

	identityH := identity.NewHandler(pool, issuer)
	automationH := automation.NewHandler(pool, pushSender)
	salesH := sales.NewHandler(pool, automationH.Engine)
	serviceH := service.NewHandler(pool, automationH.Engine)
	collabH := collab.NewHandler(pool, hub, pushSender)

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
	api.Get("/invites/:token", identityH.GetInviteByToken)
	api.Post("/invites/:token/accept", identityH.AcceptInvite)
	api.Get("/push/vapid-public-key", pushH.VAPIDPublicKey)

	authed := api.Group("", auth.Middleware(issuer))
	// Every authenticated route runs inside one transaction with the caller's
	// org set as the RLS session variable — see internal/db/tenant.go. Must
	// be mounted before any subgroup so RequireAdmin/entitlement.Require can
	// rely on db.Tx(c) being populated.
	authed.Use(db.Middleware(pool))
	authed.Get("/me", identityH.Me)
	authed.Get("/me/manifest", identityH.Manifest)

	teamGroup := authed.Group("/team", identity.RequireAdmin())
	teamGroup.Get("/members", identityH.ListMembers)
	teamGroup.Patch("/members/:id", identityH.UpdateMember)
	teamGroup.Get("/invites", identityH.ListInvites)
	teamGroup.Post("/invites", identityH.CreateInvite)
	teamGroup.Delete("/invites/:id", identityH.RevokeInvite)
	teamGroup.Get("/business-units", identityH.ListBusinessUnits)
	teamGroup.Post("/business-units", identityH.CreateBusinessUnit)
	teamGroup.Patch("/business-units/:id", identityH.UpdateBusinessUnit)
	teamGroup.Delete("/business-units/:id", identityH.DeleteBusinessUnit)
	teamGroup.Get("/roles", identityH.ListRoles)
	teamGroup.Post("/roles", identityH.CreateRole)
	teamGroup.Delete("/roles/:id", identityH.DeleteRole)
	teamGroup.Get("/roles/:id/privileges", identityH.GetRolePrivileges)
	teamGroup.Put("/roles/:id/privileges", identityH.UpdateRolePrivileges)

	salesGroup := authed.Group("/sales", entitlement.Require("sales"))
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

	serviceGroup := authed.Group("/service", entitlement.Require("service"))
	serviceGroup.Get("/queues", serviceH.ListQueues)
	serviceGroup.Post("/queues", serviceH.CreateQueue)
	serviceGroup.Delete("/queues/:id", serviceH.DeleteQueue)
	serviceGroup.Patch("/queues/:id/routing", identity.RequireAdmin(), serviceH.UpdateQueueRouting)
	serviceGroup.Get("/queues/:id/members", identity.RequireAdmin(), serviceH.ListQueueMembers)
	serviceGroup.Post("/queues/:id/members", identity.RequireAdmin(), serviceH.AddQueueMember)
	serviceGroup.Delete("/queues/:id/members/:userId", identity.RequireAdmin(), serviceH.RemoveQueueMember)
	serviceGroup.Get("/cases", serviceH.ListCases)
	serviceGroup.Post("/cases", serviceH.CreateCase)
	serviceGroup.Get("/cases/:id", serviceH.GetCase)
	serviceGroup.Put("/cases/:id", serviceH.UpdateCase)
	serviceGroup.Delete("/cases/:id", serviceH.DeleteCase)
	serviceGroup.Post("/cases/:id/respond", serviceH.RespondToCase)
	serviceGroup.Get("/sla-policy", serviceH.GetSLAPolicy)
	serviceGroup.Put("/sla-policy", identity.RequireAdmin(), serviceH.UpdateSLAPolicy)
	serviceGroup.Get("/kb/suggest", serviceH.SuggestKBArticles)
	serviceGroup.Get("/kb", serviceH.ListKBArticles)
	serviceGroup.Post("/kb", serviceH.CreateKBArticle)
	serviceGroup.Get("/kb/:id", serviceH.GetKBArticle)
	serviceGroup.Put("/kb/:id", serviceH.UpdateKBArticle)
	serviceGroup.Post("/kb/:id/publish", identity.RequireAdmin(), serviceH.PublishKBArticle)
	serviceGroup.Post("/kb/:id/unpublish", identity.RequireAdmin(), serviceH.UnpublishKBArticle)
	serviceGroup.Delete("/kb/:id", identity.RequireAdmin(), serviceH.DeleteKBArticle)

	collabGroup := authed.Group("/collab", entitlement.Require("collab"))
	collabGroup.Get("/users", collabH.ListUsers)
	collabGroup.Get("/messages/:userID", collabH.MessageHistory)

	workflowGroup := authed.Group("/automation/workflows", identity.RequireAdmin())
	workflowGroup.Get("", automationH.ListWorkflows)
	workflowGroup.Post("", automationH.CreateWorkflow)
	workflowGroup.Put("/:id", automationH.UpdateWorkflow)
	workflowGroup.Delete("/:id", automationH.DeleteWorkflow)
	workflowGroup.Get("/:id/runs", automationH.ListWorkflowRuns)

	notifGroup := authed.Group("/notifications")
	notifGroup.Get("", automationH.ListNotifications)
	notifGroup.Get("/unread-count", automationH.UnreadNotificationCount)
	notifGroup.Post("/:id/read", automationH.MarkNotificationRead)

	pushGroup := authed.Group("/push")
	pushGroup.Post("/subscribe", pushH.Subscribe)
	pushGroup.Delete("/subscribe", pushH.Unsubscribe)

	authed.Get("/audit", identity.RequireAdmin(), auditH.List)

	go serviceH.StartBreachScanner(ctx, 30*time.Second)

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
