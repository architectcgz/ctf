ALTER TABLE public.contests
    ALTER COLUMN start_time TYPE timestamp without time zone USING start_time AT TIME ZONE 'UTC',
    ALTER COLUMN end_time TYPE timestamp without time zone USING end_time AT TIME ZONE 'UTC',
    ALTER COLUMN freeze_time TYPE timestamp without time zone USING freeze_time AT TIME ZONE 'UTC';
