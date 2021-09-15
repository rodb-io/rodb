# Japanese dictionary API

This is a more advanced example, that aims to provide a Japanese-English dictionary API. It is based on the [JMdict project](https://www.edrdg.org/jmdict/j_jmdict.html).

It takes a large [XML](/documentation/inputs/#xml) file as an input, and provides an [HTTP](/documentation/services/#http) service with a single endpoint and filtering features.

It also demonstrates the possibility of having the endpoints available via both HTTP (`http://localhost:3001`) and HTTPS (`https://localhost:3002`), using a self-signed certificate.

You can download and run this example locally (using Docker) with the following script.

<pre show-example-script="japanese-dictionary"></pre>

The source code for this example is also available [here](https://github.com/rodb-io/rodb/tree/master/examples/japanese-dictionary).

## List of words

The basic behaviour of the endpoint is to list all the words in the database:

`GET http://localhost:3001/`

Response body:

```
[
	...
	{
		"reading": "めいはく",
		"translation": "obvious",
		"writing": "明白"
	},
	{
		"reading": "あからさま",
		"translation": "plain",
		"writing": "明白"
	},
	{
		"reading": "あかん",
		"translation": "useless",
		"writing": "明かん"
	},
	{
		"reading": "あくどい",
		"translation": "gaudy",
		"writing": "悪どい"
	},
	...
]
```

## Search a specific Japanese word

The `word` filter allows filters on a Japanese word. It is based on the [SQLite](/documentation/indexes/#sqlite) index and only allows exact matching of the whole words.

`GET http://localhost:3001/?word=読む`

Response body:

```
[
	{
		"reading": "よむ",
		"translation": "to read",
		"writing": "読む"
	}
]
```

## Wildcard search in English

The `translation` filter allows to search a word by it's translation, using a [Wildcard](/documentation/indexes/#wildcard) index.

`GET http://localhost:3001/people/{id}`

Response body:

```
[
	{
		"reading": "かっさらう",
		"translation": "to snatch (and run)",
		"writing": "掻っ攫う"
	},
	{
		"reading": "サヨナラホームラン",
		"translation": "game-ending home run",
		"writing": ""
	},
	{
		"reading": "じゃりじゃり",
		"translation": "crunchy",
		"writing": ""
	},
	...
]
```

## Advanced full-text match query

The `query` filter allows more advanced searches using an [FTS5](/documentation/indexes/#fts5) index.

`GET http://localhost:3001/?query=(translation: trip AND translation: day) OR translation:work hard`

Response body:

```
[
	{
		"reading": "ひがえり",
		"translation": "day trip",
		"writing": "日帰り"
	},
	{
		"reading": "せいをだす",
		"translation": "to work hard",
		"writing": "精を出す"
	},
	{
		"reading": "べんきょうにはげむ",
		"translation": "to work hard at one's lessons",
		"writing": "勉強に励む"
	},
	...
]
```

## Paging

The results are paged using the `offset` and `limit` parameters. The default value for `limit` is `100`.

`GET http://localhost:3001/?limit=2&offset=50`

Response body:

```
[
	{
		"reading": "いかなるばあいでも",
		"translation": "in any case",
		"writing": "いかなる場合でも"
	},
	{
		"reading": "いかにも",
		"translation": "indeed",
		"writing": "如何にも"
	}
]
```

`GET http://localhost:3001/?limit=2&offset=52`

Response body:

```
[
	{
		"reading": "いくつも",
		"translation": "many",
		"writing": "幾つも"
	},
	{
		"reading": "いけない",
		"translation": "wrong",
		"writing": "行けない"
	}
]
```
