# People API

This is a more advanced example, based of a randomly generated list of one million fictious persons.

It takes a file containing multiple [JSON](/documentation/inputs/#json) documents as an input, and provides an [HTTP](/documentation/services/#http) service with two endpoints.

You can download and run this example locally (using Docker) with the following script.

<pre show-example-script="people"></pre>

The source code for this example is also available [here](https://github.com/rodb-io/rodb/tree/master/examples/people).

## List everyone in the database

This endpoint allows to list and filter the persons in the database:

`GET http://localhost:3003/people`

Response body:

```
[
	{
		"email": "ZbqFNla@XHTbHUG.net",
		"firstName": "Rudy",
		"gender": "Female",
		"id": 1,
		"lastName": "Gleason",
		"phoneNumber": "108-953-1642",
		"username": "GcSyDdb"
	},
	{
		"email": "cGNHYWP@GwUXDTY.net",
		"firstName": "Gustave",
		"gender": "Prefer to skip",
		"id": 2,
		"lastName": "Smith",
		"phoneNumber": "437-109-6851",
		"username": "Jcmkdmh"
	},
	...
]
```

## Search in the database

The `search` filter allows advanced searching and filtering a person by either it's `firstName`, `lastName` or `userName`. This is done by indexing those properties using the [FTS5](/documentation/indexes/#fts5) index.

`GET http://localhost:3003/people?search=firstName:John AND lastName:An*`

Response body:

```
[
	{
		"email": "upNHUVg@BtFHXhT.info",
		"firstName": "John",
		"gender": "Prefer to skip",
		"id": 248375,
		"lastName": "Anderson",
		"phoneNumber": "863-710-2941",
		"username": "fyIipNs"
	},
	{
		"email": "WpDQWXb@GWlwRhr.org",
		"firstName": "John",
		"gender": "Male",
		"id": 257072,
		"lastName": "Ankunding",
		"phoneNumber": "785-104-6391",
		"username": "igmHRYM"
	},
	...
]
```

## Paging

The results are paged using the `offset` and `limit` parameters. The default value for `limit` is `100`.

`GET http://localhost:3003/people?limit=2&offset=0`

Response body:

```
[
	{
		"email": "ZbqFNla@XHTbHUG.net",
		"firstName": "Rudy",
		"gender": "Female",
		"id": 1,
		"lastName": "Gleason",
		"phoneNumber": "108-953-1642",
		"username": "GcSyDdb"
	},
	{
		"email": "cGNHYWP@GwUXDTY.net",
		"firstName": "Gustave",
		"gender": "Prefer to skip",
		"id": 2,
		"lastName": "Smith",
		"phoneNumber": "437-109-6851",
		"username": "Jcmkdmh"
	}
]
```

`GET http://localhost:3003/people?limit=2&offset=2`

Response body:

```
[
	{
		"email": "waCBExK@bDdSKsu.ru",
		"firstName": "Aiyana",
		"gender": "Male",
		"id": 3,
		"lastName": "Schuppe",
		"phoneNumber": "945-108-3127",
		"username": "dnVShFW"
	},
	{
		"email": "bvwtuQh@NrPOIli.ru",
		"firstName": "Bertrand",
		"gender": "Prefer to skip",
		"id": 4,
		"lastName": "Hermann",
		"phoneNumber": "910-312-5874",
		"username": "wZNoEpK"
	}
]
```

## Get a specific person by id

This endpoint allows to get a specific record using it's id property.
The ids are indexed in an [SQLite](/documentation/indexes/#sqlite) database.

`GET http://localhost:3003/people/{id}`

Response body:

```
{
	"email": "nOUblmG@ruJJjDe.com",
	"firstName": "Brandon",
	"gender": "Male",
	"id": 42,
	"lastName": "Runte",
	"phoneNumber": "268-194-7351",
	"username": "grVSJDH"
}
```
