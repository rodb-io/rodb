{%- if include.definition.type == "array" -%}
	{% if include.definition.items -%}
		Array of {% include json-schema/type-without-title.md definition=include.definition.items %}. 
	{%- else -%}
		Array
	{%- endif -%}

	{% if include.definition.items.anyOf -%}
		Each item in this array must match one of the following definitions.
	{%- endif -%}
{%- elsif include.definition.type == "object" -%}
	Object
{%- elsif include.definition.type == "string" -%}
	String
{%- elsif include.definition.type == "integer" -%}
	Integer
{%- elsif include.definition.type == "boolean" -%}
	Boolean
{%- elsif include.definition.type == "number" -%}
	Number
{%- elsif include.definition.type == "null" -%}
	Null
{%- else -%}
	{{ include.definition.type }}
{%- endif -%}
