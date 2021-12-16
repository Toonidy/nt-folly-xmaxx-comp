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

CREATE TABLE competitions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	request_id UUID REFERENCES nt_api_team_log_requests (id),
	status TEXT NOT NULL CHECK (status IN ('DRAFT', 'STARTED', 'FINISHED', 'FAILED')) DEFAULT 'DRAFT',
	multiplier INT NOT NULL CHECK (multiplier IN (1, 2, 4, 8)) DEFAULT 1,
	grind_rewards INT[5] NOT NULL,
	point_rewards INT[5] NOT NULL,
	speed_rewards INT[5] NOT NULL,
	accuracy_rewards INT[5] NOT NULL,
	from_at TIMESTAMPTZ NOT NULL,
	to_at TIMESTAMPTZ NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX competitions_time_range_idx ON competitions (
	from_at,
	to_at,
	from_at ASC NULLS FIRST
);

CREATE MATERIALIZED VIEW competition_results AS
SELECT r.competition_id,
	r.user_id,
	r.grind,
	rank() OVER g grind_rank,
	coalesce(r.grind_rewards[rank() OVER g] * r.multiplier, 0) AS grind_reward,
	r.accuracy,
	rank() OVER a AS accuracy_rank,
	coalesce(r.accuracy_rewards[rank() OVER a] * r.multiplier, 0) AS accuracy_reward,
	r.speed,
	rank() OVER s AS speed_rank,
	coalesce(r.speed_rewards[rank() OVER s] * r.multiplier, 0) AS speed_reward,
	r.point,
	rank() OVER p AS point_rank,
	coalesce(r.point_rewards[rank() OVER p] * r.multiplier, 0) AS point_reward
FROM (
	SELECT ur.user_id,
		c.id AS competition_id,
		c.multiplier,
		c.grind_rewards,
		c.accuracy_rewards,
		c.speed_rewards,
		c.point_rewards,
		ur.played AS grind,
		((1.0 - (ur.errs / ur.typed::decimal)) * 100.0) AS accuracy,
		(ur.typed / 5.0 / (ur.secs / 60.0)) AS speed,
		ROUND(ur.played
			* (
				(100.0 + ((ur.typed / 5.0 / (ur.secs / 60.0)) / 2.0))
					* (1.0 - (ur.errs / ur.typed::decimal))
			)
		) AS point
	FROM competitions c 
		INNER JOIN user_records ur ON ur.request_id = c.request_id 
		INNER JOIN users u ON u.id = ur.user_id AND u.status != 'DISQUALIFIED'
) r
WINDOW g as (PARTITION BY r.competition_id ORDER BY r.grind DESC),
	a AS (PARTITION BY r.competition_id ORDER BY r.accuracy DESC, r.grind DESC),
	s AS (PARTITION BY r.competition_id ORDER BY r.speed DESC, r.grind DESC),
	p AS (PARTITION BY r.competition_id ORDER BY r.point DESC, r.grind DESC);

CREATE UNIQUE INDEX ON competition_results (competition_id, user_id);
