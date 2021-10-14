{%- if include.definition["$ref"] -%}
	{%- assign keys = include.definition["$ref"] | remove: "./" | remove: ".yaml" | split: "/" -%}

	{%- assign definition = include.namespace -%}
	{%- for key in keys -%}
		{%- assign definition = definition[key] -%}
	{%- endfor -%}
{% else %}
	{%- assign definition = include.definition -%}
{% endif %}

<h{{ include.level }} id="{{ include.breadcrumb | escape }}">
	<a href="#{{ include.breadcrumb | escape }}">
		{%- if definition.title -%}
			{{ definition.title }}
		{%- else -%}
			{{ include.title }}
		{%- endif -%}

		{% if include.required == false %}
			(optional)
		{% endif %}
	</a>
</h{{ include.level }}>

<div class="json-schema-object" markdown="1">

<span class="breadcrumb">{{ include.breadcrumb }}</span>

{% if definition.type %}
	{%- include json-schema/type.md definition=definition -%}
{% endif %}

{%- if definition.minItems -%}
	{%- include json-schema/min-items.md definition=definition -%}
{%- endif -%}

{% if definition.default %}
	{%- include json-schema/default.md definition=definition -%}
{% endif %}

{% if definition.const %}
	{%- include json-schema/const.md definition=definition -%}
{% endif %}

{{ definition.description }}

{% if definition.examples %}
	{%- include json-schema/examples.md definition=definition -%}
{% endif %}

{%- if definition.type == "array" -%}
	{%- include json-schema/items.md namespace=include.namespace definition=definition level=include.level breadcrumb=include.breadcrumb -%}
{%- endif -%}

{%- if definition.type == "object" -%}
	{%- include json-schema/properties.md namespace=include.namespace definition=definition level=include.level breadcrumb=include.breadcrumb -%}
{%- endif -%}

</div>
