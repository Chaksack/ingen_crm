ALTER TABLE "support_tickets" ADD COLUMN "preferred_contact" text DEFAULT 'email' NOT NULL;--> statement-breakpoint
ALTER TABLE "support_tickets" ADD COLUMN "access_token" text;--> statement-breakpoint
ALTER TABLE "support_tickets" ADD CONSTRAINT "support_tickets_access_token_unique" UNIQUE("access_token");