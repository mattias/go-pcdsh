-- +migrate Up
CREATE TABLE log_attributes (
  id int not null auto_increment,
  log_id int not null,
  `key` varchar(255),
  value varchar(255),
  primary key (id),
  index (`key`),
  index (value ),
  foreign key (log_id) references logs(id)
);

-- +migrate Down
DROP TABLE log_attributes;
