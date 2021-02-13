# signaller

Monitor a list of files and send a signal to a pid when an event takes place.


docker run --rm -it \
  -v $PWD:/go/src/github.com/johnbuhay/signaller \
  -w /go/src/github.com/johnbuhay/signaller \
  -e GITHUB_TOKEN \
  --entrypoint /bin/bash \
  goreleaser/goreleaser