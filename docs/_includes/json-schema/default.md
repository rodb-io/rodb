
{% if include.definition.type == "string" and include.definition.default -%}
	**Default value:**  `"{{ include.definition.default }}"`
{% elsif include.definition.default -%}
	**Default value:**  `{{ include.definition.default }}`
{% endif %}
