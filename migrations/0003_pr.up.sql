CREATE TABLE IF NOT EXISTS pull_requests (
                                             id TEXT PRIMARY KEY,
                                             name TEXT NOT NULL,
                                             author_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')) DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMPTZ
    );

CREATE TABLE IF NOT EXISTS pr_reviewers (
                                            pr_id TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (pr_id, user_id)
    );