# Kobold

> [!NOTE]  
> If you are migrating from v.0.2 to v.0.3, you can use the `confix` command to
> migrate your config. See the [ConFix](#confix) section for more details.

Update container image references in git repositories, based on recieved events.

Kobold is meant to be used as a companion to other gitops tools, like ArgoCD or
Flux. It keeps image references in your gitops repos up to date, while
delegating the actual deployment to your gitops tool of choice.

Typically you would configure your container registry to send events to kobolds
webhook endpoint, whenever a new image is pushed.

## Webhook

In servermode Kobold exposes a webhook endpoint at
`$KOBOLD_ADDR_WEBHOOK/events?chan=<channel>`. The channel is used to identify
the pipelines to run. It accepts any content in the body since it is the
[channels decoder's](#decoders) responsibility to parse the body.

For example below is a request to a channel using the plain lines decoder.

```text
POST /events?chan=plain HTTP/1.1
Host: kobold.myorg.io:8080

docker.io/library/busybox:v1.4@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248
```

## Matching Rules

Kobold uses special inline comments in yaml files, to identify image references
that should be updated. It will match the reference from the incoming event,
against the rules in the comment, and update the reference if a match is found.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - name: my-app
    image: my.org/amazing/app # kobold: tag: ^1; type: semver
```

The comments parsed by kobold have the following format:

```console
# kobold: tag: <tag-name>; type: <tag-type>
```

The tag-type can be either `exact`, `semver`, or `regex`, and specifies how
kobold should interpret the tag-name.

For example, if tag-type is semver, the tag-name can include common semantic
versioning semantics, such as ^1 to denote that any tag between v1 and v2 should
be matched (not including v2).

Note that the regex type will add `^` and `$` to the tag-name, to ensure the
regex matches the entire tag, so dont include them in the tag-name.

## Configuration

Kobold is configured by setting up named channels, and pipelines to run whenever
a message is recieved on a channel. The pipeline will clone the repo, and run a
krm filter, before potentially pushing the changes back to the repo.

In the below example, the `example` pipeline is run, whenever a message is
recieved on the `a` or `b` channel.

```toml
version = "2"

[[channel]]
name = "a"

[[channel]]
name = "b"

[[pipeline]]
name = "example"
repo_uri = "git@github.com:bluebrown/example.git@main"
channels = ["a", "b"]
```

<span id="decoders"></span>

Messages received on a channel have to be decoded, before they can be processed
by the pipeline. By default, each line in a message is treated as a fully
qualified image reference. You can change this behaviour by providing a decoder
in the channel config. See the [builtin](#builtins) section for more details.

```toml
[[channel]]
name = "example"
decoder = "builtin.distribution@v1"
```

Optionally, you can set a destination branch, to push the changes to. If you
dont set a destination branch, the changes will be pushed to the source branch.

```toml
[[pipeline]]
name = "example"
dest_branch = "release"
```

If you want to perform an action after the changes have been pushed to git, you
can attach a post hook to the pipeline. See the [builtin](#builtins) section for
more details.

```toml
[[pipeline]]
name = "example"
post_hook = "builtin.github-pr@v1"
```

### Scoping

Pipelines are scoped through the following URI format. Note the package is
optional.

```text
<repo>[.git]@<ref>[/<pkg>]
```

If you want to scope a pipline beyond a sub directory (package), you can place a
.krmignore file at the package root, ignoring parts of the package.

> ignoreFilesMatcher handles `.krmignore` files, which allows for ignoring files
> or folders in a package. The format of this file is a subset of the gitignore
> format, with recursive patterns (like a/**/c) not supported. If a file or
> folder matches any of the patterns in the .krmignore file for the package, it
> will be excluded.

For example, you have a helm like below, and want to ignore its template
directory:

```text
. (package root)
├── charts
│   ├── my-chart
│   └── .krmignore
└── pod.yaml
```

Then you would place the following in the `charts/.krmignore` file:

```text
/*/charts/
/*/templates/
```

This works because the presense of a `.krmignore` makes the direcotry,
containing the file, a sub package, and kobold will recurse into it.

### Builtins

There are a few builtin decoders and post hooks, to support some common use
cases. For example decoding events from a oci distribution registry, or opening
a pull request on github.

The starlark code for the builtins can be reviewed in the [builtin](./builtin)
directory.

#### Decoder

##### `builtin.lines@v1`

The lines reads lines of image references:

```text
docker.io/bluebrown/busybox:v1.3@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248
test.azurecr.io/nginx:v1.1.1@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248
```

##### `builtin.distribution@v1`

The distribution decoder parses a json message using the schema defined in
<https://github.com/distribution/distribution>.

It understands single events:

```json
{
  "request": {"host": "example.azurecr.io"},
  "target": {
    "digest": "sha256:xxxxd5c8786bb9e621a45ece0dbxxxx1cdc624ad20da9fe62e9d25490f33xxxx",
    "repository": "bluebrown/busybox",
    "tag": "v1"
  }
}
```

Or wrapped in an envelope:

```json
{"events": [{"..."}]}
```

##### `builtin.dockerhub@v1`

The dockerhub decoder parses a json message using the schema defined in
<https://docs.docker.com/docker-hub/webhooks/#push-event>.

```json
{
  "push_data": {
    "tag": "v1"
  },
  "repository": {
    "repo_name": "bluebrown/busybox"
  }
}
```

> [!WARNING]  
> Since dockerhub does not set the image digest in the payload, kobold will
> update only the tag of matched image references. This will not work as
> expected when pusing the same tag repeatedly.

#### Post Hooks

##### `builtin.github-pr@v1`

Performs a github pull request. It requires the `GITHUB_TOKEN` environment
variable to be set.

#### `builtin.ado-pr@v1`

Performs a azure devops pull request. It requires the `ADO_USR` and `ADO_PAT`
environment variables to be set.

##### `builtin.gitea-pr@v1`

Performs a gitea pull request. It requires the`GITEA_HOST` and
`GITEA_AUTH_HEADER` environment variable to be set.

### Extending Kobold

Kobold is designed to be extended. You can write your own decoders and post
hooks.

For example, below is the builtin lines decoder. It simply splits the message
recieved on the channel by newlines, and returns the resulting list. Treating
each line as a fully qalified image reference.

```toml
[[decoder]]
name = "builtin.lines@v1"
script = """
def main(input):
    return input.split("\n")
"""
```

This post hook does nothing, but print the message to stdout.

```toml
[[post_hook]]
name = "builtin.print@v1"
script = """
def main(repo, src_branch, dest_branch, title, body, changes, warnings):
    print(repo, src_branch, dest_branch, title, body, changes, warnings)
"""
```

This allows to integrate with any event producer and any git provider, since
producer and provider specific logic can be implemented in starlark.

### Git

Kobold uses the git command line tool to interact with git. That means you can
configure the underlying git client directly and according to gits
documentation. For example you can mount a file to `/.gitconfig` in the kobold
container, or set git specific environment variables.

## Cook Book

This section showcases some common use cases.

### SSH

To use ssh, you can mount a directory to `/.ssh` in the kobold container.

The directory should, at least contain, a known_hosts file and a default
identity file. The default identity file is usually named id_rsa. If more than
one key is required, you can place a config file in the .ssh directory, and
configure the identity file for each host.

```bash
# this dir will be mounted
mkdir -p .ssh
# scan your git providers host key
ssh-keyscan github.com > .ssh/known_hosts
# generate a key (dont forget to add it to your git provider)
ssh-keygen -t ed25519 -f .ssh/id_ed25519 -N ''
# set the permission to the kobold user
chown -R 65532:65532 .ssh
# run kobold
docker run -v "$(pwd)/.ssh:/.ssh" ...
```

### Basic Auth

You can provide the credentials in the git uri, but this is not recommended,
since it will be stored in the kobold database.

Alternatively, you can configure gits credentails helper, to use a file as
backend. For example via `/.gitconfig`:

```conf
[credential]
    helper = store --file /.git-credentials
```

When doing this, you can mount a credentials file to `/.git-credentials` in the
kobold container.

Below is an example entry in the credentials file:

```text
https://bob:s3cre7@mygithost
```

### Pull Requests

For a pull request setup, you want to set the destionation branch to something
other than the source branch. Then you can run a post hook, opening the pull
request. There is already, amongst others, a [builtin](#post-hooks) post hook
for github, using the `GITHUB_TOKEN` environment variable, to authenticate.

```toml
[[pipeline]]
name = "my-github-pr"
channels = ["example"]
repo_uri = "git@github.com:bluebrown/foobar.git@main/manifests"
dest_branch = "kobold"
post_hook = "builtin.github-pr@v1"
```

### Environment Promotion

The below example uses package scoping, to perform different actions based on
the environment. For stage, the updated image ref is directly applied to the
cluster. For prod, a pull request is opened, and the image ref is updated after
the pull request is merged.

```toml
[[pipeline]]
name = "stage"
channels = ["distribution"]
repo_uri = "http://gitea/dev/test.git@main/stage"

[[pipeline]]
name = "prod"
channels = ["distribution"]
repo_uri = "http://gitea/dev/test.git@main/prod"
dest_branch = "release"
post_hook = "builtin.gitea-pr@v1"
```

## Deployment

### Kubernetes

The kobold manifests come with a few secrets and config maps intended to be
overwritten by the user. They are mostly empty by default, but mounted to the
right location, so any values provided will be picked up.

| Resources                  | Description                                            | Mount Location                 |
| -------------------------- | ------------------------------------------------------ | -------------------------------|
| configmap/kobold-confd     | multiple .toml files containing kobold configs         | `/etc/kobold/conf.d`           |
| configmap/kobold-gitconfig | `.gitconfig` key containing the git config             | `/etc/kobold/.gitconfig`       |
| secret/kobold-gitcreds     | `.git-credentials` key containing the git credentials  | `/etc/kobold/.git-credentials` |
| secret/kobold-ssh          | arbitrary keys representing the contents of `.ssh`     | `/etc/kobold/.ssh`             |
| secret/kobold-env          | arbitrary environment variables                        | *used as env vars*             |

You can use kustomize to merge your own configs into the kobold manifests. For
example:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- https://github.com/bluebrown/kobold/releases/latest/download/kobold-manifests.yaml
configMapGenerator:
- name: kobold-confd
  behavior: merge
  files:
  - team-a.toml
  - team-b.toml
secretGenerator:
- name: kobold-env
  behavior: merge
  literals:
  - GITHUB_TOKEN=xxx
```

## Metrics

Kobold exposes prometheus metrics on port 8080. The metrics are exposed in the
`$KOBOLD_ADDR_API/metrics` endpoint. The metrics are prefixed with `kobold_`.

```python
# HELP kobold_git_fetch number of git fetches
# TYPE kobold_git_fetch counter
kobold_git_fetch{repo="git@github.com:bluebrown/foobar"} 3
# HELP kobold_git_push number of git pushes
# TYPE kobold_git_push counter
kobold_git_push{repo="git@github.com:bluebrown/foobar"} 1
# HELP kobold_image_seen number of images seen
# TYPE kobold_image_seen counter
kobold_image_seen{ref="docker.io/bluebrown/busybox"} 3
kobold_image_seen{ref="docker.io/bluebrown/nginx"} 2
# HELP kobold_msg_recv number of messages received
# TYPE kobold_msg_recv counter
kobold_msg_recv{channel="dockerhub",rejected="false"} 5
# HELP kobold_run_active number of active runs
# TYPE kobold_run_active gauge
kobold_run_active 0
# HELP kobold_run_status run status (task groups)
# TYPE kobold_run_status counter
kobold_run_status{repo="git@github.com:bluebrown/foobar",status="success"} 2
kobold_run_status{repo="git@github.com:bluebrown/foobar",status="failure"} 1
```

## Web API

In server mode, kobold exposes a json over http api. The api docs are available
at `$KOBOLD_ADDR_API/api/docs/`. Note the trailing slash.

## SQL

Kobold uses sqlite3 as a database. You can interact with the database using the
sqlite cli, built into the kobold container.

```bash
# inspect the db of a running kobold server
docker exec -ti my-kobold sqlite3 ~/.config/kobold.db
# inspect local db files
docker run -ti -v "$HOME/.config/kobold:/tmp" bluebrown/kobold sqlite3 /tmp/kobold.db
```

If you want to use your own sqlite binary, make sure that you have the `uuid`
and `sha1` [extensions](https://sqlite.org/loadext.html) enabled.

## Binaries

```bash
GOBIN="$(pwd)/bin" go install ./cmd/...
```

### Server

The server is the main kobold binary. It runs a webhook server, decoding and
emitting events received on an http endpoint. Decoded events are sheduled and
processed in groups after a debounce period. This ensures the minimum amount of
git operations are performed.

```bash
bin/server
```

### Command Line Interface

The CLI, reads messages from stdin. One message per line. The messages are
treated according to the kobold.toml. The same way they would in the server
binary.

The difference is, that the cli will not wrap the pool in the scheduler, meaning
the task will run immedaitly after grouping them. The cli will exit after all
messages have been processed.

```bash
bin/cli -channel default -handler print < testdata/events.txt
```

### Image Reference Updater

The `image-ref-updater` command is kobolds business logic, as a standalone krm
filter, that can be used by other tools like kpt or kustomize.

It will update image references in the provided resources list. Image references
to match against, are read from the `functionConfig`, which has the following
format:

```yaml
apiVersion: kobold/v1alpha3
kind: List
metadata:
  name: my-fn-config
items:
- docker.io/bluebrown/busybox:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248
- test.azurecr.io/nginx:v1@sha256:993518ca49ede3c4e751fe799837ede16e60bc410452e3922602ebceda9b4c73
```

### Git Read/Writer

The `grw` command is a git read/writer. It reads from a source repository and
emits the resources to stdout. Optionally, by using the `-a` flag, it sets
tracking annotation, to improve the write performance, by preventing multiple
clones.

This example reads from git, pipes to a krm filter, and writes back to git:

```bash
bin/grw -a source 'git@github.com:bluebrown/foobar.git@main/manifests' \
  | bin/image-ref-updater testdata/events.yaml - \
  | bin/grw sink 'git@github.com:bluebrown/foobar.git@main/manifests'
```

### ConFix

The `confix` command can be used to migrate from the v1 to the v2 config format.
I tries to conver the config on a best effort basis, but careful review is
required. Some features previously supported, are not supported anymore. For
example, the scoping sematics have changed such that only a subset of previously
supported scopes can be converted in a meaningful way. Others require manual
intervention.

```bash
bin/confix -f v1.yaml -o path/to/dir/
```

The command takes an output directory, because it potentially prodcued more than
one file. For example, if there are repos with username and password in the
previous config, the command will emit an extra `.git-credentials` file, which
can be mounted to the kobold container.

## Development

You can start a local kind cluster for an end to end setp. The gitea page can be
viewed at <http://localhost:8080> with the credentials `dev:dev123`.

```bash
make testinfra
make dev
docker tag "$(docker pull busybox -q)" localhost:8080/library/busybox:v.1.2.3
docker push localhost:8080/library/busybox:v.1.2.3
```
