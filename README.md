# MediaType
MediaType is a fully [RFC 7231](https://tools.ietf.org/html/rfc7231#section-5.3) compliant helper library for doing content-type negotiation.

## Usage
The core of the library is a single `ContentType` object that represents a single full definition for a transfer encoding type like `application/json`. This can range from the extremely simple types like the wildcard `*/*` all the way up to very complex negotiation types like `text/html; charset=utf-8; q=0.1`.

`ContentType`'s can be created directly from strings with `ParseSingle()`, or lists of them with `Parse()`. The list is more common since content negotiation may include more than one option.

When performing negotiation, if your service only supports one type, you can check if it's supported by the client with `ContentList.SupportsType()`, which will check if your content type is supported at all by the client.

If you support multiple return types, you can choose the best type based on the clients request by the `ContentTypeList.PreferredMatch()` against a list of types your server supports. This will return a single `ContentType` from the options that should be used.

## Example
```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {


    // Define what we support returning responses in
    //
    // in actual code you'd do this once at startup and not in the
    // response handler each time
    ExpectedRequestType, _ := mediatype.ParseSingle("application/json")
    SupportedResponseTypes, _ := mediatype.Parse("application/json, text/html")

    // Get the client supported types
    reqType, resOptions, err := mediatype.ParseRequest(r)
    if err != nil {
        http.Error(...)
        return
    }

    // check the request contains what we're expecting
    if !reqType.Matches(ExpectedRequestType) {
        http.Error(...)
        return
    }

    // Determine if we can respond
    resType := SupportedResponseTypes.PreferredMatch(resOptions)
    if resType == nil {
        http.Error(...)
        return
    }

    // Do something
    responseData := getData(r)

    // Render response
    if resType.SubType == "json" {
        writeJson(w, responseData)
        return
    }

    if resType.SubType == "html" {
        writeHtml(w, responseData)
        return
    }

}
```