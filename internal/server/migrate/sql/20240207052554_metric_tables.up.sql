create table gauges
(
 name  varchar(50)      not null
  constraint gauges_name
   primary key,
 value double precision not null
);

create table counters
(
 name  varchar(50) not null
  constraint counters_name
   primary key,
 value bigint      not null
);

