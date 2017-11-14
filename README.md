# Liftit test

This test is the solution to liftit exercise

## Project installation guide

- clone the repository

```
go get github.com/jogeraca/gotest_back
```

- Install the database

```
brew install postgres
createdb general
psql general < database.db
```

- build and install the project

```
cd ~/go/src/github.com/jogeraca/gotest_back
go get
go build
go install
```

- run the project

```
gotest_back
```

- view the frontend at [this repo](https://github.com/jogeraca/gotest_front)