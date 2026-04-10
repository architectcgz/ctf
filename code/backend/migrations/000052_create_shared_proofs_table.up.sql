CREATE TABLE IF NOT EXISTS shared_proofs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    challenge_id BIGINT NOT NULL,
    contest_id BIGINT NULL,
    instance_id BIGINT NOT NULL,
    proof_hash VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_shared_proofs_hash ON shared_proofs(proof_hash);
CREATE INDEX IF NOT EXISTS idx_shared_proofs_user_challenge_contest_status ON shared_proofs(user_id, challenge_id, contest_id, status);
CREATE INDEX IF NOT EXISTS idx_shared_proofs_instance_id ON shared_proofs(instance_id);
CREATE INDEX IF NOT EXISTS idx_shared_proofs_expires_at ON shared_proofs(expires_at);
