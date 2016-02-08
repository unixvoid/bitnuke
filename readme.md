bitnuke
--------

bitnuke is a fully volitile data storage solution currently running at https://bitnuke.io  
This repo is strictly the API that supports the following REST calls:  

```
/upload  ::  takes a POST of multipart data to be stored. returns a token
/compress  ::  takes a POST of multipart form (a url) to be stored. returns a token
/{data}  ::  a token previously generated by '/upload', '/compress'. returns data
```