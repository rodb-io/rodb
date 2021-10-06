**Properties:**

{% for propertyItem in include.definition.properties %}
	{%- assign propertyName = propertyItem[0] -%}
	{%- assign property = propertyItem[1] -%}
	{%- assign level = include.level | plus: 1 -%}
	{%- assign breadcrumb = include.breadcrumb | append: "." | append: propertyName -%}

	{%- assign required = false -%}
	{%- for requiredPropertyName in include.definition.required -%}
		{%- if propertyName == requiredPropertyName -%}
			{%- assign required = true -%}
			{%- break -%}
		{%- endif -%}
	{%- endfor -%}

	{%- include json-schema/entity.md namespace=include.namespace definition=property key=propertyName level=level required=required breadcrumb=breadcrumb -%}
{% endfor %}
