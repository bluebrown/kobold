-- this is for the web api. So that it can display the channels embedded in the
-- pipeline json object
create view if not exists pipeline_list_item as
select
  pipeline.*,
  json_group_array(subscription.channel_name) as channels
from pipeline
left join subscription on subscription.pipeline_name = pipeline.name
group by pipeline.name;

-- the run view is like a counter part to the task group view. Its primarly
-- use case is to show the actual run information of the task groups returned by
-- the task group view. runs with a fingerprint have been executed and will
-- never change. However, pending runs, have not been executed yet, and they can
-- still change until they are executed
create view if not exists run as
select
  ifnull(task_group_fingerprint, '') as fingerprint,
  repo_uri,
  dest_branch,
  post_hook_name as post_hook,
  status,
  max(timestamp) as timestamp,
  max(warnings) as warnings,
  max(failure_reason) as error,
  json_group_array(json(msgs)) as msgs
from task
group by
  task_group_fingerprint,
  repo_uri,
  dest_branch,
  post_hook_name,
  status;

