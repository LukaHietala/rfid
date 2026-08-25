let timer;

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

function formatForIndex(seconds) {
		return (seconds < 0) ? `Ylitöitä tehty: ${secondsToHuman(seconds)}`
				: `Töitä jäljellä: ${secondsToHuman(seconds)}`;
}

function secondsToHuman(seconds) {
		const hours = Math.floor(Math.abs(seconds) / 3600);
		const minutes = Math.floor((Math.abs(seconds) % 3600) / 60);
		return `${hours} t ${minutes} min`;
}

