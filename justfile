build:
    nom build .#

run: build
    result/bin/vpod

run-loki endpoint: build
    result/bin/vpod | LOKI={{endpoint}} nix develop --command vector --config=vector.yaml


build-image:
    nom build .#oci-image

run-image: build-image
    docker run --rm -it $(docker load < result | hck -f 3)
