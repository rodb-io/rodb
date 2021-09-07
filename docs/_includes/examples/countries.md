# Countries API

This is the most basic example of how to use RODB.

It takes a single JSON file as an input, and provides an HTTP endpoint which return a list of countries.

You can download and run this example locally (using Docker) with the following script.

<pre show-example-script="countries"></pre>

The API provides a single endpoint with filtering:
- `GET http://localhost:3004/` lists all the countries
- `GET http://localhost:3004/?continentCode=EU`: lists all the countries in Europe
- `GET http://localhost:3004/?countryCode=US`: returns a list containing only the USA

Since this example does not set-up any indexes, all of the parameters must be an exact match.
It also means that RODB iterates all the records to find the requested ones, which is not recommended with a large data-set.
