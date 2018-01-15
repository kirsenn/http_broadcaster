## Simple http broadcaster written on Go.
Creates http listener on local port, every http request will be broadcasted to endpoints from config.json file. The FIRST successful response from endpoint (200 OK) will be returned as is (headers, body)

### Usage
* Compile using go build
* Create configuration file (see example in repository). Adjust your port, environment (log level) and endpoint list
* Run program using first argument as config file `http_broadcaster config.json`
* Check it works by requesting localhost:port/endpoints
