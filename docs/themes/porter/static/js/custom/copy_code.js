document.querySelectorAll("pre").forEach(ele => {
    var codeElement = ele.querySelector("code")

    if (codeElement != null) {
        var copyButtonContainer = document.createElement('div')
        copyButtonContainer.classList.add("copy-button-container")
        var copyButton = document.createElement('button');
        copyButtonContainer.append(copyButton)

        copyButton.className = "copy-button"
        copyButton.innerText = 'copy';

        copyButton.addEventListener('click', () => {
            navigator.clipboard.writeText(codeElement.innerText).then(e => {
                copyButton.innerText = "copied"
            }).catch(e => {
                copyButton.innerText = "error"
            })
            setTimeout(() => innerText = "copy", 900)
        });

        ele.parentElement.insertBefore(copyButtonContainer, ele)
    }
})

