run:
    rm ./*.db && go run .

build:
    nom build .#

build-image:
    nom build .#oci-image

run-image: build-image
    docker run --rm -it $(docker load < result | hck -f 3)
