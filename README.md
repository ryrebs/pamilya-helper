## PamilyaHelper web app.

Dependencies:

1. Node v20.9.0

2. Golang 1.21.4

3. Docker

### A. Getting Started

1. Build the app: `go build`

2. Install the app with: `go install`

3. Run the app: `webapp <arg>`. See `webapp help`

4. Or without installation, run the built binary with: `./webapp <arg>`

### Running the app in Docker.

A. Build the image that compiles the app.

1.  Build the container:

        docker build -t pamilyahelper-builder:latest -t pamilyahelper-builder:v0.0.1

2.  Run the container to compile the app:

        docker run \
                --volume=<absolute-path-to>/dist:/pamilyahelper/dist \
                --volume=<absolute-path-to>/main.go:/pamilyahelper/main.go \
                --volume=<absolute-path-to>/server:/pamilyahelper/server \
                --volume=<absolute-path-to>/go.mod:/pamilyahelper/go.mod \
                --volume=<absolute-path-to>/build.sh:/pamilyahelper/build.sh \
                --rm \
                -it \
                pamilyahelper-builder

B. Running the compiled app.

1. Build the container: `docker build -t pamilyahelper:latest -t pamilyahelper:v0.0.1 .` (Update build version accordingly)

2. Run the app: `docker run --rm -it -p 127.0.0.1:5000:5000 pamilyahelper`

3. View `localhost:5000` in browser.

Optionally, attached a database named: `pamilyahelper.db`

        docker run --rm \
                --volume /run/media/amniuz/data/Rps/temp/pamilya-helper/pamilyahelper.db:/pamilyahelper/pamilyahelper.db \
                -it -p 127.0.0.1:5000:5000 pamilyahelper

### Fixtures loaded:

Default users:pass:

- `admin@admin.com:admin1234`

- `aubrey@pmh.com:aubrey1234`

- `darren@pmh.com:darren1234`

### Running tests

- See [test](test) folder
