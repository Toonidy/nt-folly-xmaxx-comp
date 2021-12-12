CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

/********************
*  Nitro Type Logs  *
********************/

CREATE TABLE nt_api_team_logs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	hash BYTEA NOT NULL UNIQUE,
	log_data JSON NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE nt_api_team_log_requests (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	prev_id UUID REFERENCES nt_api_team_log_requests (id),
	api_team_log_id UUID NOT NULL REFERENCES nt_api_team_logs (id),
	response_type TEXT NOT NULL CHECK (response_type IN ('ERROR', 'CACHE', 'NEW')),	
	description TEXT NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

/*********************
*  Competition Data  *
*********************/

CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	reference_id INT NOT NULL UNIQUE,
	username TEXT NOT NULL UNIQUE,
	display_name TEXT NOT NULL,
	membership_type TEXT NOT NULL CHECK (membership_type IN ('BASIC', 'GOLD')),
	status TEXT NOT NULL CHECK (status IN ('NEW', 'ACTIVE', 'DISQUALIFIED')),

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_records (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	request_id UUID NOT NULL REFERENCES nt_api_team_log_requests (id),
	user_id UUID NOT NULL REFERENCES users (id),

	played INT NOT NULL,
	typed INT NOT NULL,
	errs INT NOT NULL,
	secs INT NOT NULL,
	from_at TIMESTAMPTZ NOT NULL,
	to_at TIMESTAMPTZ NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	UNIQUE (request_id, user_id)
);

CREATE TABLE competition_rewards (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	result_id UUID REFERENCES nt_api_team_log_requests (id),
	status TEXT NOT NULL CHECK (status IN ('DRAFT', 'STARTED', 'FINISHED', 'FAILED')) DEFAULT 'DRAFT',
	multiplier INT NOT NULL CHECK (multiplier IN (1, 2, 4, 8)) DEFAULT 1,
	point_rewards INT[5] NOT NULL,
	speed_rewards INT[5] NOT NULL,
	accuracy_rewards INT[5] NOT NULL,
	from_at TIMESTAMPTZ NOT NULL,
	to_at TIMESTAMPTZ NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
