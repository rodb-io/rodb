{% assign root = include.content[include.key] %}

<h{{ include.level }} id="{{ include.key }}">
	{{ root.title }}
</h{{ include.level }}>

{{ root.description }}

{% include json-schema/type.md definition=root key=include.key %}

{% include json-schema/examples.md examples=root.examples %}

{%- for type in root.items.anyOf -%}
	{%- assign key = type["$ref"] | remove: "./" | remove: ".yaml" -%}
	{%- assign level = include.level | plus: 1 -%}
	{%- include json-schema/object.md content=include.content key=key level=level -%}
{%- endfor -%}
