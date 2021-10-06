{% assign root = include.content[include.key] %}

# {{ root.title }}

{{ root.description }}

{% include json-schema/type.md definition=root key=include.key %}

{% include json-schema/examples.md examples=root.examples %}

{% for type in root.items.anyOf %}
{% assign key = type["$ref"] | remove: "./" | remove: ".yaml" %}
{% include json-schema/object.md content=include.content key=key %}
{% endfor %}
