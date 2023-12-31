## PamilyaHelper web app

### A. How to setup within the host

Requirements:

- Golang knowledge
- Golang 1.21.4 installed in the system
- App backend source code

_Note: All of the above requirements are required. Skip this part if any requirement is not met._

A. Manually build

1. Compile for the target platform e.g Linux, Windows, run: `go build`

2. Run the generated binary with: `./webapp <arg>`

B. Automatic build with `go install`

1. Install and compile the app with: `go install`, it will compile and add the binary to `GOBIN` (GOBIN should be in the path.)

2. Run the app: `webapp <arg>`. See `webapp help`

### B. How to setup inside a Docker container

Requirements:

- Docker
- App backend source code

A. Compile the app inside a linux container

1.  Build the container that will compile the app:

        docker build -f dockerfile-builder -t pamilyahelper-builder:latest -t pamilyahelper-builder:v1.0.0 .

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

3.  After compilation, the compiled app should be present under `dist/` and named _webapp_.

B. Running the compiled app inside the docker container

Requirements:

- Docker
- Compiled app

1.  Build the container that will hold the compiled app:

        docker build -t pamilyahelper:latest -t pamilyahelper:v1.0.0 .

2.  Run the app:

        docker run --rm -it \
        --volume ./storage:/pamilyahelper/storage \
        -p 127.0.0.1:5000:5000 pamilyahelper

Note: `--rm` will DELETE the container once the container exits. Remove argument if necessary.

3. View `localhost:5000` in browser.

### Load fixtures:

- docker exec -it <container-name> ./webapp loadfixtures

Default users:pass:

- `admin@admin.com:admin1234`

- `aubrey@pmh.com:aubrey1234`

- `darren@pmh.com:darren1234`

### Running tests

Dependencies:

- Node v20.9.0

1. See [test](test) folder
