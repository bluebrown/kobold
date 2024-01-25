-- a channel gives a name to an incoming data stream. if a decoder is provded,
-- it will be used before the data is stored
create table if not exists channel (
  name    text not null primary key,
  decoder_name text
);

-- a decoder is a starlark script that should normalize the incoming data into a
-- format that can be used by the pipeline
create table if not exists decoder (
  name text not null primary key,
  script blob
);

-- a post hook is a starlark script that will be run after a pipeline has been
-- run. it it can be used to perform additional actions such opening a pull
-- request
create table if not exists post_hook (
  name text not null primary key,
  script blob
);

-- a pipeline respresents a set of mutations against a git repository it a
-- function over input data
create table if not exists pipeline (
  name        text not null primary key,
  repo_uri    text not null,
  dest_branch text,
  post_hook_name text
);

-- the subscription links a pipeline to a channel- The intention is that
-- everytime a message is received on the channel, the pipeline will be run with
-- the decoded message as input
create table if not exists subscription (
  pipeline_name text not null,
  channel_name  text not null,
  primary key (pipeline_name, channel_name)
);

-- a task represents a single mutation against a git repository it is the
-- combination of a pipeline and concrete input data
create table if not exists task (
  id             text not null primary key default (uuid()),
  msgs           text not null,
  repo_uri       text not null,
  dest_branch    text,
  post_hook_name text,
  status         text not null check (status in ('pending', 'running', 'success', 'failure')) default 'pending',
  timestamp      text not null,
  warnings       text,
  failure_reason text,
  task_group_fingerprint text check (status == 'pending' or task_group_fingerprint is not null)
);

-- task groups are used to coordinate the execution of tasks. since pipelines
-- operate on a single repo, we can group tasks by repo. This is only a view in
-- order to help sqlc to generate the right types. The result of selecting this
-- view is highly dynamic. Selecting twice will probably never return the same
-- result. The way to correlate tasks later is by looking at the fingerprint.
-- all tasks with the same fingerprint, have been executed as group called a run
create view if not exists task_group as
select
  sha1(group_concat(id)) as fingerprint,
  repo_uri,
  dest_branch,
  ph.script as post_hook,
  json_group_array(id) as task_ids,
  json_group_array(json(msgs)) as msgs
from task
left join post_hook ph on task.post_hook_name = ph.name
where status = 'pending'
group by repo_uri, dest_branch, post_hook_name;
