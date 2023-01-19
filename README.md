# Sample code for msgraph

Just sample go code for using msgraph without https://github.com/microsoftgraph/msgraph-sdk-go

## Why ?

https://github.com/microsoftgraph/msgraph-sdk-go/issues/129

## Is there an alternative ?

If you are not using many graph API calls
you can just call it directly using standard go libraries

Most of the stuff can be figured out looking at
https://developer.microsoft.com/en-us/graph/graph-explorer


The needed go structures can be generated from the json the API returns using any json to go converter.

## Sample execution

```
$ az login --use-device-code
To sign in, use a web browser to open the page https://microsoft.com/devicelogin ...

$ DISPLAY=appname go run main.go
2023/01/19 21:34:07 app exists, app id: 0ee180f0-c476-4ff0-85e2-1ddeca88a78b
```

