build:
    nom build .#

test:
    go test ./... -v

run: build
    result/bin/vpod

run-loki loki vlogs: build
    result/bin/vpod | LOKI={{loki}} VLOGS={{vlogs}} nix develop --command vector --config=vector.yaml

build-image:
    nom build .#oci-image

run-image: build-image
    docker run --rm -it $(docker load < result | hck -f 3)
