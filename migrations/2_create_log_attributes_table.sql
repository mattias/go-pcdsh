-- +migrate Up
CREATE TABLE log_attributes (
  id int not null auto_increment,
  log_id int not null,
  `key` text,
  value text,
  primary key (id),
  foreign key (log_id) references logs(id)
);

-- +migrate Down
DROP TABLE log_attributes;
