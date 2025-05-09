const tabContents = document.querySelectorAll('.tab-content');
const contentArea = document.querySelector('.content-area');
const modelSelector = document.getElementById('modelSelector');
const roleSelector = document.getElementById('roleSelector');
const topTabButtons = document.querySelectorAll('.top-tab-button');
var currentActiveTabId = null;

var Storage = {
    models: [],
    sessions: [],
    roles: [],
    mcps: [],
    currentSession: null
}

monitor(Storage, 'models', 'storage:models')
monitor(Storage, 'sessions', 'storage:sessions')
monitor(Storage, 'roles', 'storage:roles')
monitor(Storage, 'mcps', 'storage:mcps')
monitor(Storage, 'currentSession', 'storage:current-session')


// --- Function to handle tab switching and panel toggle ---
function handleTopTabClick(event) {
    const clickedButton = event.currentTarget;
    const targetTabId = clickedButton.getAttribute('data-tab');
    const isPanelOpen = contentArea.classList.contains('side-panel-open');
    const isCurrentActiveButton = clickedButton.classList.contains('active');

    // --- Logic: Click active button when panel is open -> Close panel ---
    if (isPanelOpen && isCurrentActiveButton) {
        contentArea.classList.remove('side-panel-open');
        clickedButton.classList.remove('active');
        currentActiveTabId = null;
    }
    // --- Logic: Click any button when panel is closed, or different button when open -> Open/Switch tab ---
    else {
        // 1. Open panel if closed
        if (!isPanelOpen) {
            contentArea.classList.add('side-panel-open');
        }

        // 2. Update button active state
        topTabButtons.forEach(btn => btn.classList.remove('active'));
        clickedButton.classList.add('active');

        // 3. Update content active state
        tabContents.forEach(content => {
            if (content.id === targetTabId) {
                content.classList.add('active');
            } else {
                content.classList.remove('active');
            }
        });
        currentActiveTabId = targetTabId;
    }
}

// --- Initial Connection on Load ---
document.addEventListener('DOMContentLoaded', async () => {
    console.log("App loaded. Attempting to connect to agent...");

    populateModelSelector()
    populateRoleSelector()
    sendEvent('sessions:reload')

    // --- Set initial active tab state based on HTML ---
    const initiallyActiveButton = document.querySelector('.top-tab-button.active');
    if (initiallyActiveButton) {
        currentActiveTabId = initiallyActiveButton.getAttribute('data-tab');
        // Ensure corresponding content is visible
        tabContents.forEach(content => {
            content.classList.toggle('active', content.id === currentActiveTabId);
        });
        if (currentActiveTabId) {
            contentArea.classList.add('side-panel-open');
        }
    } else if (topTabButtons.length > 0) {
        const activeContent = document.querySelector('.tab-content.active');
        if (activeContent) {
            currentActiveTabId = activeContent.id;
            // Find the corresponding button and mark it active
            const correspondingButton = document.querySelector(`.top-tab-button[data-tab="${currentActiveTabId}"`);
            if (correspondingButton) {
                correspondingButton.classList.add('active');
                contentArea.classList.add('side-panel-open');
            }
        } else {
            // If no button and no content is active, ensure panel is closed.
            contentArea.classList.remove('side-panel-open');
        }
    }
});

// --- Add New Top Tab Button Listener ---
topTabButtons.forEach(button => {
    button.addEventListener('click', handleTopTabClick);
});

async function populateModelSelector() {
    models = await apiListModels()
    if (models) {
        modelSelector.innerHTML = '<option value="" disabled>Select a Model</option>';
        models.sort((a, b) => a.name.localeCompare(b.name));

        let activeModelFound = false;
        models.forEach(model => {
            const option = document.createElement('option');
            option.value = model.id;
            option.textContent = model.name;
            if (!activeModelFound) {
                option.selected = true;
                activeModelFound = true;
            }
            modelSelector.appendChild(option);
        });

        // If the active model wasn't in the list or no active model was specified, select the placeholder
        if (!activeModelFound) {
            const placeholder = modelSelector.querySelector('option[disabled]');
            if (placeholder) {
                placeholder.selected = true;
            }
        }
    } else {
        modelSelector.innerHTML = '<option value="" disabled selected>Error loading models</option>';
    }
}

function getSelectedModelId() {
    return modelSelector.value;
}

function getSelectedRoleId() {
    return roleSelector.value;
}

async function populateRoleSelector() {
    const roles = await apiListRoles()
    if (roles && roles.length > 0) {
        roleSelector.innerHTML = '<option value="" disabled>Select a Role</option>';
        roles.sort((a, b) => a.name.localeCompare(b.name));

        let activeRoleFound = false;
        roles.forEach(role => {
            const option = document.createElement('option');
            option.value = role.id;
            option.textContent = role.name;
            if (!activeRoleFound) {
                option.selected = true;
                activeRoleFound = true;
            }
            roleSelector.appendChild(option);
        });

        // If the active role wasn't in the list or no active role was specified, select the placeholder
        if (!activeRoleFound) {
            const placeholder = roleSelector.querySelector('option[disabled]');
            if (placeholder) {
                placeholder.selected = true;
            }
        }
    } else {
        roleSelector.innerHTML = '<option value="" disabled selected>Default role</option>';
    }
}

function getSelectedRoleId() {
    return roleSelector.value;
}

function updateLastMessage(sessionId, message) {
    for (let session of Storage.sessions) {
        if (session.id == sessionId) {
            if (!session.messages) {
                console.error("Trying to update last message in empty message array")
                return
            }

            session.messages[session.messages.length - 1].text += message
            sendEvent('chat:last-message-update', { sessionId: sessionId })
            break
        }
    }
}

function addMessage(sessionId, message) {
}