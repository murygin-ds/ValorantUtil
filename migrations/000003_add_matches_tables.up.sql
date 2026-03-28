-- Basic match info (immutable once stored)
CREATE TABLE matches (
    match_id              TEXT PRIMARY KEY,
    map_id                TEXT NOT NULL DEFAULT '',
    queue_id              TEXT NOT NULL DEFAULT '',
    game_start_time       TIMESTAMPTZ NOT NULL,
    game_length_ms        BIGINT NOT NULL DEFAULT 0,
    team_red_won          BOOLEAN,
    team_red_rounds_won   INT NOT NULL DEFAULT 0,
    team_blue_rounds_won  INT NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Per-player stats for each match
CREATE TABLE match_players (
    id           BIGSERIAL PRIMARY KEY,
    match_id     TEXT NOT NULL REFERENCES matches(match_id) ON DELETE CASCADE,
    puuid        TEXT NOT NULL,
    team_id      TEXT NOT NULL DEFAULT '',
    character_id TEXT NOT NULL DEFAULT '',
    score        INT NOT NULL DEFAULT 0,
    kills        INT NOT NULL DEFAULT 0,
    deaths       INT NOT NULL DEFAULT 0,
    assists      INT NOT NULL DEFAULT 0,
    UNIQUE (match_id, puuid)
);

CREATE INDEX ON match_players (puuid);

-- Kill events per match (for duel stats)
CREATE TABLE match_kills (
    id           BIGSERIAL PRIMARY KEY,
    match_id     TEXT NOT NULL REFERENCES matches(match_id) ON DELETE CASCADE,
    round        INT NOT NULL DEFAULT 0,
    killer_puuid TEXT NOT NULL DEFAULT '',
    victim_puuid TEXT NOT NULL DEFAULT '',
    assistants   TEXT[] NOT NULL DEFAULT '{}'
);

CREATE INDEX ON match_kills (match_id);
