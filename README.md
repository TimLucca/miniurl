# MiniURL

## Usage of the live server
Send a `POST` request to `https://damp-brushlands-93308.herokuapp.com/api/new` with the following JSON payload:
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
The miniurl on this request will include the entire url (ie. `https://damp-brushlands-93308.herokuapp.com/<miniurl>`). This will redirect you to the original url and increment the number of hits on that mini.

To get statistics on your minified url, send a `POST` request to `https://damp-brushlands-93308.herokuapp.com/api/current` with the following payload:
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

## Running
GODEPS is used for the dependency management. In order to connect to MongoDB, you must add `MONGOURI`(uri to connect to the database), `DBNAME`(the name of the database to connect to), and `DBCOL`(the name of the collection you are connecting to) to your enviornment variables.

## Design
This tool is designed to take in URLs and make them smaller. The original long URL is run through a 32 bit Fowler–Noll–Vo hashing function. This hash is then encoded in **URL safe base64**. The reason for using a 32 bit hasing function rather than a larger one (ie 128/256) is that it will produce only 6 characters when base64 encoded. This makes it a reasonable length for a small URL. 

MongoDB was used as the database simply because it is easy to set up and this project deals with passing JSON around which made it a good fit. 

Collision handling was given thought, although due to the scope of the project and who would be using it there is a very small chance collison would happen (first collision would be expected around with around 2^16 entries, or 65,536 unique URLs).