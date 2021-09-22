---
layout: default
title: Configuration
permalink: /configuration
---

<iframe class="schema-doc"></iframe>
<script type="text/javascript">
	window.addEventListener("load", function () {
		var iframe = document.querySelector('iframe.schema-doc');
		iframe.src = "/schema/html/schema.html" + window.location.hash;

		iframe.addEventListener("load", function () {
			var iframeDocument = iframe.contentWindow.document;

			// Automatically set the iframe height to it's content
			var setHeight = function () {
				iframe.style.height = (iframeDocument.body.scrollHeight + 1) + 'px';
			};
			iframeDocument.querySelector('.generated-by-footer').style.marginBottom = 0;
			(new ResizeObserver(setHeight)).observe(iframeDocument.documentElement)
			setHeight();

			// There is no event available to handle this properly, so we are using a timeout instead...
			var lastIframeHash = window.location.hash;
			setInterval(function () {
				var currentIframeHash = iframe.contentWindow.location.hash;
				if (currentIframeHash != lastIframeHash) {
					lastIframeHash = currentIframeHash;
					window.location.hash = currentIframeHash;
				}
			}, 500);
		});
	});
</script>
