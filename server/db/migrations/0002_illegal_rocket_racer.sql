CREATE TABLE "company_tax_settings" (
	"id" text PRIMARY KEY NOT NULL,
	"company_name" text DEFAULT 'Ingenicx' NOT NULL,
	"tin" text,
	"vat_number" text,
	"address" text,
	"phone" text,
	"email" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "invoice_taxes" (
	"id" text PRIMARY KEY NOT NULL,
	"invoice_id" text NOT NULL,
	"name" text NOT NULL,
	"rate" numeric(6, 3) NOT NULL,
	"amount" numeric(14, 2) NOT NULL,
	"compound" boolean DEFAULT false NOT NULL,
	"order" integer DEFAULT 0 NOT NULL
);
--> statement-breakpoint
CREATE TABLE "quotation_taxes" (
	"id" text PRIMARY KEY NOT NULL,
	"quotation_id" text NOT NULL,
	"name" text NOT NULL,
	"rate" numeric(6, 3) NOT NULL,
	"amount" numeric(14, 2) NOT NULL,
	"compound" boolean DEFAULT false NOT NULL,
	"order" integer DEFAULT 0 NOT NULL
);
--> statement-breakpoint
CREATE TABLE "tax_rates" (
	"id" text PRIMARY KEY NOT NULL,
	"name" text NOT NULL,
	"rate" numeric(6, 3) NOT NULL,
	"compound" boolean DEFAULT false NOT NULL,
	"order" integer DEFAULT 0 NOT NULL,
	"active" boolean DEFAULT true NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
--> statement-breakpoint
ALTER TABLE "invoices" ALTER COLUMN "tax_rate" DROP DEFAULT;--> statement-breakpoint
ALTER TABLE "invoices" ALTER COLUMN "tax_rate" DROP NOT NULL;--> statement-breakpoint
ALTER TABLE "quotations" ALTER COLUMN "tax_rate" DROP DEFAULT;--> statement-breakpoint
ALTER TABLE "quotations" ALTER COLUMN "tax_rate" DROP NOT NULL;--> statement-breakpoint
ALTER TABLE "invoices" ADD COLUMN "receipt_number" text;--> statement-breakpoint
ALTER TABLE "invoices" ADD COLUMN "tax_exempt" boolean DEFAULT false NOT NULL;--> statement-breakpoint
ALTER TABLE "quotations" ADD COLUMN "tax_exempt" boolean DEFAULT false NOT NULL;--> statement-breakpoint
ALTER TABLE "invoice_taxes" ADD CONSTRAINT "invoice_taxes_invoice_id_invoices_id_fk" FOREIGN KEY ("invoice_id") REFERENCES "public"."invoices"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "quotation_taxes" ADD CONSTRAINT "quotation_taxes_quotation_id_quotations_id_fk" FOREIGN KEY ("quotation_id") REFERENCES "public"."quotations"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "invoices" ADD CONSTRAINT "invoices_receipt_number_unique" UNIQUE("receipt_number");