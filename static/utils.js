let timer

function popup(text) {
    if (timer) {
        clearTimeout(timer);
        timer = null;
    }
		
		const popupElement = document.getElementById("popup")
		popupElement.innerHTML = text

		timer = setTimeout(function() {
		        popupElement.innerHTML = ""
		}, 5000);
}

