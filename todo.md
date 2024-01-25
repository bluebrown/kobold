# todo

## invalid yaml

find a way to skip invalid yaml files, i.e. helm templates. this is already
possible with the help of th user, when using a .krmignore file. However, it
would be more solid, if we didnt have to rely on the user to do this.

## starlark files

store starlark scripts in file instead of db? This would make it more easy to mount
custom scripts into the container

## generic pipeline handler

option to use a handler that runs a kio pipline with multiple external filters.
This provides the greatest flexibility, at the cost of performance. Since the
resources need to be serialized and deserialized multiple times. If this handler
is selected, the decoder should return function configs instead of image refs.
The configs will be merged into a single config, which is passed to all filters.
This is still problematic as each filter uses a different format. Perhaps a
normalizer script should run before the actual filter

## starlark main func exit code

all main funcs shall return an exit code. Non zero code indicates an error. or
maybe a tuple (result, error) ?

## configurable commit/pr messages

allow to configure the commit and pr messages

## add option to set ref partially

some tools split image references across fields (e.g. image and tag). This
should be supported by kobold. There could be another option tag to specify what
part of the image ref should be set.
