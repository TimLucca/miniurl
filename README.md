# MiniURL

## Usage
Send a `POST` request to `<domain>/api/new` with the following JSON payload:
```
{"long": "<long url here>}
```
The server will create the minified url and return the following JSON:
```
{
  "miniurl": "<minified url here>",
  "long": "<long url here>",
  "hits": <number of times link was used>
}
```
To use the minified url, just enter the following in the browser: `<doman>/<minified value>`. This will redirect you to the original url.

To get statistics on your minified url, send a `POST` request to `<domain>/api/current` with the following payload:
```
{"miniurl": "<minified url here>"}
```
This will return the following JSON:
```
{
  "miniurl": "<minified url here>",
  "long": "<long url here>",
  "hits": <number of times link was used>
}
```

## Design
This tool is designed to take in URLs and make them smaller. The original long URL is run through a 32 bit Fowler–Noll–Vo hashing function. This hash is then encoded in **URL safe base64**. The reason for using a 32 bit hasing function rather than a larger one (ie 128/256) is that it will produce only 6 characters when base64 encoded. This makes it a reasonable length for a small URL. 