-- Workflow/automation engine (PRD §5.1, E9): trigger -> conditions -> actions,
-- "the Power Automate-lite." Conditions/actions are stored as JSONB since
-- they're always loaded wholesale with their workflow, never queried
-- individually — a relational schema here would just be joins with no
-- indexing benefit.

CREATE TABLE IF NOT EXISTS workflows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  entity TEXT NOT NULL,
  trigger_event TEXT NOT NULL, -- 'created' | 'updated' | 'sla_breach'
  conditions JSONB NOT NULL DEFAULT '[]',
  actions JSONB NOT NULL DEFAULT '[]',
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS workflows_lookup_idx
  ON workflows(organization_id, entity, trigger_event)
  WHERE is_active;

CREATE TABLE IF NOT EXISTS workflow_runs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  record_id UUID NOT NULL,
  status TEXT NOT NULL, -- 'success' | 'error'
  detail TEXT NOT NULL DEFAULT '',
  ran_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS workflow_runs_workflow_idx ON workflow_runs(workflow_id, ran_at DESC);

-- Minimal in-app notifications: the first (and so far only) action target.
-- Push/email fan-out (PRD §7.5) is separate, later work.
CREATE TABLE IF NOT EXISTS notifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  body TEXT NOT NULL DEFAULT '',
  link TEXT,
  is_read BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notifications_user_idx ON notifications(user_id, created_at DESC);

-- Dedup flags so the SLA-breach background scanner fires a given breach
-- exactly once per case, not on every scan tick.
ALTER TABLE cases
  ADD COLUMN IF NOT EXISTS first_response_breach_notified_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS resolution_breach_notified_at TIMESTAMPTZ;
