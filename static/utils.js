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

// ALLOW chrome://flags/  or about:config autoplay!
function beep() {
    const audioCtx = new(window.AudioContext || window.webkitAudioContext)();

    const oscillator = audioCtx.createOscillator();
    const gainNode = audioCtx.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(audioCtx.destination);

    gainNode.gain.value = 1;
    oscillator.frequency.value = 1200;
    oscillator.type = "sine";

    oscillator.start();
    oscillator.stop(audioCtx.currentTime + 0.2);
}
