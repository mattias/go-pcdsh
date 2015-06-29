-- +migrate Up
CREATE TABLE sessions (
  id int not null auto_increment,
  start_log_id int not null,
  end_log_id int not null,
  start_time timestamp not null,
  end_time timestamp not null,
  log_count int,
  primary key (id),
  foreign key (start_log_id) references logs(id),
  foreign key (end_log_id) references logs(id),
  index (start_time),
  index (end_time)
);

-- +migrate Down
DROP TABLE sessions;
