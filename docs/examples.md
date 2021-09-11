---
layout: default
title: Examples
permalink: /examples/
---

<script type="text/javascript">
	window.addEventListener('load', function () {
		var examples = document.querySelectorAll('pre[show-example-script]');
		if (examples.length == 0) {
			return;
		}

		function showExampleScripts(assets) {
			examples.forEach(function(example) {
				var exampleName = example.getAttribute('show-example-script');
				var zipFile = assets.find(function (asset) {
					return asset.name == exampleName + '.zip';
				});

				if (zipFile == null) {
					console.error("Zip not found in the latest release assets for the example '" + exampleName + "'");
					return
				}

				example.innerHTML = [
					'wget "' + zipFile.browser_download_url + '" -O ' + exampleName + '.zip',
					'unzip ' + exampleName + '.zip',
					'rm ' + exampleName + '.zip',
					'cd ' + exampleName + ' && docker-compose up',
				].join(" && ");
			});
		}

		fetch('//api.github.com/repos/rodb-io/rodb/releases/latest')
			.then(function (response) {
				return response.json();
			})
			.then(function (data) {
				showExampleScripts(data.assets);
			})
			.catch(function (error) {
				console.error(error);
			});
	});
</script>

Examples page coming soon

{% include examples/countries.md %}
{% include examples/people.md %}
{% include examples/japanese-dictionary.md %}
{% include examples/zip-codes-tokyo.md %}
