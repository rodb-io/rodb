{% if include.examples.size == 1 %}
**Example:**
{% else %}
**Examples:**
{% endif %}

{% for example in include.examples %}
{% highlight yaml %}
{{ example }}
{% endhighlight %}
{% endfor %}
