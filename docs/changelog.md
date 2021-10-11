---
layout: default
title: Changelog
permalink: /changelog
---

# Changelog

{% for release in site.github.releases %}

{%- assign date = release.created_at | split: " " | first -%}
## Release notes for version {{ release.name }} <time datetime="{{ date }}">({{ date }})</time>

{{ release.body }}

{% endfor %}
