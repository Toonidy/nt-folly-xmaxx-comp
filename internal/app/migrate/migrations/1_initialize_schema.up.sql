CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

/********************
*  Nitro Type Logs  *
********************/

CREATE TABLE nt_api_team_logs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	log_data JSON NOT NULL,
	hash BYTEA NOT NULL UNIQUE,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE nt_api_team_log_requests (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	api_team_log_id UUID NOT NULL REFERENCES nt_api_team_logs (id),
	response_type TEXT NOT NULL CHECK (response_type IN ('ERROR', 'CACHE', 'NEW')),	
	description TEXT NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
