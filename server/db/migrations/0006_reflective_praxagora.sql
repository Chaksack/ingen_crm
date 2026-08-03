ALTER TABLE "business_bank_accounts" DISABLE ROW LEVEL SECURITY;--> statement-breakpoint
ALTER TABLE "businesses" DISABLE ROW LEVEL SECURITY;--> statement-breakpoint
ALTER TABLE "financial_institutions" DISABLE ROW LEVEL SECURITY;--> statement-breakpoint
ALTER TABLE "vendors" DISABLE ROW LEVEL SECURITY;--> statement-breakpoint
DROP TABLE "business_bank_accounts" CASCADE;--> statement-breakpoint
DROP TABLE "businesses" CASCADE;--> statement-breakpoint
DROP TABLE "financial_institutions" CASCADE;--> statement-breakpoint
-- "vendors" CASCADE also drops expenses_vendor_id_vendors_id_fk, so the column drop
-- below no longer needs (and must not repeat) an explicit DROP CONSTRAINT first.
DROP TABLE "vendors" CASCADE;--> statement-breakpoint
ALTER TABLE "expenses" DROP COLUMN "vendor_id";