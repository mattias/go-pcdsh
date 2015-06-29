-- +migrate Up
CREATE TABLE logs (
  id int not null auto_increment,
  `index` int not null,
  time timestamp not null,
  name varchar(255) not null,
  refid int,
  participantid int,
  primary key (id),
  index (`index`),
  index (time),
  index (name),
  index (refid)
);

-- +migrate Down
DROP TABLE logs;
