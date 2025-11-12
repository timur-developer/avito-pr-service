CREATE TABLE IF NOT EXISTS teams (
                                     name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
                                     user_id TEXT PRIMARY KEY,
                                     username TEXT NOT NULL,
                                     team_name TEXT NOT NULL REFERENCES teams(name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true
    );

CREATE TABLE IF NOT EXISTS pull_requests (
                                             id TEXT PRIMARY KEY,
                                             name TEXT NOT NULL,
                                             author_id TEXT NOT NULL REFERENCES users(user_id),
    status TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')) DEFAULT 'OPEN',
    assigned_reviewers TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMPTZ
    );

CREATE INDEX IF NOT EXISTS idx_users_team_active ON users(team_name) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewer ON pull_requests USING GIN (assigned_reviewers);