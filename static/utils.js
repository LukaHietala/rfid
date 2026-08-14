function popup(text) {
		const div = document.createElement("div");
		const newContent = document.createTextNode(text);
		div.style.position = "fixed";
		div.style.right = "50px";
		div.style.top = "50px";
		div.append(newContent);
		document.body.appendChild(div);

		setTimeout(function() {
				document.body.removeChild(div);
		}, 5000);
}

