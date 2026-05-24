CREATE OR REPLACE FUNCTION public.store_batch_urls(
    long_urls text[],
    short_urls text[],
    p_user_id bigint
)
RETURNS TABLE (
    short_url text,
    long_url text,
    stored_short_url text,
    retry boolean
)
LANGUAGE sql
AS $function$
    WITH input(long_url, short_url) AS (
        SELECT *
        FROM unnest(long_urls, short_urls) AS t(long_url, short_url)
    ),
    locked AS (
        SELECT u.long_url, u.short_url
        FROM urls u
        JOIN input i ON u.deleted_at_utc IS NULL
            AND (u.long_url = i.long_url OR u.short_url = i.short_url)
        FOR UPDATE OF u
    ),
    marked AS (
        SELECT
            i.long_url,
            i.short_url,
            (
                SELECT l.short_url
                FROM locked l
                WHERE l.long_url = i.long_url
                LIMIT 1
            ) AS stored_short_url,
            NOT EXISTS (
                SELECT 1
                FROM locked l
                WHERE l.long_url = i.long_url OR l.short_url = i.short_url
            ) AS should_insert,
            EXISTS (
                SELECT 1
                FROM locked l
                WHERE l.short_url = i.short_url
                  AND l.long_url <> i.long_url
            ) AS retry
        FROM input i
    ),
    inserted AS (
        INSERT INTO urls (long_url, short_url, user_id)
        SELECT m.long_url, m.short_url, p_user_id
        FROM marked m
        WHERE m.should_insert AND NOT m.retry
        ON CONFLICT (long_url) WHERE (deleted_at_utc IS NULL)
        DO UPDATE SET long_url = EXCLUDED.long_url
    )
    SELECT
        m.short_url::text,
        m.long_url::text,
        COALESCE(m.stored_short_url, u.short_url)::text,
        m.retry
    FROM marked m
    LEFT JOIN urls u ON u.long_url = m.long_url
        AND u.deleted_at_utc IS NULL
        AND m.stored_short_url IS NULL
        AND m.should_insert
    WHERE m.retry
       OR m.stored_short_url IS NOT NULL
       OR (m.should_insert AND u.short_url IS NOT NULL AND u.short_url <> m.short_url);
$function$;
