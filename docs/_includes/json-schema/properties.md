**Properties:**

{% for propertyItem in include.properties %}
	{%- assign propertyName = propertyItem[0] -%}
	{%- assign property = propertyItem[1] -%}
	{%- assign level = include.level | plus: 1 -%}

	{%- assign required = false -%}
	{%- for requiredPropertyName in include.required -%}
		{%- if propertyName == requiredPropertyName -%}
			{%- assign required = true -%}
			{%- break -%}
		{%- endif -%}
	{%- endfor -%}

	{%- include json-schema/entity.md namespace=include.namespace definition=property key=propertyName level=level required=required -%}
{% endfor %}
