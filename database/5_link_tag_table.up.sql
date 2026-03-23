CREATE TABLE IF NOT EXISTS link_tag(
    link_id int8 NOT NULL REFERENCES links(id),
    tag_id int8 NOT NULL REFERENCES tags(id),
    primary key (link_id, tag_id)
)