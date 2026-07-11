-- Service core (PRD §5.3): SLA timers, queue routing, knowledge base.

-- SLA policy: one per organization for v1 (per-priority/per-queue matrices
-- are future work — see README known gaps).
CREATE TABLE IF NOT EXISTS sla_policies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
  first_response_minutes INT NOT NULL DEFAULT 60,
  resolution_minutes INT NOT NULL DEFAULT 480,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO sla_policies (organization_id, first_response_minutes, resolution_minutes)
SELECT id, 60, 480 FROM organizations
ON CONFLICT (organization_id) DO NOTHING;

-- Case SLA tracking. paused_at/paused_seconds implement pause conditions:
-- while a case sits in a "pausing" status (see service.pausingStatuses),
-- both timers are frozen.
ALTER TABLE cases
  ADD COLUMN IF NOT EXISTS first_response_due_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS first_response_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS resolution_due_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS paused_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS paused_seconds INT NOT NULL DEFAULT 0;

-- Queue routing: manual (default, unchanged behavior) / round_robin / capacity.
ALTER TABLE queues
  ADD COLUMN IF NOT EXISTS routing_strategy TEXT NOT NULL DEFAULT 'manual',
  ADD COLUMN IF NOT EXISTS last_assigned_user_id UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS queue_members (
  queue_id UUID NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  PRIMARY KEY (queue_id, user_id)
);

-- Knowledge base.
CREATE TABLE IF NOT EXISTS kb_articles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  body TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'draft',
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS kb_articles_org_idx ON kb_articles(organization_id);
