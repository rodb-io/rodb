{%- if include.definition.examples.size == 1 -%}
	**Example:**
{%- else -%}
	**Examples:**
{%- endif -%}

{%- for example in include.definition.examples -%}
	{%- highlight yaml -%}
		{{ example }}
	{%- endhighlight -%}
{%- endfor -%}
