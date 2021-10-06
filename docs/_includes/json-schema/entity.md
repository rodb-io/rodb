{%- if include.definition["$ref"] -%}
	{%- assign keys = include.definition["$ref"] | remove: "./" | remove: ".yaml" | split: "/" -%}

	{%- assign definition = include.namespace -%}
	{%- for key in keys -%}
		{%- assign definition = definition[key] -%}
	{%- endfor -%}
{% else %}
	{%- assign definition = include.definition -%}
{% endif %}

<h{{ include.level }} id="{{ include.key }}">
	{%- if definition.title -%}
		{{ definition.title }}
	{%- else -%}
		{{ include.key }}
	{%- endif -%}
</h{{ include.level }}>

<div class="json-schema-object" markdown="1">

{% if definition.type %}
	{%- include json-schema/type.md definition=definition key=include.key -%}
{% endif %}

{% if definition.const %}
	{%- include json-schema/const.md const=definition.const -%}
{% endif %}

{{ definition.description }}

{% if definition.examples %}
	{%- include json-schema/examples.md examples=definition.examples -%}
{% endif %}

{%- if definition.type == "array" -%}
	{%- include json-schema/items.md items=definition.items namespace=include.namespace level=include.level -%}
{%- endif -%}

{%- if definition.type == "object" -%}
	{%- include json-schema/properties.md namespace=include.namespace properties=definition.properties level=include.level -%}
{%- endif -%}

</div>
