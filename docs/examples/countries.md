---
layout: default
title: Examples - Countries API
permalink: /examples/countries
---

{% include menus/examples.html %}

# Countries API

This is the most basic example of how to use RODB.

It takes a single [CSV]({{ site.link.input.csv }}) file as an input, and provides an [HTTP]({{ site.link.service.http }}) service with one endpoint which returns a list of countries.

You can download and run this example locally (using Docker) with the following script.

{% include scripts/get-example.md exampleName="countries" %}

The source code for this example is also available [here](https://github.com/rodb-io/rodb/tree/master/examples/countries).

The API provides a single endpoint with basic filtering:
- `GET http://localhost:3004/`: lists all the countries
- `GET http://localhost:3004/?continentCode=EU`: lists all the countries in Europe
- `GET http://localhost:3004/?countryCode=US`: returns a list containing only the USA

Here is a sample response body of this endpoint:

{% highlight json %}
[
	{
		"continentCode": "EU",
		"continentName": "Europe",
		"countryCode": "AL",
		"countryName": "Albania, Republic of"
	},
	{
		"continentCode": "EU",
		"continentName": "Europe",
		"countryCode": "AD",
		"countryName": "Andorra, Principality of"
	},
	...
]
{% endhighlight %}

Since this example does not set-up any indexes, all of the parameters must be an exact match.
It also means that RODB iterates all the records to find the requested ones, which is not recommended with a large data-set.
