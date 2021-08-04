document.querySelectorAll("pre").forEach(preElement => {
    var codeElement = preElement.querySelector("code")

    /* There should be code element inside pre element and
       code element should not have language-console class */
    if (codeElement != null && !codeElement.classList.contains("language-console")) {
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

            // Restore original "copy" text on button after 900 milliseconds
            setTimeout(() => innerText = "copy", 900)
        });

        // Insert before pre element so that it can appear in top of code block
        preElement.parentElement.insertBefore(copyButtonContainer, preElement)
    }
})

