-- +migrate Up
CREATE TABLE sessions (
  id int not null auto_increment,
  start_log_id int not null,
  end_log_id int not null,
  start_time timestamp not null,
  end_time timestamp not null,
  track_id int,
  log_count int,
  valid int,
  primary key (id),
  foreign key (start_log_id) references logs(id),
  foreign key (end_log_id) references logs(id),
  index (start_time),
  index (end_time),
  index (valid)
);

-- +migrate Down
DROP TABLE sessions;
