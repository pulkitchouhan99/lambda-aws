CREATE TABLE IF NOT EXISTS interventions (
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

CREATE INDEX idx_interventions_tenant_id ON interventions(tenant_id);
CREATE INDEX idx_interventions_patient_id ON interventions(patient_id);
CREATE INDEX idx_interventions_screening_id ON interventions(screening_id);
CREATE INDEX idx_interventions_status ON interventions(status);
CREATE INDEX idx_interventions_created_by ON interventions(created_by);
