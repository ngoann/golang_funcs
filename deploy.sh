echo ">>> Build to exe application"
env GOOS=windows GOARCH=amd64 go build SeoTool.go
echo ">>> Build to linux application"
go build SeoTool.go
echo ">>> Ziping file"
zip -r deploy/seo-tool SeoTool SeoTool.exe index.html images
echo ">>> Finished"
