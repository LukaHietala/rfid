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

function secondsToHuman(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours} t ${minutes} min`;
}

