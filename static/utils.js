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

class StudentBox extends HTMLElement {
		constructor() {
				super();
				const shadowRoot = this.attachShadow({mode: 'open'});

				
				
				shadowRoot.appendChild(cont);
		}
}













































