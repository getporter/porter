document.querySelectorAll("pre").forEach(ele => {
    var codeElement = ele.querySelector("code")

    if (codeElement != null) {
        with (document) {
            var copyButtonContainer = createElement('div')
            copyButtonContainer.classList.add("copy-button-container")
            var copyButton = createElement('button');
            copyButtonContainer.append(copyButton)
        }

        with (copyButton) {
            className = "copy-button"
            innerText = 'copy';

            addEventListener('click', () => {                
                navigator.clipboard.writeText(codeElement.innerText).then(e => {
                    innerText = "copied"
                }).catch(e => {
                    innerText = "error"
                })
                setTimeout(() => innerText = "copy", 900)
            });
        }

        ele.parentElement.insertBefore(copyButtonContainer, ele)
    }
})

