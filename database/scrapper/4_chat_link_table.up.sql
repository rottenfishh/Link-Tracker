CREATE TABLE IF NOT EXISTS chat_link(
    chat_id int8 NOT NULL REFERENCES chats(id),
    link_id int8 NOT NULL REFERENCES links(id),
    primary key (chat_id, link_id)
)