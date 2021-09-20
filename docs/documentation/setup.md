---
layout: default
title: Documentation - Setup
permalink: /documentation/setup
---

{% include menus/documentation.html %}

TODO Setup documentation and figure out how to integrate the json schema

<iframe class="schema-doc" src="/schema/html/schema.html"></iframe>
<script type="text/javascript">
	window.addEventListener("load", function () {
		var iframe = document.querySelector('iframe.schema-doc');
		var iframeDocument = iframe.contentWindow.document;

		// Automatically set the iframe height to it's content
		var setHeight = function() {
			iframe.style.height = (iframeDocument.body.scrollHeight + 1) + 'px';
		};
		iframeDocument.querySelector('.generated-by-footer').style.marginBottom = 0;
		(new ResizeObserver(setHeight)).observe(iframeDocument.documentElement)
		setHeight();
	});
</script>
