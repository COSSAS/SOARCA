# Build folder

To build the project use the make file in the root.


```bash
make build 
```

## Installing packages
Install gin:

see https://github.com/gin-gonic/gin

```bash
go get -u github.com/gin-gonic/gin
```

Install swaggo:

see https://github.com/swaggo/gin-swagger

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
```