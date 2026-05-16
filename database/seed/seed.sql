INSERT INTO chats(id)
SELECT generate_series(1,1000) +10000 as id;

INSERT INTO links(link, domain, last_updated, formatted_link, title)
SELECT 'https://www.github.com/user/' || repo, 'github', now(), 'api.github.com/user/' || repo, repo
FROM generate_series(1,100_000) as repo;

WITH numbered_links AS (
    SELECT id, row_number() OVER (ORDER BY id) as rn
    FROM links
)
INSERT INTO chat_link (chat_id, link_id)
SELECT
    c.id,
    nl.id
FROM chats c
     JOIN numbered_links nl
          ON nl.rn BETWEEN ((c.id - 10001) * 100 + 1)
              AND ((c.id - 10001) * 100 + 100);