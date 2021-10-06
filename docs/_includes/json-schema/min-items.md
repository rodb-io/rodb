{%- if include.definition.minItems == 1 -%}
	Empty array not allowed.
{%- elsif include.definition.minItems > 1 -%}
	Must have at least *{{ include.definition.minItems }}* items.
{%- endif -%}
