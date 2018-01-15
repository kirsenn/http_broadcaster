package main

import (
    "net/http"
    "go.uber.org/zap"
    "strings"

    "time"
    "io/ioutil"
)

const connectTimeOut = time.Second * 5
const checkPath = "endpoints"

type server struct {
    Config
    Logger *zap.SugaredLogger
}

func (server *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path[1:] == checkPath {
        server.writeEndpoints(w)
        return
    }

    client := &http.Client{
        Timeout: connectTimeOut,
    }
    var channelBufferSize = len(server.Config.Endpoints)
    responseChan := make(chan *http.Response, channelBufferSize)

    for _, endpoint := range server.Config.Endpoints {
        go server.requestEndpoint(r, endpoint, client, responseChan)
    }

    var lastResponse bool
    for i := 0; i < channelBufferSize; i++ {
        lastResponse = channelBufferSize == i+1
        response := <-responseChan

        if response == nil {
            if lastResponse {
                server.Logger.Errorw("No one endpoint responded at all", "requestedUrl", r.URL)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            continue
        }

        if response.StatusCode == http.StatusOK || lastResponse {
            for headerKey, headerVal := range response.Header {
                w.Header().Add(headerKey, strings.Join(headerVal, ","))
            }

            w.WriteHeader(response.StatusCode)

            body, _ := ioutil.ReadAll(response.Body)
            w.Write([]byte(body))

            if lastResponse && response.StatusCode != http.StatusOK {
                server.Logger.Errorw("No one endpoint responded 200", "requestedUrl", r.URL)
            }

            response.Body.Close()
            return
        }
    }
}

func (server *server) writeEndpoints(w http.ResponseWriter) {
    for _, endpoint := range server.Config.Endpoints {
        w.Write([]byte("<html>"))
        w.Write([]byte(endpoint + "<br>"))
        w.Write([]byte("</html>"))
    }
}

func (server *server) requestEndpoint(r *http.Request, endpoint string, client *http.Client, responseChan chan *http.Response) {
    var uri = endpoint

    if len(r.URL.Path) > 1 {
        uri += r.URL.Path
    }

    if r.URL.RawQuery != "" {
        uri += "?" + r.URL.RawQuery
    }

    request, _ := http.NewRequest(r.Method, uri, r.Body)
    request.Header = r.Header

    response, err := client.Do(request)

    if response != nil {
        server.Logger.Debugw("Endpoint answered",
            "request", r.URL,
            "endpointUrl", uri,
            "responseCode", response.StatusCode,
        )
    } else {
        server.Logger.Debugw("Endpoint didn't answer",
            "request", r.URL,
            "endpointUrl", uri,
            "error", err.Error(),
        )
    }

    responseChan <- response
}
