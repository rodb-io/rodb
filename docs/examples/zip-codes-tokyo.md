---
layout: default
title: Examples - Tokyo's zip codes API
permalink: /examples/zip-codes-tokyo
---

{% include menus.html menu=site.data.menus.examples %}

# Example - Tokyo's zip-codes API

This is another example, based of an official list of zip codes in Tokyo.

It shows some parsing capabilities, and how to handle relationships between multiple [inputs](/configuration/inputs).

It takes a [CSV](/configuration/inputs#inputs[type = &quot;csv&quot;]) file encoded as `Shift_JIS` as an input, and provides an [HTTP](/configuration/services#services[type = &quot;http&quot;]) service with two endpoints.

The encoding is handled with a custom [string](/configuration/parsers#parsers[type = &quot;boolean&quot;]) parser that is assigned to the relevant columns.

Using a [boolean](/configuration/parsers#parsers[type = &quot;boolean&quot;]) parser, the strings `0` and `1` are converted to booleans.

Other [CSV](/configuration/inputs#inputs[type = &quot;csv&quot;]) files containing some type definitions are loaded, and declared as relationships in the [output](/configuration/outputs#outputs), which means it gets inserted as a sub-object in the resulting JSON.

You can download and run this example locally (using Docker) with the following script.

{% include get-example.md exampleName="zip-codes-tokyo" %}

The source code for this example is also available [here](https://github.com/rodb-io/rodb/tree/master/examples/zip-codes-tokyo).

## List all the zip codes

This endpoint allows to list and filter the persons in the database:

`GET http://localhost:3000/zip-codes`

Response body:

{% highlight json %}
[
	{
		"code": 13101,
		"hasSubdivision": false,
		"municipality": "千代田区",
		"municipalityKana": "ﾁﾖﾀﾞｸ",
		"oldZipCode": 100,
		"prefecture": "東京都",
		"prefectureKana": "ﾄｳｷﾖｳﾄ",
		"reasonForUpdateId": 0,
		"reasonsForUpdate": {
			"id": 0,
			"name": "Unchanged"
		},
		"streetNumberAssignedPerKana": false,
		"town": "以下に掲載がない場合",
		"townHasMultipleZipCodes": false,
		"townKana": "ｲｶﾆｹｲｻｲｶﾞﾅｲﾊﾞｱｲ",
		"updated": {
			"id": 0,
			"name": "Unchanged"
		},
		"updatedId": 0,
		"zipCode": "1000000",
		"zipCodeHasMultipleTowns": false
	},
	...
]
{% endhighlight %}

## List the zip codes of a specific municipality

This endpoint provides a `municipality` parameter, that uses a [map](/configuration/indexes#indexes[type = &quot;map&quot;]) index.

`GET http://localhost:3000/zip-codes?municipality=渋谷区`

Response body:

{% highlight json %}
[
	{
		"code": 13113,
		"hasSubdivision": false,
		"municipality": "渋谷区",
		"municipalityKana": "ｼﾌﾞﾔｸ",
		"oldZipCode": 150,
		"prefecture": "東京都",
		"prefectureKana": "ﾄｳｷﾖｳﾄ",
		"reasonForUpdateId": 0,
		"reasonsForUpdate": {
			"id": 0,
			"name": "Unchanged"
		},
		"streetNumberAssignedPerKana": false,
		"town": "以下に掲載がない場合",
		"townHasMultipleZipCodes": false,
		"townKana": "ｲｶﾆｹｲｻｲｶﾞﾅｲﾊﾞｱｲ",
		"updated": {
			"id": 0,
			"name": "Unchanged"
		},
		"updatedId": 0,
		"zipCode": "1500000",
		"zipCodeHasMultipleTowns": false
	},
	...
]
{% endhighlight %}

## Paging

The results are paged using customized `offset_from` and `max_per_page` parameters. The default value for `max_per_page` is `30`.

Getting the second page with the default size:

`GET http://localhost:3000/zip-codes?offset_from=30`

Getting the third page with a size of 10 results per page

`GET http://localhost:3000/zip-codes?max_per_page=10&offset_from=20`

## Get a specific zip-code's information

This endpoint allows to get a specific record using it's zip-code.

The zip-codes are also indexed using a [map](/configuration/indexes#indexes[type = &quot;map&quot;]) index.

`GET http://localhost:3000/zip-codes/1350064`

Response body:

{% highlight json %}
{
	"code": 13108,
	"hasSubdivision": true,
	"municipality": "江東区",
	"municipalityKana": "ｺｳﾄｳｸ",
	"oldZipCode": 135,
	"prefecture": "東京都",
	"prefectureKana": "ﾄｳｷﾖｳﾄ",
	"reasonForUpdateId": 0,
	"reasonsForUpdate": {
		"id": 0,
		"name": "Unchanged"
	},
	"streetNumberAssignedPerKana": false,
	"town": "青海",
	"townHasMultipleZipCodes": false,
	"townKana": "ｱｵﾐ",
	"updated": {
		"id": 0,
		"name": "Unchanged"
	},
	"updatedId": 0,
	"zipCode": "1350064",
	"zipCodeHasMultipleTowns": false
}
{% endhighlight %}
