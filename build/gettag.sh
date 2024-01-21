#!/usr/bin/env sh

if [ -n "$RELEASE_TAG" ]; then
  echo "$RELEASE_TAG"
else
  git describe --tags --always --dirty
fi
