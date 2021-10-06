{%- if include.definition.items.anyOf -%}
	{%- for type in include.definition.items.anyOf -%}
		{%- assign level = include.level | plus: 1 -%}

		{%- if type.title -%}
			{%- assign breadcrumb = include.breadcrumb | append: "[" | append: type.title | append: "]" -%}
		{%- else -%}
			{%- assign breadcrumb = include.breadcrumb | append: "[]" -%}
		{%- endif -%}

		{%- include json-schema/entity.md namespace=include.namespace definition=type level=level breadcrumb=breadcrumb -%}
	{%- endfor -%}
{%- else -%}
	{%- assign level = include.level | plus: 1 -%}

	{%- if include.definition.items.title -%}
		{%- assign breadcrumb = include.breadcrumb | append: "[" | append: include.definition.items.title | append: "]" -%}
	{%- else -%}
		{%- assign breadcrumb = include.breadcrumb | append: "[]" -%}
	{%- endif -%}

	{%- include json-schema/entity.md namespace=include.namespace definition=include.definition.items title="Items of the array" level=level breadcrumb=breadcrumb -%}
{%- endif -%}
