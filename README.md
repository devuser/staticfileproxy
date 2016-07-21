staticfileproxy
===============

imports `gin-gonic` and `go-bindata`

quickly static file server service for small team or web server.

-	List your static file in html which was stored in the local special folder
-	List your file which was bind in the bindata.go by go-bindata

用于快速的文件分享，把特定文件夹的满足特定要求比如扩展名的文件，展示在浏览器中。 适用于团队间的快速分享。

```
go get -u github.com/devuser/staticfileproxy
cd $GOPATH/src/github.com/devuser/staticfileproxy
cd cmd
cp config/config.default.json config/config.json
echo "modify the default config json file with your local static files folder such as ./staticfiles"
go run staticfilep-server.go
```
