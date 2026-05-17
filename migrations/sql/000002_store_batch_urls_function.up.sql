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
    marked AS (
        SELECT
            i.long_url,
            i.short_url,
            (
                SELECT u.short_url
                FROM urls u
                WHERE u.long_url = i.long_url AND u.deleted_at_utc IS NULL
                LIMIT 1
            ) AS stored_short_url,
            NOT EXISTS (
                SELECT 1
                FROM urls u
                WHERE (u.long_url = i.long_url OR u.short_url = i.short_url)
                  AND u.deleted_at_utc IS NULL
            ) AS should_insert,
            EXISTS (
                SELECT 1
                FROM urls u
                WHERE u.short_url = i.short_url
                  AND u.long_url <> i.long_url
                  AND u.deleted_at_utc IS NULL
            ) AS retry
        FROM input i
    ),
    inserted AS (
        INSERT INTO urls (long_url, short_url, user_id)
        SELECT m.long_url, m.short_url, p_user_id
        FROM marked m
        WHERE m.should_insert
    )
    SELECT
        m.short_url::text,
        m.long_url::text,
        m.stored_short_url::text,
        m.retry
    FROM marked m
    WHERE m.retry OR m.stored_short_url IS NOT NULL;
$function$;
