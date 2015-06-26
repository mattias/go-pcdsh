-- +migrate Up
CREATE TABLE logs (
  id int not null auto_increment,
  time timestamp not null,
  name text not null,
  refid int,
  participantid int,
  primary key (id)
);

-- +migrate Down
DROP TABLE logs;
