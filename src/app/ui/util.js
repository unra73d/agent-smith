const confirmOverlay = document.getElementById('confirmOverlay');
const confirmMessageEl = document.getElementById('confirmMessage');
const confirmYesBtn = document.getElementById('confirmYes');
const confirmNoBtn = document.getElementById('confirmNo');

function confirmDialog(message) {
    return new Promise((resolve) => {
        confirmMessageEl.textContent = message;
        confirmOverlay.style.display = 'flex';

        const yesListener = () => {
            cleanup();
            resolve(true);
        };

        const noListener = () => {
            cleanup();
            resolve(false);
        };

        // Function to remove listeners and hide dialog
        const cleanup = () => {
            confirmYesBtn.removeEventListener('click', yesListener);
            confirmNoBtn.removeEventListener('click', noListener);
            confirmOverlay.removeEventListener('click', backgroundClickListener); // Remove background click listener
            confirmOverlay.style.display = 'none';
        };

        // Close dialog if clicking outside the dialog box
        const backgroundClickListener = (event) => {
            if (event.target === confirmOverlay) {
                noListener(); // Treat clicking outside as "No"
            }
        };

        // Add listeners
        confirmYesBtn.addEventListener('click', yesListener);
        confirmNoBtn.addEventListener('click', noListener);
        confirmOverlay.addEventListener('click', backgroundClickListener); // Add background click listener

    });
}
function initShortcuts(){
    document.addEventListener('keydown', (event) => {
        const isModifier = event.metaKey || event.ctrlKey;

        if (!isModifier) {
            return;
        }

        let command = null;
        switch (event.key.toLowerCase()) {
            case 'c':
                command = 'copy';
                break;
            case 'v':
                command = 'paste';
                break;
            case 'x':
                command = 'cut';
                break;
            case 'a':
                command = 'selectAll';
                break;
            case 'z':
                if (event.shiftKey) {
                    command = 'redo';
                } else {
                    command = 'undo';
                }
                break
            case 'y':
                command = 'redo';
                break;
        }

        if (command) {
            try {
                if (document.execCommand(command)) {
                    event.preventDefault();
                } else {
                    console.warn(`document.execCommand('${command}') failed or did not apply.`);
                }
            } catch (e) {
                console.error(`Error executing document.execCommand('${command}'):`, e);
            }
        }
    });
}

initShortcuts();