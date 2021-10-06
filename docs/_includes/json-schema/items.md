{%- if include.definition.items.anyOf -%}
	{%- for type in include.definition.items.anyOf -%}
		{%- assign level = include.level | plus: 1 -%}
		{%- include json-schema/entity.md namespace=include.namespace definition=type level=level -%}
	{%- endfor -%}
{%- else -%}
	{%- assign level = include.level | plus: 1 -%}
	{%- include json-schema/entity.md namespace=include.namespace definition=include.definition.items key="Items of the array" level=level -%}
{%- endif -%}
