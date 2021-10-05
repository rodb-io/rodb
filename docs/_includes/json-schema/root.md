{% assign root = include.content[include.rootKey] %}

# {{ root.title }}

{{ root.description }}

**Root config key:** `{{ include.rootKey }}`

{% include json-schema/type.md definition=root %}

{% include json-schema/examples.md examples=root.examples %}

# Map

TODO

# SQLite

TODO

# FTS5

TODO

# Wildcard

TODO

# Noop

TODO
