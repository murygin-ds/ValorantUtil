CREATE TABLE assets (
    id               BIGSERIAL PRIMARY KEY,

    type_id          TEXT NOT NULL,
    item_id          TEXT NOT NULL,
    quantity         INT NOT NULL DEFAULT 0,
    price            INT,

    display_name_ru  TEXT,
    display_name_en  TEXT,
    display_icon_url TEXT,
    stream_video_url TEXT,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (type_id, item_id)
);

CREATE INDEX ON assets (item_id);