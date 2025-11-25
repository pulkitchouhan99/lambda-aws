CREATE TABLE IF NOT EXISTS interventions_projection (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    patient_id TEXT NOT NULL,
    screening_id TEXT NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    priority TEXT DEFAULT 'medium',
    created_by TEXT NOT NULL,
    assigned_to TEXT,
    assigned_team TEXT,
    due_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    linked_task_id TEXT,
    referral_reasons TEXT[],
    problems TEXT[],
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_interventions_projection_tenant_id ON interventions_projection(tenant_id);
CREATE INDEX IF NOT EXISTS idx_interventions_projection_patient_id ON interventions_projection(patient_id);
CREATE INDEX IF NOT EXISTS idx_interventions_projection_screening_id ON interventions_projection(screening_id);
CREATE INDEX IF NOT EXISTS idx_interventions_projection_status ON interventions_projection(status);
CREATE INDEX IF NOT EXISTS idx_interventions_projection_created_by ON interventions_projection(created_by);
CREATE INDEX IF NOT EXISTS idx_interventions_projection_type ON interventions_projection(type);
