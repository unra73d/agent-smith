'use strict'
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
deepMonitor(Storage, 'mcps', 'storage:mcps')
monitor(Storage, 'currentSession', 'storage:current-session')

document.addEventListener('sessions:reload', async (e) => {
    let sessions = await apiListSessions()
    if (sessions && sessions.length > 0) {
        sessions.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
        Storage.sessions = sessions
        Storage.currentSession = Storage.sessions[0]
    } else {
        sendEvent('sessions:new')
    }
})

document.addEventListener('mcp:reloaded', async (e) => {
    let mcps = e.detail
    if (mcps && mcps.length > 0) {
        for (let i in mcps) {
            mcps[i].active = true
            if (!mcps[i].tools) mcps[i].tools = []
            for (let k in mcps[i].tools) {
                mcps[i].tools[k].active = true
            }
        }
        Storage.mcps = mcps
    }
})

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
    await apiAgentConnect()

    populateModelSelector()
    populateRoleSelector()
    sendEvent('sessions:reload')

    document.querySelectorAll('mcp-list').forEach(list => list.items.data = Storage.mcps)
    apiListMCPServers().then(res => sendEvent('mcp:reloaded', res))

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
    const models = await apiListModels();
    if (models) {
        const options = [
            { value: '', label: 'Select a Model', disabled: true, selected: true },
            ...models.sort((a, b) => a.name.localeCompare(b.name)).map((model, idx) => ({
                value: model.id,
                label: model.name,
                selected: idx === 0
            }))
        ];
        modelSelector.options = options;
    } else {
        modelSelector.options = [
            { value: '', label: 'Error loading models', disabled: true, selected: true }
        ];
    }
}

function getSelectedModelId() {
    return modelSelector.value;
}

async function populateRoleSelector() {
    const roles = await apiListRoles();
    if (roles && roles.length > 0) {
        const options = [
            { value: '', label: 'Select a Role', disabled: true, selected: true },
            ...roles.sort((a, b) => a.name.localeCompare(b.name)).map((role, idx) => ({
                value: role.id,
                label: role.name,
                selected: idx === 0
            }))
        ];
        roleSelector.options = options;
    } else {
        roleSelector.options = [
            { value: '', label: 'Default role', disabled: true, selected: true }
        ];
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

            session.messages[session.messages.length - 1].text = message.text
            session.messages[session.messages.length - 1].toolRequests = message.toolRequests
            sendEvent('chat:last-message-update', { sessionId: sessionId })
            break
        }
    }
}

function addMessage(sessionId, message) {
}