create table certmagic_locks (
  key string primary key,
  expires timestamp not null default current_timestamp()
);

create table certmagic_values (
  key string primary key,
  value bytes not null,
  created timestamp not null default current_timestamp(),
  updated timestamp not null default current_timestamp()
);
