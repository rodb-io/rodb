{% assign exampleFileName = include.exampleName | append: '.zip' %}
{% assign example = site.github.latest_release.assets | where:"name",exampleFileName | first %}

{% capture command %}
wget "{{ example.browser_download_url }}" -O {{ exampleFileName }}
 && unzip {{ exampleFileName }}
 && rm {{ exampleFileName }}
 && cd {{ include.exampleName }}
 && docker-compose up
{% endcapture %}

{% highlight bash %}
{{ command | strip_newlines }}
{% endhighlight %}
