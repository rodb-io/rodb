---
layout: default
title: Documentation - Setup
permalink: /documentation/setup
---

{% include menus/documentation.html %}

TODO Setup documentation and figure out how to integrate the json schema

<iframe class="schema-doc" src="/schema/schema.html"></iframe>
<script type="text/javascript">
	window.addEventListener("load", function () {
		var iframe = document.querySelector('iframe.schema-doc');
		var setHeight = function() {
			iframe.style.height = (iframe.contentWindow.document.body.scrollHeight + 1) + 'px';
		};

		iframe.contentWindow.document.querySelector('.generated-by-footer').style.marginBottom = 0;
		(new ResizeObserver(setHeight)).observe(iframe.contentWindow.document.documentElement)

		setHeight();
	});
</script>
