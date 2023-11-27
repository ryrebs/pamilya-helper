## PamilyaHelper web app.

### Getting Started

1. Build the app: `CGO_ENABLED=1 GOOS=linux go build`

2. Install the app with: `go install`

3. Run the app: `webapp <arg>`. See `webapp help`

4. Or without installation, run the built binary with: `./webapp <arg>`

### Build app in Docker container

1. Build the container:
        
        docker build -t pamilyahelper-builder:latest -t pamilyahelper-builder:v0.0.1

2. Run the container to build the app:

        docker run \
                --volume=<absolute-path-to>/dist:/pamilyahelper/dist \
                --volume=<absolute-path-to>/main.go:/pamilyahelper/main.go \
                --volume=<absolute-path-to>/server:/pamilyahelper/server \
                --volume=<absolute-path-to>/go.mod:/pamilyahelper/go.mod \
                --volume=<absolute-path-to>/build.sh:/pamilyahelper/build.sh \
                --rm \
                -it \
                pamilyahelper-builder


### Run the app in Docker container

1. Build the container: `docker build -t pamilyahelper:latest -t pamilyahelper:v0.0.1 .`

2. Run the app: `docker run --rm -it -p 127.0.0.1:5000:5000 pamilyahelper`

Optionally, attached a database named: `pamilyahelper.db`


        docker run --rm \
                --volume /run/media/amniuz/data/Rps/temp/pamilya-helper/pamilyahelper.db:/pamilyahelper/pamilyahelper.db \
                -it -p 127.0.0.1:5000:5000 pamilyahelper


### Fixtures loaded:

- Default admin:pass: `admin@admin.com:admin1234`